// Copyright © 2017-2018 Stratumn SAS
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as
// published by the Free Software Foundation, either version 3 of the
// License, or (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <https://www.gnu.org/licenses/>.

package protocol

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"io"

	pb "github.com/stratumn/go-node/app/coin/pb"
	"github.com/stratumn/go-node/app/coin/protocol/chain"
	"github.com/stratumn/go-node/app/coin/protocol/engine"
	"github.com/stratumn/go-node/app/coin/protocol/gossip"
	"github.com/stratumn/go-node/app/coin/protocol/miner"
	"github.com/stratumn/go-node/app/coin/protocol/p2p"
	"github.com/stratumn/go-node/app/coin/protocol/processor"
	"github.com/stratumn/go-node/app/coin/protocol/state"
	"github.com/stratumn/go-node/app/coin/protocol/synchronizer"
	"github.com/stratumn/go-node/app/coin/protocol/validator"

	logging "github.com/ipfs/go-log"
	inet "github.com/libp2p/go-libp2p-net"
	protocol "github.com/libp2p/go-libp2p-protocol"
	protobuf "github.com/multiformats/go-multicodec/protobuf"
)

// ProtocolID is the protocol ID of the protocol.
var ProtocolID = protocol.ID("/stratumn/node/coin/v1.0.0")

// log is the logger for the service.
var log = logging.Logger("coin")

// Protocol describes the interface exposed to other nodes in the network.
type Protocol interface {
	// AddTransaction validates a transaction and adds it to the transaction pool.
	AddTransaction(tx *pb.Transaction) error

	// AppendBlock validates the incoming block and adds it at the end of
	// the chain, updating internal state to reflect the block's transactions.
	// Note: it might cause the consensus engine to stop mining on an outdated
	// block and mine on top of the newly added block.
	AppendBlock(block *pb.Block) error

	// StartMining starts mining blocks.
	// This method will not return until the input context is canceled.
	StartMining(ctx context.Context) error
}

// Coin implements Protocol with a PoW engine.
type Coin struct {
	genesisBlock *pb.Block
	engine       engine.Engine
	state        state.State
	chain        chain.Chain
	gossip       gossip.Gossip
	txpool       state.TxPool
	processor    processor.Processor
	validator    validator.Validator
	synchronizer synchronizer.Synchronizer
	p2p          p2p.P2P

	miner *miner.Miner
}

// NewCoin creates a new Coin.
func NewCoin(
	b *pb.Block,
	txp state.TxPool,
	e engine.Engine,
	s state.State,
	c chain.Chain,
	g gossip.Gossip,
	v validator.Validator,
	p processor.Processor,
	p2p p2p.P2P,
	sync synchronizer.Synchronizer,
) *Coin {

	miner := miner.NewMiner(txp, e, s, c, v, p, g)

	return &Coin{
		genesisBlock: b,
		engine:       e,
		state:        s,
		chain:        c,
		txpool:       txp,
		processor:    p,
		validator:    v,
		gossip:       g,
		p2p:          p2p,
		synchronizer: sync,
		miner:        miner,
	}
}

// Run starts the coin.
func (c *Coin) Run(ctx context.Context) error {
	if err := c.processGenesisBlock(ctx); err != nil {
		return err
	}

	if err := c.StartTxGossip(ctx); err != nil {
		return err
	}

	return c.StartBlockGossip(ctx)
}

// processGenesisBlock checks that the genesis block is in the chain and adds it if not.
func (c *Coin) processGenesisBlock(ctx context.Context) error {
	e := log.EventBegin(ctx, "ProcessGenesisBlock")
	defer e.Done()

	_, err := c.chain.CurrentBlock()
	if err == chain.ErrBlockNotFound {
		return c.processor.Process(ctx, c.genesisBlock, c.state, c.chain)
	}
	if err != nil {
		e.SetError(err)
		return err
	}

	e.SetError(errors.New("chain already initialized"))
	return nil
}

// StreamHandler handles incoming messages from peers.
func (c *Coin) StreamHandler(ctx context.Context, stream inet.Stream) {
	event := log.EventBegin(ctx, "handleStream", logging.Metadata{
		"stream": stream,
	})
	defer func() {
		err := inet.FullClose(stream)
		if err != nil {
			event.Append(logging.Metadata{"close_err": err.Error()})
		}

		event.Done()
	}()

	dec := protobuf.Multicodec(nil).Decoder(stream)
	enc := protobuf.Multicodec(nil).Encoder(stream)

	for {
		if err := ctx.Err(); err != nil {
			break
		}

		var gossip pb.Request
		err := dec.Decode(&gossip)
		if err == io.EOF {
			return
		}
		if err != nil {
			event.Append(logging.Metadata{"decode_err": err.Error()})
			continue
		}

		switch m := gossip.Msg.(type) {
		case *pb.Request_HeaderReq:
			if err := c.p2p.RespondHeaderByHash(ctx, m.HeaderReq, enc, c.chain); err != nil {
				event.Append(logging.Metadata{"header_req_err": err.Error()})
			}
		case *pb.Request_HeadersReq:
			if err := c.p2p.RespondHeadersByNumber(ctx, m.HeadersReq, enc, c.chain); err != nil {
				event.Append(logging.Metadata{"headers_req_err": err.Error()})
			}
		case *pb.Request_BlockReq:
			if err := c.p2p.RespondBlockByHash(ctx, m.BlockReq, enc, c.chain); err != nil {
				event.Append(logging.Metadata{"block_req_err": err.Error()})
			}
		case *pb.Request_BlocksReq:
			if err := c.p2p.RespondBlocksByNumber(ctx, m.BlocksReq, enc, c.chain); err != nil {
				event.Append(logging.Metadata{"blocks_req_err": err.Error()})
			}
		default:
			event.Append(logging.Metadata{
				"err":  "Unexpected type",
				"type": fmt.Sprintf("%T", m),
			})
		}
	}
}

// GetAccount gets the account details of a user identified
// by his public key. It returns &pb.Account{} if the account is not
// found.
func (c *Coin) GetAccount(peerID []byte) (*pb.Account, error) {
	return c.state.GetAccount(peerID)
}

// GetAccountTransactions gets the transaction history of a user identified
// by his public key.
func (c *Coin) GetAccountTransactions(peerID []byte) ([]*pb.Transaction, error) {
	txKeys, err := c.state.GetAccountTxKeys(peerID)
	if err != nil {
		return nil, err
	}
	transactions := make([]*pb.Transaction, len(txKeys))
	for i, txKey := range txKeys {
		blk, err := c.chain.GetBlockByHash(txKey.BlkHash)
		if err != nil {
			return nil, err
		}
		transactions[i] = blk.GetTransactions()[txKey.TxIdx]
	}
	return transactions, nil
}

// GetTransactionPool returns the size of the transaction pool
// and a few random transactions from the pool.
func (c *Coin) GetTransactionPool(count uint32) (uint64, []*pb.Transaction, error) {
	if count == 0 {
		count = 1
	}

	txCount := c.txpool.Pending()
	if txCount == 0 {
		return 0, nil, errors.New("transaction pool is empty")
	}

	txs := c.txpool.Peek(count)

	return txCount, txs, nil
}

// GetBlockchain gets blocks from the blockchain.
// It returns the blocks in decreasing block number,
// starting from the block requested.
func (c *Coin) GetBlockchain(blockNumber uint64, hash []byte, count uint32) ([]*pb.Block, error) {
	var start *pb.Block
	var err error
	if blockNumber != 0 {
		start, err = c.chain.GetBlockByNumber(blockNumber)
	} else if hash != nil {
		start, err = c.chain.GetBlockByHash(hash)
	} else {
		start, err = c.chain.CurrentBlock()
	}

	if err != nil {
		return nil, err
	}

	if count < 2 {
		return []*pb.Block{start}, nil
	}

	blocks := make([]*pb.Block, count)
	blocks[0] = start
	for i := uint32(1); i < count; i++ {
		blocks[i], err = c.chain.GetParentBlock(blocks[i-1].Header)
		if err != nil {
			return nil, err
		}

		if blocks[i].BlockNumber() == 0 {
			blocks = blocks[:i+1]
			break
		}
	}

	return blocks, nil
}

// PublishTransaction publishes and adds transaction received via grpc.
func (c *Coin) PublishTransaction(tx *pb.Transaction) error {
	return c.gossip.PublishTx(tx)
}

// AddTransaction validates incoming transactions against the latest state
// and adds them to the pool.
func (c *Coin) AddTransaction(tx *pb.Transaction) error {
	err := c.validator.ValidateTx(tx, c.state)
	if err != nil {
		return err
	}

	return c.AddValidTransaction(tx)
}

// AddValidTransaction adds a valid transaction to the mempool.
func (c *Coin) AddValidTransaction(tx *pb.Transaction) error {
	return c.txpool.AddTransaction(tx)
}

// AppendBlock validates the incoming block and adds it at the end of
// the chain, updating internal state to reflect the block's transactions.
// The miner will be notified of the new block and can decide to mine
// on top of it or keep mining on another fork.
func (c *Coin) AppendBlock(ctx context.Context, block *pb.Block) error {
	// Validate block contents.
	err := c.validator.ValidateBlock(block, nil)
	if err != nil {
		return err
	}

	// Validate block header according to consensus.
	err = c.engine.VerifyHeader(c.chain, block.Header)
	if err != nil {
		return err
	}

	return c.processor.Process(ctx, block, c.state, c.chain)
}

// StartMining starts the underlying miner.
func (c *Coin) StartMining(ctx context.Context) error {
	log.Event(ctx, "StartMiner")
	return c.miner.Start(ctx)
}

// StartTxGossip starts gossiping transactions.
func (c *Coin) StartTxGossip(ctx context.Context) error {
	log.Event(ctx, "StartTxGossip")
	return c.gossip.ListenTx(ctx, c.AddValidTransaction)
}

// StartBlockGossip starts gossiping blocks.
func (c *Coin) StartBlockGossip(ctx context.Context) error {
	log.Event(ctx, "StartBlockGossip")

	return c.gossip.ListenBlock(
		ctx,
		func(block *pb.Block) error {
			return c.AppendBlock(ctx, block)
		},
		func(h []byte) error {
			return c.synchronize(ctx, h)
		},
	)
}

// synchronize synchronizes the local chain.
func (c *Coin) synchronize(ctx context.Context, hash []byte) error {
	event := log.EventBegin(ctx, "Synchronize", logging.Metadata{"hash": hex.EncodeToString(hash)})
	defer event.Done()

	resCh, errCh := c.synchronizer.Synchronize(ctx, hash, c.chain)

	for {
		select {
		case b, ok := <-resCh:
			if !ok {
				// No more blocks to process.
				return nil
			}
			if err := c.AppendBlock(ctx, b); err != nil {
				event.SetError(err)
				return err
			}
		case err := <-errCh:
			event.SetError(err)
			return err
		}
	}
}
