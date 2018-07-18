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

package miner

import (
	"bytes"
	"context"
	"errors"
	"sync"
	"time"

	"github.com/stratumn/go-indigonode/app/coin/pb"
	"github.com/stratumn/go-indigonode/app/coin/protocol/chain"
	"github.com/stratumn/go-indigonode/app/coin/protocol/coinutil"
	"github.com/stratumn/go-indigonode/app/coin/protocol/engine"
	"github.com/stratumn/go-indigonode/app/coin/protocol/gossip"
	"github.com/stratumn/go-indigonode/app/coin/protocol/processor"
	"github.com/stratumn/go-indigonode/app/coin/protocol/state"
	"github.com/stratumn/go-indigonode/app/coin/protocol/validator"

	logging "gx/ipfs/QmSpJByNKFX1sCsHBEp3R73FL4NF6FnQTEGyNAXHm2GS52/go-log"
)

var (
	// ErrTxPoolConfig is returned when the miner's txpool is misconfigured.
	ErrTxPoolConfig = errors.New("txpool is misconfigured: cannot start miner")
)

// log is the logger for the miner.
var log = logging.Logger("coin.miner")

// Miner produces new blocks, adds them to the chain and notifies
// other participants of the blocks it produces.
type Miner struct {
	mu      sync.RWMutex
	running bool

	chain     chain.Chain
	engine    engine.Engine
	gossip    gossip.Gossip
	txpool    state.TxPool
	processor processor.Processor
	state     state.State
	validator validator.Validator

	newBlockChan chan *pb.Header
	txsChan      chan []*pb.Transaction
}

// NewMiner creates a new miner that will start mining on the given chain.
// To stop the miner, you should cancel the input context.
func NewMiner(
	txp state.TxPool,
	e engine.Engine,
	s state.State,
	c chain.Chain,
	v validator.Validator,
	p processor.Processor,
	g gossip.Gossip) *Miner {

	miner := &Miner{
		chain:        c,
		engine:       e,
		gossip:       g,
		txpool:       txp,
		processor:    p,
		state:        s,
		validator:    v,
		newBlockChan: g.AddBlockListener(),
		txsChan:      make(chan []*pb.Transaction),
	}

	return miner
}

// IsRunning returns whether the miner is running or not.
func (m *Miner) IsRunning() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.running
}

func (m *Miner) setRunning(running bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.running = running
}

// Start starts mining to produce blocks.
func (m *Miner) Start(ctx context.Context) error {
	m.setRunning(true)
	defer m.setRunning(false)

	txErrChan := make(chan error)
	go func() {
		txErrChan <- m.startTxLoop(ctx)
	}()

	var miningErr error
miningLoop:
	for {
		select {
		case txs := <-m.txsChan:
			err := m.produce(ctx, txs)
			if err != nil {
				if err == context.Canceled {
					log.Event(ctx, "BlockProductionCanceled")
				} else {
					log.Event(ctx, "BlockProductionFailed", logging.Metadata{"error": err.Error()})
				}
			}
		case <-ctx.Done():
			miningErr = ctx.Err()
			log.Event(ctx, "Stopped", logging.Metadata{"error": miningErr.Error()})
			break miningLoop
		}
	}

	txErr := <-txErrChan
	if miningErr != nil {
		return miningErr
	}

	return txErr
}

// startTxLoop starts the transaction selection process.
// It queries the txpool for transactions and chooses
// a batch of transactions to use for a new block.
func (m *Miner) startTxLoop(ctx context.Context) error {
	if m.txpool == nil {
		log.Event(ctx, "NilTxPool")
		return ErrTxPoolConfig
	}

	// For now we simply pop as many transactions as possible from
	// the txpool, ordered by their score.
	// It's a simple way to get good transaction fees.
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			var txs []*pb.Transaction
			for {
				if uint32(len(txs)) == m.validator.MaxTxPerBlock() {
					break
				}

				tx := m.txpool.PopTransaction()
				if tx == nil {
					break
				}

				txs = append(txs, tx)
			}

			if len(txs) == 0 {
				<-time.After(10 * time.Millisecond)
			} else {
				m.txsChan <- txs
			}
		}
	}
}

// produce produces a new block from the input transactions.
// It returns when it has finished producing the block or when
// a new block is advertised (which makes the current work obsolete).
func (m *Miner) produce(ctx context.Context, txs []*pb.Transaction) (err error) {
	defer func() {
		if err != nil {
			m.putBackInTxPool(ctx, txs)
		}
	}()

	if err = m.validator.ValidateTransactions(txs, m.state); err != nil {
		return
	}

	block := &pb.Block{Transactions: txs}

	header := &pb.Header{}
	if err = m.engine.Prepare(m.chain, header); err != nil {
		return
	}

	powCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	powSuccessChan := make(chan *pb.Block)
	powErrorChan := make(chan error)

	go func() {
		block, err = m.engine.Finalize(powCtx, m.chain, header, m.state, txs)
		if err != nil {
			powErrorChan <- err
		} else {
			log.Event(powCtx, "PowSuccess", &logging.Metadata{"nonce": block.Nonce()})
			powSuccessChan <- block
		}
	}()

	for {
		select {
		case incoming := <-m.newBlockChan:
			if m.headUpdated(powCtx, incoming, header.PreviousHash) {
				err = context.Canceled
				return
			}
		case block := <-powSuccessChan:
			err = m.processNewBlock(ctx, block)
			if err == nil {
				log.Event(ctx, "NewBlockProduced", &logging.Metadata{"block": block.Loggable()})
			}
			return
		case err = <-powErrorChan:
			return
		}
	}
}

// processNewBlock takes a finalized block, processes it and publishes it.
func (m *Miner) processNewBlock(ctx context.Context, block *pb.Block) error {
	if err := m.gossip.PublishBlock(block); err != nil {
		log.Event(ctx, "PublishBlockFailure", &logging.Metadata{
			"error": err.Error(),
			"block": block.Loggable(),
		})
	}

	return m.processor.Process(ctx, block, m.state, m.chain)
}

// putBackInTxPool puts back transactions into the txpool.
// It discards invalid transactions.
func (m *Miner) putBackInTxPool(ctx context.Context, txs []*pb.Transaction) {
	for _, tx := range txs {
		if err := m.validator.ValidateTx(tx, m.state); err == nil {
			err := m.txpool.AddTransaction(tx)
			if err != nil {
				log.Event(ctx, "AddTransactionFailure", &logging.Metadata{
					"error": err.Error(),
					"tx":    tx.Loggable(),
				})
			}
		}
	}
}

// headUpdated returns true if the given header
// is an updated head that we aren't currently mining on.
func (m *Miner) headUpdated(ctx context.Context, incoming *pb.Header, previousHash []byte) bool {
	current, err := m.chain.CurrentHeader()
	if err != nil {
		log.Event(ctx, "CurrentHeaderError", &logging.Metadata{"error": err.Error()})
		return false
	}

	incomingHash, err := coinutil.HashHeader(incoming)
	if err != nil {
		return false
	}

	currentHash, err := coinutil.HashHeader(current)
	if err != nil {
		return false
	}

	log.Event(ctx, "HeaderReceived", &logging.Metadata{
		"head": current.Loggable(),
		"new":  incoming.Loggable(),
	})

	// If incoming wasn't set as the chain head, we can ignore.
	if !bytes.Equal(incomingHash, currentHash) {
		return false
	}

	// If we're already mining on top of the incoming header, we can ignore.
	if bytes.Equal(incomingHash, previousHash) {
		return false
	}

	return true
}
