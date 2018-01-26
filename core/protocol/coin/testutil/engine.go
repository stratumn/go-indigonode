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

// VerifyHeaders records that headers were verified.
func (e *DummyEngine) VerifyHeaders(ctx context.Context, chain chain.Reader, headers []*pb.Header) <-chan error {
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
func (e *DummyEngine) Finalize(chain chain.Reader, header *pb.Header, state *state.State, txs []*pb.Transaction) (*pb.Block, error) {
	return nil, nil
}
