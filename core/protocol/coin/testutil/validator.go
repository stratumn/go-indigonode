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

	"github.com/stratumn/alice/core/protocol/coin/state"
	pb "github.com/stratumn/alice/pb/coin"
)

var (
	// ErrRejected is the error returns by Rejector.
	ErrRejected = errors.New("Rejected")
)

// Rejector implements the Validator interface and rejects
// incoming transactions and blocks.
// It records incoming transactions and blocks for assertions.
type Rejector struct {
	mu     sync.RWMutex
	txs    []*pb.Transaction
	blocks []*pb.Block
}

// ValidateTx records the incoming tx and rejects it.
func (r *Rejector) ValidateTx(tx *pb.Transaction, state state.Reader) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.txs = append(r.txs, tx)
	return ErrRejected
}

// ValidateBlock records the incoming block and rejects it.
func (r *Rejector) ValidateBlock(block *pb.Block, _ state.Reader) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.blocks = append(r.blocks, block)
	return ErrRejected
}

// ValidatedTx returns true if the rejector saw the given transaction.
func (r *Rejector) ValidatedTx(tx *pb.Transaction) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	matcher := NewTxMatcher(tx)
	for _, txx := range r.txs {
		if matcher.Matches(txx) {
			return true
		}
	}

	return false
}

// ValidatedBlock returns true if the rejector saw the given block.
func (r *Rejector) ValidatedBlock(block *pb.Block) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	matcher := NewHeaderMatcher(block.Header)
	for _, b := range r.blocks {
		if matcher.Matches(b.Header) {
			return true
		}
	}

	return false
}
