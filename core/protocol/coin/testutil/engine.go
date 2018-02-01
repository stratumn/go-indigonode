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

// DummyEngine is an engine that simply records method calls.
type DummyEngine struct {
	mu       sync.RWMutex
	verified []*pb.Header
}

// VerifyHeader records that header was verified.
func (e *DummyEngine) VerifyHeader(chain chain.Reader, header *pb.Header) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.verified = append(e.verified, header)
	return nil
}

// VerifiedHeader checks if the input header was verified by the engine.
func (e *DummyEngine) VerifiedHeader(header *pb.Header) bool {
	e.mu.RLock()
	defer e.mu.RUnlock()

	matcher := NewHeaderMatcher(header)
	for _, h := range e.verified {
		if matcher.Matches(h) {
			return true
		}
	}

	return false
}

// Prepare does nothing.
func (e *DummyEngine) Prepare(chain chain.Reader, header *pb.Header) error {
	return nil
}

// Finalize does nothing.
func (e *DummyEngine) Finalize(chain chain.Reader, header *pb.Header, state state.Reader, txs []*pb.Transaction) (*pb.Block, error) {
	return nil, nil
}

// ErrFaultyEngine is the error returned by a FaultyEngine.
var ErrFaultyEngine = errors.New("unknown error: engine must be faulty")

// FaultyEngine always returns an ErrFaultyEngine error.
type FaultyEngine struct {
	mu            sync.RWMutex
	verifyCount   int
	prepareCount  int
	finalizeCount int
}

// VerifyHeader records that header was verified.
func (e *FaultyEngine) VerifyHeader(chain chain.Reader, header *pb.Header) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.verifyCount++
	return ErrFaultyEngine
}

// VerifyCount returns the number of calls to VerifyHeader.
func (e *FaultyEngine) VerifyCount() int {
	e.mu.RLock()
	defer e.mu.RUnlock()

	return e.verifyCount
}

// Prepare does nothing.
func (e *FaultyEngine) Prepare(chain chain.Reader, header *pb.Header) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.prepareCount++
	return ErrFaultyEngine
}

// PrepareCount returns the number of calls to Prepare.
func (e *FaultyEngine) PrepareCount() int {
	e.mu.RLock()
	defer e.mu.RUnlock()

	return e.prepareCount
}

// Finalize does nothing.
func (e *FaultyEngine) Finalize(chain chain.Reader, header *pb.Header, state state.Reader, txs []*pb.Transaction) (*pb.Block, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.finalizeCount++
	return nil, ErrFaultyEngine
}

// FinalizeCount returns the number of calls to Finalize.
func (e *FaultyEngine) FinalizeCount() int {
	e.mu.RLock()
	defer e.mu.RUnlock()

	return e.finalizeCount
}

// DummyProofOfWait is a dummy proof-of-wait © engine.
type DummyProofOfWait struct {
	DummyEngine

	min int
	max int

	mu            sync.RWMutex
	intervalCount int
}

// NewDummyProofOfWait creates a new DummyProofOfWait.
func NewDummyProofOfWait(min, max int) *DummyProofOfWait {
	return &DummyProofOfWait{
		min: min,
		max: max,
	}
}

// Interval increments a counter and returns a configurable interval for the proof-of-wait.
func (e *DummyProofOfWait) Interval() (int, int) {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.intervalCount++
	return e.min, e.max
}

// IntervalCount returns the number of calls to Interval.
func (e *DummyProofOfWait) IntervalCount() int {
	e.mu.RLock()
	defer e.mu.RUnlock()

	return e.intervalCount
}
