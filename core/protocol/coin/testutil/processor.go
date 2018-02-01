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
	"sync"

	"github.com/stratumn/alice/core/protocol/coin/chain"
	"github.com/stratumn/alice/core/protocol/coin/state"
	pb "github.com/stratumn/alice/pb/coin"
)

// DummyProcessor is a dummy processor that always returns nil
// as if processing ended correctly.
type DummyProcessor struct {
	mu             sync.RWMutex
	processedCount int
}

// Process increments a counter and returns nil.
// Simulates a processing success.
func (p *DummyProcessor) Process(block *pb.Block, state state.Writer, chain chain.Writer) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.processedCount++
	return nil
}

// ProcessedCount returns the number of blocks processed.
func (p *DummyProcessor) ProcessedCount() int {
	p.mu.RLock()
	defer p.mu.RUnlock()

	return p.processedCount
}

// ErrProcessingFailed is the error returned by FaultyProcessor.
var ErrProcessingFailed = errors.New("processing failed")

// FaultyProcessor is a dummy processor that always returns an error
// as if processing failed.
type FaultyProcessor struct {
	mu             sync.RWMutex
	processedCount int
}

// Process increments a counter and returns an error.
// Simulates a processing failure.
func (p *FaultyProcessor) Process(block *pb.Block, state state.Writer, chain chain.Writer) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.processedCount++
	return ErrProcessingFailed
}

// ProcessedCount returns the number of blocks processed.
func (p *FaultyProcessor) ProcessedCount() int {
	p.mu.RLock()
	defer p.mu.RUnlock()

	return p.processedCount
}
