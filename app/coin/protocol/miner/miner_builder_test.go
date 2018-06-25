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
	"testing"

	"github.com/stratumn/go-indigonode/app/coin/protocol/chain"
	"github.com/stratumn/go-indigonode/app/coin/protocol/engine"
	"github.com/stratumn/go-indigonode/app/coin/protocol/gossip"
	"github.com/stratumn/go-indigonode/app/coin/protocol/processor"
	"github.com/stratumn/go-indigonode/app/coin/protocol/state"
	"github.com/stratumn/go-indigonode/app/coin/protocol/testutil"
	"github.com/stratumn/go-indigonode/app/coin/protocol/validator"
)

// MinerBuilder is a utility to create miners with custom mocks
// to simulate a wide variety of miner configuration.
type MinerBuilder struct {
	chain     chain.Chain
	engine    engine.Engine
	gossip    gossip.Gossip
	txpool    state.TxPool
	processor processor.Processor
	state     state.State
	validator validator.Validator
}

// NewMinerBuilder creates a MinerBuilder with a context and
// good default values to build a test Miner.
func NewMinerBuilder(t *testing.T) *MinerBuilder {
	return &MinerBuilder{
		chain:     &testutil.SimpleChain{},
		engine:    &testutil.DummyEngine{},
		gossip:    testutil.NewDummyGossip(t),
		txpool:    &testutil.InMemoryTxPool{},
		processor: &testutil.DummyProcessor{},
		validator: &testutil.DummyValidator{},
	}
}

// WithChain configures the builder to use the given chain.
func (m *MinerBuilder) WithChain(chain chain.Chain) *MinerBuilder {
	m.chain = chain
	return m
}

// WithEngine configures the builder to use the given engine.
func (m *MinerBuilder) WithEngine(engine engine.Engine) *MinerBuilder {
	m.engine = engine
	return m
}

// WithTxPool configures the builder to use the given txpool.
func (m *MinerBuilder) WithTxPool(txpool state.TxPool) *MinerBuilder {
	m.txpool = txpool
	return m
}

// WithProcessor configures the builder to use the given processor.
func (m *MinerBuilder) WithProcessor(processor processor.Processor) *MinerBuilder {
	m.processor = processor
	return m
}

// WithValidator configures the builder to use the given validator.
func (m *MinerBuilder) WithValidator(validator validator.Validator) *MinerBuilder {
	m.validator = validator
	return m
}

// WithGossip configures the builder to use the given gossip.
func (m *MinerBuilder) WithGossip(gossip gossip.Gossip) *MinerBuilder {
	m.gossip = gossip
	return m
}

// Build builds the underlying miner and returns it.
// It's now ready to use in your tests.
func (m *MinerBuilder) Build() *Miner {
	return NewMiner(
		m.txpool,
		m.engine,
		m.state,
		m.chain,
		m.validator,
		m.processor,
		m.gossip,
	)
}
