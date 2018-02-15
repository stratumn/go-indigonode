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
	"context"

	"github.com/stratumn/alice/core/protocol/coin/chain"
	"github.com/stratumn/alice/core/protocol/coin/coinutil"
	"github.com/stratumn/alice/core/protocol/coin/state"
	pb "github.com/stratumn/alice/pb/coin"

	logging "gx/ipfs/QmSpJByNKFX1sCsHBEp3R73FL4NF6FnQTEGyNAXHm2GS52/go-log"
	cid "gx/ipfs/QmcZfnkapfECQGcLZaf9B79NRg7cRa9EnZh4LSbkCzwNvY/go-cid"
)

// log is the logger for the processor.
var log = logging.Logger("processor")

// Processor is an interface for processing blocks using a given initial state.
type Processor interface {
	// Process applies the state changes from the block contents
	// and adds the block to the chain.
	Process(ctx context.Context, block *pb.Block, state state.State, ch chain.Chain) error
}

// ContentProvider is an interface used to let the network know we provide a resource.
// The resource is identified by a content ID.
type ContentProvider interface {
	Provide(ctx context.Context, key *cid.Cid, brdcst bool) error
}

type processor struct {
	// provider is used to let the network know we have a block once we added it to the local chain.
	provider ContentProvider
}

// NewProcessor creates a new processor.
func NewProcessor(provider ContentProvider) Processor {
	return &processor{provider: provider}
}

type stateTransition struct {
	stateID      []byte
	transactions []*pb.Transaction
}

func (p *processor) Process(ctx context.Context, block *pb.Block, state state.State, ch chain.Chain) error {
	mh, err := coinutil.HashHeader(block.Header)
	if err != nil {
		return err
	}

	// Check block has already been processed.
	if _, err := ch.GetBlock(mh, block.BlockNumber()); err != nil && err != chain.ErrBlockHashNotFound {
		return err
	} else if err == nil {
		log.Event(context.Background(), "blockAlreadyProcessed", logging.Metadata{"hash": mh.String(), "height": block.BlockNumber()})
		return nil
	}

	// Update chain.
	if err := ch.AddBlock(block); err != nil {
		return err
	}

	// Tell the network we have that block.
	if p.provider != nil {
		contentID, err := cid.Cast(mh)
		if err != nil {
			log.Event(ctx, "failCastHashToCID", logging.Metadata{"hash": mh.String()})
		}
		if err = p.provider.Provide(ctx, contentID, true); err != nil {
			log.Event(ctx, "failProvide", logging.Metadata{"cid": contentID.String(), "error": err.Error()})
		}
	}

	head, err := ch.CurrentBlock()
	if err != nil {
		return err
	}

	// If the block is not higher than the current head,
	// do not update state, end processing.
	if head != nil && head.BlockNumber() > block.BlockNumber() {
		return nil
	}

	// Set the new head.
	if err := ch.SetHead(block); err != nil {
		return err
	}

	// If the new head is on a different branch, reorganize state.
	if head != nil {
		hash, err := coinutil.HashHeader(head.Header)
		if err != nil {
			return err
		}
		if !bytes.Equal(hash, block.PreviousHash()) {
			return p.reorg(head, block, state, ch)
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
	mainNumber := mainBlock.BlockNumber()
	mainHash, err := coinutil.HashHeader(mainBlock.Header)
	if err != nil {
		return err
	}

	forkBlock := prevHead
	forkNumber := forkBlock.BlockNumber()
	forkHash, err := coinutil.HashHeader(forkBlock.Header)
	if err != nil {
		return err
	}

	// First gets the two chains to the same height.
	// Then go back to first common ancestor.
	for !bytes.Equal(mainHash, forkHash) {
		if mainNumber >= forkNumber {
			// Prepend transition to forward to be then played in good order.
			forward = append(
				[]*stateTransition{&stateTransition{stateID: mainHash, transactions: mainBlock.Transactions}},
				forward...,
			)
			mainHash = mainBlock.PreviousHash()
			mainBlock, err = chain.GetBlock(mainBlock.PreviousHash(), mainBlock.BlockNumber()-1)
			if err != nil {
				return err
			}
		}

		if mainNumber <= forkNumber {
			backward = append(backward, &stateTransition{stateID: forkHash, transactions: forkBlock.Transactions})
			forkHash = forkBlock.PreviousHash()
			forkBlock, err = chain.GetBlock(forkBlock.PreviousHash(), forkBlock.BlockNumber()-1)
			if err != nil {
				return err
			}
		}
		mainNumber = mainBlock.BlockNumber()
		forkNumber = forkBlock.BlockNumber()
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
