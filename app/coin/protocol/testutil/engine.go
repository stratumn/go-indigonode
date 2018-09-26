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
	"sync"
	"sync/atomic"

	"github.com/stratumn/go-node/app/coin/pb"
	"github.com/stratumn/go-node/app/coin/protocol/chain"
	"github.com/stratumn/go-node/app/coin/protocol/engine"
	"github.com/stratumn/go-node/app/coin/protocol/state"
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

	reward uint64
}

// NewDummyPoW creates a new DummyPoW.
func NewDummyPoW(e engine.Engine, difficulty, reward uint64) *DummyPoW {
	powEngine := &DummyPoW{difficulty: difficulty, reward: reward}
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

// Reward returns the statically configured miner reward.
func (e *DummyPoW) Reward() uint64 {
	return e.reward
}
