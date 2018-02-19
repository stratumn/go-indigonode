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

package miner

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/stratumn/alice/core/protocol/coin/chain"
	"github.com/stratumn/alice/core/protocol/coin/engine"
	"github.com/stratumn/alice/core/protocol/coin/gossip"
	"github.com/stratumn/alice/core/protocol/coin/processor"
	"github.com/stratumn/alice/core/protocol/coin/state"
	"github.com/stratumn/alice/core/protocol/coin/validator"
	pb "github.com/stratumn/alice/pb/coin"

	logging "gx/ipfs/QmSpJByNKFX1sCsHBEp3R73FL4NF6FnQTEGyNAXHm2GS52/go-log"
)

var (
	// ErrTxPoolConfig is returned when the miner's txpool is misconfigured.
	ErrTxPoolConfig = errors.New("txpool is misconfigured: cannot start miner")
)

// log is the logger for the miner.
var log = logging.Logger("miner")

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

	txsChan chan []*pb.Transaction
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
		chain:     c,
		engine:    e,
		gossip:    g,
		txpool:    txp,
		processor: p,
		state:     s,
		validator: v,
		txsChan:   make(chan []*pb.Transaction),
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
				log.Event(ctx, "BlockProductionFailed", logging.Metadata{"error": err.Error()})
			} else {
				log.Event(ctx, "NewBlockProduced")
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

	block, err = m.engine.Finalize(ctx, m.chain, header, m.state, txs)
	if err != nil {
		return
	}

	if err := m.gossip.PublishBlock(block); err != nil {
		log.Event(ctx, "PublishBlockFailure", &logging.Metadata{"error": err.Error()})
	}

	err = m.processor.Process(ctx, block, m.state, m.chain)
	if err != nil {
		return
	}

	return
}

// putBackInTxPool puts back transactions into the txpool.
// It discards invalid transactions.
func (m *Miner) putBackInTxPool(ctx context.Context, txs []*pb.Transaction) {
	for _, tx := range txs {
		if err := m.validator.ValidateTx(tx, m.state); err == nil {
			err := m.txpool.AddTransaction(tx)
			if err != nil {
				log.Event(ctx, "AddTransactionFailure", &logging.Metadata{"error": err.Error()})
			}
		}
	}
}
