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

//go:generate mockgen -package mockprocessor -destination mockprocessor/mockprocessor.go github.com/stratumn/alice/core/protocol/coin/processor Processor

package processor

import (
	"crypto/sha256"

	"github.com/stratumn/alice/core/protocol/coin/chain"
	"github.com/stratumn/alice/core/protocol/coin/state"
	pb "github.com/stratumn/alice/pb/coin"
)

// Processor is an interface for processing blocks using a given initial state.
type Processor interface {
	// Process applies the state changes from the block contents
	// and adds the block to the chain.
	Process(block *pb.Block, state state.State, chain chain.Writer) error
}

type processor struct{}

// NewProcessor creates a new processor.
func NewProcessor() Processor {
	return processor{}
}

func (processor) Process(block *pb.Block, state state.State, chain chain.Writer) error {
	// TODO: miner reward.

	// Update chain.
	if err := chain.AddBlock(block); err != nil {
		return err
	}
	if err := chain.SetHead(block); err != nil {
		return err
	}

	// Update state.
	// TODO: instead of recomputing the hash here,
	// we can have chain.AddBlock return it.
	headerBytes, err := block.Header.Marshal()
	if err != nil {
		return err
	}
	h := sha256.Sum256(headerBytes)

	return state.ProcessTransactions(h[:], block.Transactions)
}
