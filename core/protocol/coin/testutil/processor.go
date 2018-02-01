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

package testutil

import (
	"errors"
	"sync/atomic"

	"github.com/stratumn/alice/core/protocol/coin/chain"
	"github.com/stratumn/alice/core/protocol/coin/processor"
	"github.com/stratumn/alice/core/protocol/coin/state"
	pb "github.com/stratumn/alice/pb/coin"
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
func (p *InstrumentedProcessor) Process(block *pb.Block, state state.State, chain chain.Writer) error {
	atomic.AddUint32(&p.processedCount, 1)
	return p.processor.Process(block, state, chain)
}

// ProcessedCount returns the number of blocks processed.
func (p *InstrumentedProcessor) ProcessedCount() uint32 {
	return p.processedCount
}

// DummyProcessor is a dummy processor that always returns nil
// as if processing ended correctly.
type DummyProcessor struct{}

// Process returns nil to simulate a processing success.
func (p *DummyProcessor) Process(block *pb.Block, state state.State, chain chain.Writer) error {
	return nil
}

// ErrProcessingFailed is the error returned by FaultyProcessor.
var ErrProcessingFailed = errors.New("processing failed")

// FaultyProcessor is a dummy processor that always returns an error
// as if processing failed.
type FaultyProcessor struct{}

// Process returns an error to simulate a processing failure.
func (p *FaultyProcessor) Process(block *pb.Block, state state.State, chain chain.Writer) error {
	return ErrProcessingFailed
}
