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

package state

import (
	"encoding/binary"

	"github.com/pkg/errors"
	db "github.com/stratumn/alice/core/protocol/coin/db"
)

var (
	// ErrAmountTooBig is returned when subtracting an amount greater than
	// the current balance.
	ErrAmountTooBig = errors.New("amount is too big")

	// ErrInvalidBatch is returned when a batch is invalid.
	ErrInvalidBatch = errors.New("batch is invalid")
)

// State stores users' account balances.
type State interface {
	Reader
	Writer

	// AddBalance atomically adds coins to an account. It returns the new
	// balance.
	AddBalance(pubKey []byte, amount uint64) (uint64, error)

	// SubBalance atomically removes coins from an account. It returns the
	// new balance. It returns an error if the amount is greater than the
	// current balance.
	SubBalance(pubKey []byte, amount uint64) (uint64, error)
}

// Reader gives read access to users' account balances.
type Reader interface {
	// GetBalance gets the account balance of a user identified
	// by his public key. It returns 0 if the account is not found.
	GetBalance(pubKey []byte) (uint64, error)
}

// Writer gives write access to users' account balances.
type Writer interface {
	// SetBalance sets the account balance of a user identified by his
	// public key.
	SetBalance(pubKey []byte, value uint64) error

	// Batch creates a batch which can be used to execute multiple write
	// operations efficiently.
	Batch() Batch

	// Write executes all the operations in a batch. It may write
	// concurrently.
	Write(Batch) error
}

// Batch can execute multiple write operations efficiently. It is only intended
// to be used once. When writing the batch to the database, the operations may
// execute concurrently, so order is not guaranteed.
type Batch interface {
	// SetBalance sets the account balance of a user identified by his
	// public key.
	SetBalance(pubKey []byte, value uint64)
}

// stateDB implements the State interface using a key-value database.
type stateDB struct {
	db     db.DB
	prefix []byte
}

// NewState creates a new state from a DB instance.
//
// Prefix is used to prefix keys in the database.
func NewState(db db.DB, prefix []byte) State {
	return &stateDB{db: db, prefix: prefix}
}

func (s *stateDB) GetBalance(pubKey []byte) (uint64, error) {
	v, err := s.db.Get(prefixKey(s.prefix, pubKey))
	if errors.Cause(err) == db.ErrNotFound {
		return 0, nil
	}

	if err != nil {
		return 0, err
	}

	return binary.LittleEndian.Uint64(v), nil
}

func (s *stateDB) SetBalance(pubKey []byte, value uint64) error {
	v := make([]byte, 8)
	binary.LittleEndian.PutUint64(v, value)

	return s.db.Put(prefixKey(s.prefix, pubKey), v)
}

func (s *stateDB) AddBalance(pubKey []byte, amount uint64) (uint64, error) {
	return s.incBalance(pubKey, amount, true)
}

func (s *stateDB) SubBalance(pubKey []byte, amount uint64) (uint64, error) {
	return s.incBalance(pubKey, amount, false)
}

// incBalance adds or subtracts coins from an account.
func (s *stateDB) incBalance(pubKey []byte, amount uint64, add bool) (uint64, error) {
	// Use a transaction to make it atomic.
	tx, err := s.db.Transaction()
	if err != nil {
		return 0, err
	}

	// Get current value.
	current := uint64(0)

	c, err := tx.Get(prefixKey(s.prefix, pubKey))
	if err != nil {
		if errors.Cause(err) != db.ErrNotFound {
			tx.Discard()
			return 0, err
		}
	} else {
		current = binary.LittleEndian.Uint64(c)
	}

	// Update value.
	if add {
		amount += current
	} else {
		if amount > current {
			tx.Discard()
			return 0, ErrAmountTooBig
		}

		amount = current - amount
	}

	v := make([]byte, 8)
	binary.LittleEndian.PutUint64(v, amount)

	if err := tx.Put(prefixKey(s.prefix, pubKey), v); err != nil {
		tx.Discard()
		return 0, err
	}

	return amount, tx.Commit()
}

func (s *stateDB) Batch() Batch {
	return &dbBatch{s.db.Batch(), s.prefix}
}

func (s *stateDB) Write(batch Batch) error {
	b, ok := batch.(*dbBatch)
	if !ok {
		return ErrInvalidBatch
	}

	return s.db.Write(b.batch)
}

// dbBatch implements a batch for stateDB.
type dbBatch struct {
	batch  db.Batch
	prefix []byte
}

func (b dbBatch) SetBalance(pubKey []byte, value uint64) {
	v := make([]byte, 8)
	binary.LittleEndian.PutUint64(v, value)
	b.batch.Put(prefixKey(b.prefix, pubKey), v)
}

// prefixKey prefixes the given key.
func prefixKey(prefix, key []byte) []byte {
	k := make([]byte, len(prefix)+len(key))
	copy(k, prefix)
	copy(k[len(prefix):], key)

	return k
}
