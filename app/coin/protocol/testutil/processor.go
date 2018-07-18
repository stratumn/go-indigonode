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

package testutil

import (
	"context"
	"errors"
	"sync/atomic"

	"github.com/stratumn/go-indigonode/app/coin/pb"
	"github.com/stratumn/go-indigonode/app/coin/protocol/chain"
	"github.com/stratumn/go-indigonode/app/coin/protocol/processor"
	"github.com/stratumn/go-indigonode/app/coin/protocol/state"
)

// InstrumentedProcessor adds method calls instrumentation to a processor.
type InstrumentedProcessor struct {
	processor      processor.Processor
	processedCount uint32
}

// NewInstrumentedProcessor wraps a processor with instrumentation.
func NewInstrumentedProcessor(p processor.Processor) *InstrumentedProcessor {
	return &InstrumentedProcessor{processor: p}
}

// Process records the call.
func (p *InstrumentedProcessor) Process(ctx context.Context, block *pb.Block, state state.State, chain chain.Chain) error {
	atomic.AddUint32(&p.processedCount, 1)
	return p.processor.Process(ctx, block, state, chain)
}

// ProcessedCount returns the number of blocks processed.
func (p *InstrumentedProcessor) ProcessedCount() uint32 {
	return p.processedCount
}

// DummyProcessor is a dummy processor that always returns nil
// as if processing ended correctly.
type DummyProcessor struct{}

// Process returns nil to simulate a processing success.
func (p *DummyProcessor) Process(ctx context.Context, block *pb.Block, state state.State, chain chain.Chain) error {
	return nil
}

// ErrProcessingFailed is the error returned by FaultyProcessor.
var ErrProcessingFailed = errors.New("processing failed")

// FaultyProcessor is a dummy processor that always returns an error
// as if processing failed.
type FaultyProcessor struct{}

// Process returns an error to simulate a processing failure.
func (p *FaultyProcessor) Process(ctx context.Context, block *pb.Block, state state.State, chain chain.Chain) error {
	return ErrProcessingFailed
}
