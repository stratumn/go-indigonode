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

	// Transaction creates a transaction which can be used to execute
	// multiple operations atomically.
	Transaction() (Transaction, error)
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

	// AddBalance atomically adds coins to an account. It returns the new
	// balance.
	AddBalance(pubKey []byte, amount uint64) (uint64, error)

	// SubBalance atomically removes coins from an account. It returns the
	// new balance. It returns an error if the amount is greater than the
	// current balance.
	SubBalance(pubKey []byte, amount uint64) (uint64, error)

	// Transfer atomically transfers coins from an account to another. It
	// returns an error if the amount transfered is greater than the
	// balance the coins are sent from.
	Transfer(fromPubKey, toPubKey []byte, amount uint64) error

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

// Transaction can be used to execute multiple operations atomically. It should
// be closed exactly once by calling either Commit or Discard.
type Transaction interface {
	Reader
	Writer

	// Commit should be called to commit the transaction to the database.
	Commit() error

	// Discard should be called to discard the transaction.
	Discard()
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
	return s.db.Put(prefixKey(s.prefix, pubKey), encodeUint64(value))
}

func (s *stateDB) AddBalance(pubKey []byte, amount uint64) (uint64, error) {
	tx, err := s.db.Transaction()
	if err != nil {
		return 0, err
	}

	amount, err = s.incBalance(tx, pubKey, amount, true)
	if err != nil {
		tx.Discard()
		return 0, err
	}

	return amount, tx.Commit()
}

func (s *stateDB) SubBalance(pubKey []byte, amount uint64) (uint64, error) {
	tx, err := s.db.Transaction()
	if err != nil {
		return 0, err
	}

	amount, err = s.incBalance(tx, pubKey, amount, false)
	if err != nil {
		tx.Discard()
		return 0, err
	}

	return amount, tx.Commit()
}

func (s *stateDB) Transfer(fromPubKey, toPubKey []byte, amount uint64) error {
	tx, err := s.db.Transaction()
	if err != nil {
		return err
	}

	_, err = s.incBalance(tx, fromPubKey, amount, false)
	if err != nil {
		tx.Discard()
		return err
	}

	_, err = s.incBalance(tx, toPubKey, amount, true)
	if err != nil {
		tx.Discard()
		return err
	}

	return tx.Commit()
}

// incBalance adds or subtracts coins from an account in a transaction. It
// doesn't close the transaction.
func (s *stateDB) incBalance(tx db.Transaction, pubKey []byte, amount uint64, add bool) (uint64, error) {
	// Get current value.
	current := uint64(0)

	c, err := tx.Get(prefixKey(s.prefix, pubKey))
	if err != nil {
		if errors.Cause(err) != db.ErrNotFound {
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
			return 0, ErrAmountTooBig
		}

		amount = current - amount
	}

	if err := tx.Put(prefixKey(s.prefix, pubKey), encodeUint64(amount)); err != nil {
		return 0, err
	}

	return amount, nil
}

func (s *stateDB) Batch() Batch {
	return dbBatch{s.db.Batch(), s.prefix}
}

func (s *stateDB) Transaction() (Transaction, error) {
	tx, err := s.db.Transaction()
	if err != nil {
		return nil, err
	}

	return dbTx{dbtx: tx, prefix: s.prefix}, nil
}

func (s *stateDB) Write(batch Batch) error {
	b, ok := batch.(dbBatch)
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
	b.batch.Put(prefixKey(b.prefix, pubKey), encodeUint64(value))
}

// dbTx implements a transaction for stateDB.
type dbTx struct {
	dbtx   db.Transaction
	prefix []byte
}

func (tx dbTx) GetBalance(pubKey []byte) (uint64, error) {
	v, err := tx.dbtx.Get(prefixKey(tx.prefix, pubKey))
	if errors.Cause(err) == db.ErrNotFound {
		return 0, nil
	}

	if err != nil {
		return 0, err
	}

	return binary.LittleEndian.Uint64(v), nil
}

func (tx dbTx) SetBalance(pubKey []byte, value uint64) error {
	return tx.dbtx.Put(prefixKey(tx.prefix, pubKey), encodeUint64(value))
}

func (tx dbTx) AddBalance(pubKey []byte, amount uint64) (uint64, error) {
	current, err := tx.GetBalance(pubKey)
	if err != nil {
		return 0, err
	}

	amount += current

	return amount, tx.SetBalance(pubKey, amount)
}

func (tx dbTx) SubBalance(pubKey []byte, amount uint64) (uint64, error) {
	current, err := tx.GetBalance(pubKey)
	if err != nil {
		return 0, err
	}

	if amount > current {
		return 0, ErrAmountTooBig
	}

	amount = current - amount

	return amount, tx.SetBalance(pubKey, amount)
}

func (tx dbTx) Transfer(fromPubKey, toPubKey []byte, amount uint64) error {
	if _, err := tx.SubBalance(fromPubKey, amount); err != nil {
		return err
	}

	if _, err := tx.AddBalance(toPubKey, amount); err != nil {
		return err
	}

	return nil
}

func (tx dbTx) Batch() Batch {
	return dbBatch{batch: tx.dbtx.Batch(), prefix: tx.prefix}
}

func (tx dbTx) Write(batch Batch) error {
	b, ok := batch.(dbBatch)
	if !ok {
		return ErrInvalidBatch
	}

	return tx.dbtx.Write(b.batch)
}

func (tx dbTx) Commit() error {
	return tx.dbtx.Commit()
}

func (tx dbTx) Discard() {
	tx.dbtx.Discard()
}

// encodeUint64 encodes an uint64 to a buffer.
func encodeUint64(value uint64) []byte {
	v := make([]byte, 8)
	binary.LittleEndian.PutUint64(v, value)

	return v
}

// prefixKey prefixes the given key.
func prefixKey(prefix, key []byte) []byte {
	k := make([]byte, len(prefix)+len(key))
	copy(k, prefix)
	copy(k[len(prefix):], key)

	return k
}
