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
	"bytes"

	"github.com/stratumn/alice/core/protocol/coin/chain"
	"github.com/stratumn/alice/core/protocol/coin/coinutil"
	"github.com/stratumn/alice/core/protocol/coin/state"
	pb "github.com/stratumn/alice/pb/coin"
)

// Processor is an interface for processing blocks using a given initial state.
type Processor interface {
	// Process applies the state changes from the block contents
	// and adds the block to the chain.
	Process(block *pb.Block, state state.State, chain chain.Chain) error
}

type processor struct{}

// NewProcessor creates a new processor.
func NewProcessor() Processor {
	return &processor{}
}

type stateTransition struct {
	stateID      []byte
	transactions []*pb.Transaction
}

func (p *processor) Process(block *pb.Block, state state.State, chain chain.Chain) error {
	// Update chain.
	if err := chain.AddBlock(block); err != nil {
		return err
	}

	head, err := chain.CurrentBlock()
	if err != nil {
		return err
	}

	// If the block is not higher than the current head,
	// do not update state, end processing.
	if head != nil && head.Header.BlockNumber > block.Header.BlockNumber {
		return nil
	}

	// Set the new head.
	if err := chain.SetHead(block); err != nil {
		return err
	}

	// If the new head is on a different branch, reorganize state.
	if head != nil {
		hash, err := coinutil.HashHeader(head.Header)
		if err != nil {
			return err
		}
		if !bytes.Equal(hash, block.Header.PreviousHash) {
			return p.reorg(head, block, state, chain)
		}
	}

	// Update state.
	h, err := coinutil.HashHeader(block.Header)
	if err != nil {
		return err
	}
	return state.ProcessTransactions(h, block.Transactions)
}

// Update the state to follow the new main branch.
func (p *processor) reorg(prevHead *pb.Block, newHead *pb.Block, state state.State, chain chain.Chain) error {
	backward := []*stateTransition{}
	forward := []*stateTransition{}

	mainBlock := newHead
	mainNumber := mainBlock.Header.BlockNumber
	mainHash, err := coinutil.HashHeader(mainBlock.Header)
	if err != nil {
		return err
	}

	forkBlock := prevHead
	forkNumber := forkBlock.Header.BlockNumber
	forkHash, err := coinutil.HashHeader(forkBlock.Header)
	if err != nil {
		return err
	}

	// First gets the two chains to the same height.
	// Then go back to first common ancestor.
	for !bytes.Equal(mainHash, forkHash) {
		if mainNumber >= forkNumber {
			forward = append(forward, &stateTransition{stateID: mainHash, transactions: mainBlock.Transactions})
			mainHash = mainBlock.Header.PreviousHash
			mainBlock, err = chain.GetBlock(mainBlock.Header.PreviousHash, mainBlock.Header.BlockNumber-1)
			if err != nil {
				return err
			}
		}

		if mainNumber <= forkNumber {
			// Prepend transition to backward to be then played in good order.
			backward = append(
				[]*stateTransition{&stateTransition{stateID: forkHash, transactions: forkBlock.Transactions}},
				backward...,
			)
			forkHash = forkBlock.Header.PreviousHash
			forkBlock, err = chain.GetBlock(forkBlock.Header.PreviousHash, forkBlock.Header.BlockNumber-1)
			if err != nil {
				return err
			}
		}
		mainNumber = mainBlock.Header.BlockNumber
		forkNumber = forkBlock.Header.BlockNumber
	}

	for _, st := range backward {
		if err = state.RollbackTransactions(st.stateID, st.transactions); err != nil {
			return err
		}
	}

	for _, st := range forward {
		if err = state.ProcessTransactions(st.stateID, st.transactions); err != nil {
			return err
		}
	}
	return nil
}
