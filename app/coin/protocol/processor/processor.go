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

package processor

import (
	"bytes"
	"context"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/stratumn/go-indigonode/app/coin/pb"
	"github.com/stratumn/go-indigonode/app/coin/protocol/chain"
	"github.com/stratumn/go-indigonode/app/coin/protocol/coinutil"
	"github.com/stratumn/go-indigonode/app/coin/protocol/state"

	cid "gx/ipfs/QmPSQnBKM9g7BaUcZCvswUJVscQ1ipjmwxN5PXCjkp9EQ7/go-cid"
	logging "gx/ipfs/QmSpJByNKFX1sCsHBEp3R73FL4NF6FnQTEGyNAXHm2GS52/go-log"
)

var (
	// log is the logger for the processor.
	log = logging.Logger("coin.processor")

	// errReorgAborted is returned when the reorg is aborted.
	errReorgAborted = errors.New("reorg aborted")

	// ProvideTimeout is the timeout when letting the network we provide a resoruce.
	ProvideTimeout = time.Second * 10
)

// Processor is an interface for processing blocks using a given initial state.
type Processor interface {
	// Process applies the state changes from the block contents
	// and adds the block to the chain.
	Process(ctx context.Context, block *pb.Block, state state.State, ch chain.Chain) error
}

// ContentProvider is an interface used to let the network know we provide a resource.
// The resource is identified by a content ID.
type ContentProvider interface {
	Provide(ctx context.Context, key cid.Cid, brdcst bool) error
}

type processor struct {
	// provider is used to let the network know we have a block once we added it to the local chain.
	provider ContentProvider

	// mu is used to process only one block at a time.
	mu sync.Mutex
}

// NewProcessor creates a new processor.
func NewProcessor(provider ContentProvider) Processor {
	return &processor{provider: provider}
}

func (p *processor) Process(ctx context.Context, block *pb.Block, state state.State, ch chain.Chain) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	mh, err := coinutil.HashHeader(block.Header)
	if err != nil {
		return err
	}

	// Check block has already been processed.
	if _, err := ch.GetBlock(mh, block.BlockNumber()); err != nil && err != chain.ErrBlockNotFound {
		return err
	} else if err == nil {
		log.Event(context.Background(), "BlockAlreadyProcessed", logging.Metadata{"hash": mh.String(), "height": block.BlockNumber()})
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
			log.Event(ctx, "failCastHashToCID", logging.Metadata{"hash": mh.B58String()})
		} else {
			provideCtx, cancel := context.WithTimeout(ctx, ProvideTimeout)
			defer cancel()
			if err = p.provider.Provide(provideCtx, contentID, true); err != nil {
				log.Event(ctx, "failProvide", logging.Metadata{"cid": contentID.String(), "error": err.Error()})
			}
		}
	}

	head, err := ch.CurrentBlock()
	if err != nil && err != chain.ErrBlockNotFound {
		return err
	}

	// If the block is not higher than the current head,
	// do not update state, end processing.
	if head != nil && head.BlockNumber() >= block.BlockNumber() {
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
			if err := p.reorg(head, block, state, ch); err != nil {
				if err == errReorgAborted {
					log.Event(ctx, "ReorgAborted")

					return ch.SetHead(head)
				}

				return err
			}

			return nil
		}
	}

	// Update state.
	err = state.ProcessBlock(block)
	if err != nil {
		log.Event(ctx, "ProcessBlockFailed")
		return ch.SetHead(head)
	}

	return nil
}

// Update the state to follow the new main branch.
func (p *processor) reorg(prevHead *pb.Block, newHead *pb.Block, st state.State, ch chain.Reader) error {
	var cursor *pb.Block

	backward, forward, err := chain.GetPath(ch, prevHead, newHead)
	if err != nil {
		return err
	}

	forward = append(forward, newHead)

	for _, b := range backward {
		if err = st.RollbackBlock(b); err != nil {
			return err
		}
		cursor = b
	}

	for _, b := range forward {
		if err = st.ProcessBlock(b); err != nil {
			// Block has been rejected by state, undo reorg.
			if err == state.ErrInvalidBlock {
				if cursor != nil {
					if err := p.reorg(cursor, prevHead, st, ch); err != nil {
						return err
					}
				}

				return errReorgAborted
			}

			return err
		}
		cursor = b
	}

	return nil
}
