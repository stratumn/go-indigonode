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
	"context"
	"errors"
	"sync"
	"sync/atomic"

	"github.com/stratumn/alice/core/protocol/coin/chain"
	"github.com/stratumn/alice/core/protocol/coin/engine"
	"github.com/stratumn/alice/core/protocol/coin/state"
	pb "github.com/stratumn/alice/pb/coin"
)

// InstrumentedEngine adds method calls instrumentation to an engine.
type InstrumentedEngine struct {
	engine engine.Engine

	mu       sync.RWMutex
	verified []*pb.Header

	verifyCount   uint32
	prepareCount  uint32
	finalizeCount uint32
}

// NewInstrumentedEngine wraps an engine with instrumentation.
func NewInstrumentedEngine(e engine.Engine) *InstrumentedEngine {
	return &InstrumentedEngine{engine: e}
}

// VerifyHeader records that header was verified.
func (e *InstrumentedEngine) VerifyHeader(chain chain.Reader, header *pb.Header) error {
	atomic.AddUint32(&e.verifyCount, 1)

	e.mu.Lock()
	defer e.mu.Unlock()

	e.verified = append(e.verified, header)
	return e.engine.VerifyHeader(chain, header)
}

// VerifiedHeader checks if the input header was verified by the engine.
func (e *InstrumentedEngine) VerifiedHeader(header *pb.Header) bool {
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

// VerifyCount returns the number of calls to VerifyHeader.
func (e *InstrumentedEngine) VerifyCount() uint32 {
	return e.verifyCount
}

// Prepare records the call.
func (e *InstrumentedEngine) Prepare(chain chain.Reader, header *pb.Header) error {
	atomic.AddUint32(&e.prepareCount, 1)
	return e.engine.Prepare(chain, header)
}

// PrepareCount returns the number of calls to Prepare.
func (e *InstrumentedEngine) PrepareCount() uint32 {
	return e.prepareCount
}

// Finalize records the call.
func (e *InstrumentedEngine) Finalize(ctx context.Context, chain chain.Reader, header *pb.Header, state state.Reader, txs []*pb.Transaction) (*pb.Block, error) {
	atomic.AddUint32(&e.prepareCount, 1)
	return e.engine.Finalize(ctx, chain, header, state, txs)
}

// FinalizeCount returns the number of calls to Finalize.
func (e *InstrumentedEngine) FinalizeCount() uint32 {
	return e.finalizeCount
}

// DummyEngine is an engine that does nothing.
type DummyEngine struct{}

// VerifyHeader does nothing.
func (e *DummyEngine) VerifyHeader(chain chain.Reader, header *pb.Header) error {
	return nil
}

// Prepare does nothing.
func (e *DummyEngine) Prepare(chain chain.Reader, header *pb.Header) error {
	return nil
}

// Finalize returns an empty block with the given header.
func (e *DummyEngine) Finalize(ctx context.Context, chain chain.Reader, header *pb.Header, state state.Reader, txs []*pb.Transaction) (*pb.Block, error) {
	block := &pb.Block{Header: header}
	return block, nil
}

// ErrFaultyEngine is the error returned by a FaultyEngine.
var ErrFaultyEngine = errors.New("unknown error: engine must be faulty")

// FaultyEngine always returns an ErrFaultyEngine error.
type FaultyEngine struct{}

// VerifyHeader records that header was verified.
func (e *FaultyEngine) VerifyHeader(chain chain.Reader, header *pb.Header) error {
	return ErrFaultyEngine
}

// Prepare does nothing.
func (e *FaultyEngine) Prepare(chain chain.Reader, header *pb.Header) error {
	return ErrFaultyEngine
}

// Finalize does nothing.
func (e *FaultyEngine) Finalize(ctx context.Context, chain chain.Reader, header *pb.Header, state state.Reader, txs []*pb.Transaction) (*pb.Block, error) {
	return nil, ErrFaultyEngine
}

// DummyProofOfWait is a dummy proof-of-wait © engine.
type DummyProofOfWait struct {
	InstrumentedEngine

	intervalCount uint32

	min int
	max int
}

// NewDummyProofOfWait creates a new DummyProofOfWait.
func NewDummyProofOfWait(e engine.Engine, min, max int) *DummyProofOfWait {
	powEngine := &DummyProofOfWait{
		min: min,
		max: max,
	}

	powEngine.InstrumentedEngine = InstrumentedEngine{engine: e}

	return powEngine
}

// Interval increments a counter and returns a configurable interval for the proof-of-wait.
func (e *DummyProofOfWait) Interval() (int, int) {
	atomic.AddUint32(&e.intervalCount, 1)
	return e.min, e.max
}

// IntervalCount returns the number of calls to Interval.
func (e *DummyProofOfWait) IntervalCount() uint32 {
	return e.intervalCount
}

// DummyPoW is a dummy proof-of-work engine.
type DummyPoW struct {
	InstrumentedEngine

	difficulty      uint64
	difficultyCount uint32
}

// NewDummyPoW creates a new DummyPoW.
func NewDummyPoW(e engine.Engine, difficulty uint64) *DummyPoW {
	powEngine := &DummyPoW{difficulty: difficulty}
	powEngine.InstrumentedEngine = InstrumentedEngine{engine: e}
	return powEngine
}

// Difficulty returns a configurable difficulty.
func (e *DummyPoW) Difficulty() uint64 {
	atomic.AddUint32(&e.difficultyCount, 1)
	return e.difficulty
}

// DifficultyCount returns the number of calls to Difficulty.
func (e *DummyPoW) DifficultyCount() uint32 {
	return e.difficultyCount
}
