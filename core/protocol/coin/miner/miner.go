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
	"math"
	"sync"
	"time"

	"github.com/stratumn/alice/core/protocol/coin/chain"
	"github.com/stratumn/alice/core/protocol/coin/coinutil"
	"github.com/stratumn/alice/core/protocol/coin/engine"
	"github.com/stratumn/alice/core/protocol/coin/processor"
	"github.com/stratumn/alice/core/protocol/coin/state"
	"github.com/stratumn/alice/core/protocol/coin/validator"
	pb "github.com/stratumn/alice/pb/coin"

	logging "gx/ipfs/QmSpJByNKFX1sCsHBEp3R73FL4NF6FnQTEGyNAXHm2GS52/go-log"
)

var (
	// ErrMempoolConfig is returned when the miner's mempool is misconfigured.
	ErrMempoolConfig = errors.New("mempool is misconfigured: cannot start miner")
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
	mempool   state.Mempool
	processor processor.Processor
	state     state.State
	validator validator.Validator

	txsChan chan []*pb.Transaction
}

// NewMiner creates a new miner that will start mining on the given chain.
// To stop the miner, you should cancel the input context.
func NewMiner(
	m state.Mempool,
	e engine.Engine,
	s state.State,
	c chain.Chain,
	v validator.Validator,
	p processor.Processor) *Miner {

	miner := &Miner{
		chain:     c,
		engine:    e,
		mempool:   m,
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
				log.Event(ctx, "BlockProductionFailed", logging.Metadata{"error": err})
			} else {
				log.Event(ctx, "NewBlockProduced")
			}
		case <-ctx.Done():
			miningErr = ctx.Err()
			log.Event(ctx, "Stopped", logging.Metadata{"error": miningErr})
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
// It queries the mempool for transactions and chooses
// a batch of transactions to use for a new block.
func (m *Miner) startTxLoop(ctx context.Context) error {
	if m.mempool == nil {
		log.Event(ctx, "NilMempool")
		return ErrMempoolConfig
	}

	// For now we simply pop the oldest transaction in the queue.
	// Miners should implement a more sophisticated scheme to be
	// profitable.
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			tx := m.mempool.PopTransaction()
			if tx == nil {
				<-time.After(10 * time.Millisecond)
			} else {
				m.txsChan <- []*pb.Transaction{tx}
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
			m.putBackInMempool(txs)
		}
	}()

	block := &pb.Block{Transactions: txs}
	if err = m.validator.ValidateBlock(block, m.state); err != nil {
		return
	}

	header := &pb.Header{}
	if err = m.engine.Prepare(m.chain, header); err != nil {
		return
	}

	block, err = m.engine.Finalize(m.chain, header, m.state, txs)
	if err != nil {
		return
	}

	powEngine, ok := m.engine.(engine.PoW)
	if ok {
		m.pow(ctx, powEngine, block)
	}

	err = m.processor.Process(block, m.state, m.chain)
	if err != nil {
		return
	}

	return
}

// pow finds a solution to the proof of work problem.
// It sets the block header's nonce appropriately.
func (m *Miner) pow(ctx context.Context, e engine.PoW, block *pb.Block) {
	difficulty := e.Difficulty()
	for i := uint64(0); i < math.MaxUint64; i++ {
		block.Header.Nonce = i
		b, err := coinutil.HashHeader(block.Header)
		if err != nil {
			log.Event(ctx, "PoWError", logging.Metadata{"error": err})
			continue
		}

		difficultyMet := true
		for j := uint64(0); j < difficulty; j++ {
			if b[j] != 0 {
				difficultyMet = false
				break
			}
		}

		if difficultyMet {
			break
		}
	}
}

// putBackInMempool puts back transactions into the mempool.
// It discards invalid transactions.
func (m *Miner) putBackInMempool(txs []*pb.Transaction) {
	for _, tx := range txs {
		if err := m.validator.ValidateTx(tx, m.state); err == nil {
			err := m.mempool.AddTransaction(tx)
			if err != nil {
				log.Debugf("couldn't add tx to mempool: %s", err.Error())
			}
		}
	}
}
