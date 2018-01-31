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
	"github.com/stratumn/alice/core/protocol/coin/chain"
	"github.com/stratumn/alice/core/protocol/coin/state"
	pb "github.com/stratumn/alice/pb/coin"
)

// Processor is an interface for processing blocks using a given initial state.
type Processor interface {
	// Process applies the state changes from the block contents
	// and adds the block to the chain.
	Process(block *pb.Block, state state.State, chain chain.Writer) error

	// Rollback rolls back a block, setting the chain head to the previous
	// block and updating the state accordingly by discarding the
	// transactions contained in the block.
	//
	// The block arg is temporary. The block should be the head of the
	// chain instead of a parameter.
	//
	//	block := chain.GetBlock(something, chain.CurrentHeader().BlockNumber)
	Rollback(block *pb.Block, state state.State, chain chain.Chain) error

	// RollbackTo rolls back to the block with the given hash, updating the
	// chain head and updating the state accordingly by discarding
	// transactions from blocks being rolled back.
	RollbackTo(blockHash []byte, state state.State, chain chain.Chain) error
}

type processor struct{}

// NewProcessor creates a new processor.
func NewProcessor() Processor {
	return processor{}
}

func (processor) Process(block *pb.Block, state state.State, chain chain.Writer) error {
	// TODO: update chain, miner reward.

	stateTx, err := state.Transaction()
	if err != nil {
		return err
	}

	for _, tx := range block.Transactions {
		if err := stateTx.Transfer(tx.From, tx.To, tx.Value); err != nil {
			stateTx.Discard()
			return err
		}
	}

	return stateTx.Commit()
}

func (processor) Rollback(block *pb.Block, state state.State, chain chain.Chain) error {
	// This is temporary. The block should be the head of the chain
	// instead of a parameter.
	//
	//	block := chain.GetBlock(nil, chain.CurrentHeader().BlockNumber)

	stateTx, err := state.Transaction()
	if err != nil {
		return err
	}

	for i := len(block.Transactions) - 1; i >= 0; i-- {
		tx := block.Transactions[i]

		if err := stateTx.Transfer(tx.To, tx.From, tx.Value); err != nil {
			stateTx.Discard()
			return err
		}
	}

	return stateTx.Commit()

	// TODO: update chain head, miner reward.
}

func (processor) RollbackTo(blockHash []byte, state state.State, chain chain.Chain) error {
	// TODO: loopback from head and call rollback until target block is
	// reached.

	return nil
}
