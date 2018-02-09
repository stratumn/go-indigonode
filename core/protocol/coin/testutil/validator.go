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
	"github.com/stratumn/alice/core/protocol/coin/validator"
	pb "github.com/stratumn/alice/pb/coin"
)

// InstrumentedValidator adds method calls instrumentation to a validator.
type InstrumentedValidator struct {
	validator validator.Validator

	mu     sync.RWMutex
	txs    []*pb.Transaction
	blocks []*pb.Block
}

// NewInstrumentedValidator wraps a validator with instrumentation.
func NewInstrumentedValidator(validator validator.Validator) *InstrumentedValidator {
	return &InstrumentedValidator{validator: validator}
}

// ValidateTx records the incoming tx.
func (r *InstrumentedValidator) ValidateTx(tx *pb.Transaction, state state.Reader) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.txs = append(r.txs, tx)
	return r.validator.ValidateTx(tx, state)
}

// ValidateBlock records the incoming block.
func (r *InstrumentedValidator) ValidateBlock(block *pb.Block, state state.Reader) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.blocks = append(r.blocks, block)
	return r.validator.ValidateBlock(block, state)
}

// ValidateTransactions records the incoming block.
func (r *InstrumentedValidator) ValidateTransactions(transactions []*pb.Transaction, state state.Reader) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.txs = append(r.txs, transactions...)
	return r.validator.ValidateTransactions(transactions, state)
}

// ValidatedTx returns true if the validator saw the given transaction.
func (r *InstrumentedValidator) ValidatedTx(tx *pb.Transaction) bool {
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

// ValidatedBlock returns true if the validator saw the given block.
func (r *InstrumentedValidator) ValidatedBlock(block *pb.Block) bool {
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

var (
	// ErrRejected is the error returns by Rejector.
	ErrRejected = errors.New("rejected")
)

// Rejector implements the Validator interface and rejects
// incoming transactions and blocks.
type Rejector struct{}

// ValidateTx records the incoming tx and rejects it.
func (r *Rejector) ValidateTx(tx *pb.Transaction, state state.Reader) error {
	return ErrRejected
}

// ValidateBlock records the incoming block and rejects it.
func (r *Rejector) ValidateBlock(block *pb.Block, _ state.Reader) error {
	return ErrRejected
}

// ValidateTransactions records the incoming block and rejects it.
func (r *Rejector) ValidateTransactions(transactions []*pb.Transaction, _ state.Reader) error {
	return ErrRejected
}

// DummyValidator is a validator that always returns nil (valid).
type DummyValidator struct{}

// ValidateTx always returns nil (valid tx).
func (v *DummyValidator) ValidateTx(tx *pb.Transaction, state state.Reader) error {
	return nil
}

// ValidateBlock always returns nil (valid block).
func (v *DummyValidator) ValidateBlock(block *pb.Block, _ state.Reader) error {
	return nil
}

// ValidateTransactions always returns nil (valid block).
func (v *DummyValidator) ValidateTransactions(transactions []*pb.Transaction, _ state.Reader) error {
	return nil
}
