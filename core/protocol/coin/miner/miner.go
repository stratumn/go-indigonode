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
	"github.com/stratumn/alice/core/protocol/coin/chain"
	"github.com/stratumn/alice/core/protocol/coin/engine"
	"github.com/stratumn/alice/core/protocol/coin/processor"
	"github.com/stratumn/alice/core/protocol/coin/state"
	"github.com/stratumn/alice/core/protocol/coin/validator"
)

// Miner produces new blocks, adds them to the chain and notifies
// other participants of the blocks it produces.
type Miner struct {
	chain     chain.Chain
	engine    engine.Engine
	mempool   state.Mempool
	processor processor.Processor
	state     state.State
	validator validator.Validator
}

// NewMiner creates a new miner that will start mining on the given chain.
func NewMiner(
	m state.Mempool,
	e engine.Engine,
	s state.State,
	c chain.Chain,
	v validator.Validator,
	p processor.Processor) *Miner {
	return &Miner{
		chain:     c,
		engine:    e,
		mempool:   m,
		processor: p,
		state:     s,
		validator: v,
	}
}
