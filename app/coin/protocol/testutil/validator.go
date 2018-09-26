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
	"errors"
	"sync"

	"github.com/stratumn/go-node/app/coin/pb"
	"github.com/stratumn/go-node/app/coin/protocol/state"
	"github.com/stratumn/go-node/app/coin/protocol/validator"
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

// MaxTxPerBlock returns the maximum number of transactions
// allowed in a block.
func (r *InstrumentedValidator) MaxTxPerBlock() uint32 {
	return r.validator.MaxTxPerBlock()
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

// MaxTxPerBlock always returns 1.
func (r *Rejector) MaxTxPerBlock() uint32 {
	return 1
}

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

// MaxTxPerBlock always returns 1.
func (v *DummyValidator) MaxTxPerBlock() uint32 {
	return 1
}

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
