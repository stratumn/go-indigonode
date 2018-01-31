// Copyright © 2017-2018 Stratumn SAS
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package coin

import (
	"context"
	"fmt"
	"io"

	"github.com/stratumn/alice/core/protocol/coin/chain"
	"github.com/stratumn/alice/core/protocol/coin/engine"
	"github.com/stratumn/alice/core/protocol/coin/miner"
	"github.com/stratumn/alice/core/protocol/coin/processor"
	"github.com/stratumn/alice/core/protocol/coin/state"
	"github.com/stratumn/alice/core/protocol/coin/validator"
	pb "github.com/stratumn/alice/pb/coin"

	inet "gx/ipfs/QmQm7WmgYCa4RSz76tKEYpRjApjnRw8ZTUVQC15b8JM4a2/go-libp2p-net"
	protobuf "gx/ipfs/QmRDePEiL4Yupq5EkcK3L3ko3iMgYaqUdLu7xc1kqs7dnV/go-multicodec/protobuf"
	logging "gx/ipfs/QmSpJByNKFX1sCsHBEp3R73FL4NF6FnQTEGyNAXHm2GS52/go-log"
	protocol "gx/ipfs/QmZNkThpqfVXs9GNbexPrfBbXSLNYeKrE7jwFM2oqHbyqN/go-libp2p-protocol"
)

// ProtocolID is the protocol ID of the protocol.
var ProtocolID = protocol.ID("/alice/coin/v1.0.0")

// log is the logger for the service.
var log = logging.Logger("coin")

// Protocol describes the interface exposed to other nodes in the network.
type Protocol interface {
	// AddTransaction validates a transaction and adds it to the mempool.
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
	engine    engine.Engine
	mempool   state.Mempool
	processor processor.Processor
	validator validator.Validator

	miner *miner.Miner
}

// NewCoin creates a new Coin.
func NewCoin(
	m state.Mempool,
	e engine.Engine,
	s state.State,
	c chain.Chain,
	v validator.Validator,
	p processor.Processor) *Coin {

	miner := miner.NewMiner(m, e, s, c, v, p)

	return &Coin{
		engine:    e,
		mempool:   m,
		processor: p,
		validator: v,
		miner:     miner,
	}
}

// StreamHandler handles incoming messages from peers.
func (c *Coin) StreamHandler(ctx context.Context, stream inet.Stream) {
	log.Event(ctx, "beginStream", logging.Metadata{
		"stream": stream,
	})
	defer log.Event(ctx, "endStream", logging.Metadata{
		"stream": stream,
	})

	dec := protobuf.Multicodec(nil).Decoder(stream)

	for {
		var gossip pb.Gossip
		err := dec.Decode(&gossip)
		if err == io.EOF {
			return
		}
		if err != nil {
			log.Event(ctx, "Decode", logging.Metadata{"error": err})
			continue
		}

		switch m := gossip.Msg.(type) {
		case *pb.Gossip_Tx:
			err := c.AddTransaction(m.Tx)
			if err != nil {
				log.Event(ctx, "AddTransaction", logging.Metadata{"error": err})
			}
		case *pb.Gossip_Block:
			err := c.AppendBlock(m.Block)
			if err != nil {
				log.Event(ctx, "AppendBlock", logging.Metadata{"error": err})
			}
		default:
			log.Event(ctx, "Gossip", logging.Metadata{
				"error": "Unexpected type",
				"type":  fmt.Sprintf("%T", m),
			})
		}
	}
}

// AddTransaction validates incoming transactions against the latest state
// and adds them to the mempool.
func (c *Coin) AddTransaction(tx *pb.Transaction) error {
	err := c.validator.ValidateTx(tx, nil)
	if err != nil {
		return err
	}

	return c.mempool.AddTransaction(tx)
}

// AppendBlock validates the incoming block and adds it at the end of
// the chain, updating internal state to reflect the block's transactions.
// The miner will be notified of the new block and can decide to mine
// on top of it or keep mining on another fork.
func (c *Coin) AppendBlock(block *pb.Block) error {
	// Validate block contents.
	err := c.validator.ValidateBlock(block, nil)
	if err != nil {
		return err
	}

	// Validate block header according to consensus.
	err = c.engine.VerifyHeader(nil, block.Header)
	if err != nil {
		return err
	}

	return c.processor.Process(block, nil, nil)
}

// StartMining starts the underlying miner.
func (c *Coin) StartMining(ctx context.Context) error {
	return c.miner.Start(ctx)
}
