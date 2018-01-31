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
	"sync"
	"time"

	"github.com/stratumn/alice/core/protocol/coin/chain"
	"github.com/stratumn/alice/core/protocol/coin/engine"
	"github.com/stratumn/alice/core/protocol/coin/processor"
	"github.com/stratumn/alice/core/protocol/coin/state"
	"github.com/stratumn/alice/core/protocol/coin/validator"
	pb "github.com/stratumn/alice/pb/coin"

	logging "gx/ipfs/QmSpJByNKFX1sCsHBEp3R73FL4NF6FnQTEGyNAXHm2GS52/go-log"
)

// log is the logger for the miner.
var log = logging.Logger("miner")

// Miner produces new blocks, adds them to the chain and notifies
// other participants of the blocks it produces.
type Miner struct {
	ctx     context.Context
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
	ctx context.Context,
	m state.Mempool,
	e engine.Engine,
	s state.State,
	c chain.Chain,
	v validator.Validator,
	p processor.Processor) *Miner {

	miner := &Miner{
		ctx:       ctx,
		chain:     c,
		engine:    e,
		mempool:   m,
		processor: p,
		state:     s,
		validator: v,
		txsChan:   make(chan []*pb.Transaction),
	}

	go miner.start()
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

// start starts the mining loop.
func (m *Miner) start() {
	m.setRunning(true)
	defer m.setRunning(false)

	m.startTxLoop()

	for {
		select {
		case txs := <-m.txsChan:
			m.produce(txs)
		case <-m.ctx.Done():
			log.Event(m.ctx, "Stopped")
			return
		}
	}
}

// startTxLoop starts the transaction selection process.
// It queries the mempool for transactions and chooses
// a batch of transactions to use for a new block.
func (m *Miner) startTxLoop() {
	if m.mempool == nil {
		log.Event(m.ctx, "NilMempool")
		return
	}

	// For now we simply pop the oldest transaction in the queue.
	// Miners should implement a more sophisticated scheme to be
	// profitable.
	for {
		select {
		case <-m.ctx.Done():
			return
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

func (m *Miner) produce(txs []*pb.Transaction) {
	// TODO
}
