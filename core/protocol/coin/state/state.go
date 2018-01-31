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

	// ErrInvalidAccount is returned when the binary representation of an
	// account is invalid.
	ErrInvalidAccount = errors.New("account is invalid")

	// ErrInvalidBatch is returned when a batch is invalid.
	ErrInvalidBatch = errors.New("batch is invalid")
)

// Account describes a user account.
type Account struct {
	// Balance is the number of coins the user has.
	Balance uint64
	// Nonce is the latest transaction nonce seen by the system.
	// It can only increase.
	Nonce uint64
}

// DecodeAccount decodes an account from its binary representation.
func DecodeAccount(buf []byte) (Account, error) {
	if len(buf) != 16 {
		return Account{}, ErrInvalidAccount
	}

	return Account{
		Balance: binary.LittleEndian.Uint64(buf),
		Nonce:   binary.LittleEndian.Uint64(buf[8:]),
	}, nil
}

// Encode encodes an account to its binary representation.
func (a Account) Encode() []byte {
	v := make([]byte, 16)
	binary.LittleEndian.PutUint64(v, a.Balance)
	binary.LittleEndian.PutUint64(v[8:], a.Nonce)

	return v
}

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
	// GetAccount gets the account details of a user identified
	// by his public key. It returns Account{} if the account is not found.
	GetAccount(pubKey []byte) (Account, error)
}

// Writer gives write access to users' account balances.
type Writer interface {
	// UpdateAccount sets or updates the account of a user identified by
	// his public key.
	UpdateAccount(pubKey []byte, value Account) error

	// Transfer atomically transfers coins from an account to another. Only
	// the nonce from the sender's account is updated. It returns an error
	// if the amount transfered is greater than the balance the coins are
	// sent from.
	Transfer(fromPubKey, toPubKey []byte, amount, nonce uint64) error

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
	// UpdateAccount sets or updates the account of a user identified by
	// his public key.
	UpdateAccount(pubKey []byte, value Account)
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

func (s *stateDB) GetAccount(pubKey []byte) (Account, error) {
	v, err := s.db.Get(append(s.prefix, pubKey...))
	if errors.Cause(err) == db.ErrNotFound {
		return Account{}, nil
	}

	if err != nil {
		return Account{}, err
	}

	return DecodeAccount(v)
}

func (s *stateDB) UpdateAccount(pubKey []byte, value Account) error {
	return s.db.Put(append(s.prefix, pubKey...), value.Encode())
}

func (s *stateDB) Transfer(fromPubKey, toPubKey []byte, amount, nonce uint64) error {
	tx, err := s.db.Transaction()
	if err != nil {
		return err
	}

	err = s.incBalance(tx, fromPubKey, false, true, amount, nonce)
	if err != nil {
		tx.Discard()
		return err
	}

	err = s.incBalance(tx, toPubKey, true, false, amount, 0)
	if err != nil {
		tx.Discard()
		return err
	}

	return tx.Commit()
}

// incBalance adds or subtracts coins from an account in a transaction. It
// doesn't close the transaction.
func (s *stateDB) incBalance(
	tx db.Transaction,
	pubKey []byte,
	add, updateNonce bool,
	amount, nonce uint64,
) error {
	// Get current value.
	var a Account

	buf, err := tx.Get(append(s.prefix, pubKey...))
	if err != nil {
		if errors.Cause(err) != db.ErrNotFound {
			return err
		}
	} else {
		if a, err = DecodeAccount(buf); err != nil {
			return err
		}
	}

	// Update value.
	if add {
		a.Balance += amount
	} else {
		if amount > a.Balance {
			return ErrAmountTooBig
		}

		a.Balance -= amount
	}

	if updateNonce {
		a.Nonce = nonce
	}

	if err := tx.Put(append(s.prefix, pubKey...), a.Encode()); err != nil {
		return err
	}

	return nil
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

func (b dbBatch) UpdateAccount(pubKey []byte, value Account) {
	b.batch.Put(append(b.prefix, pubKey...), value.Encode())
}

// dbTx implements a transaction for stateDB.
type dbTx struct {
	dbtx   db.Transaction
	prefix []byte
}

func (tx dbTx) GetAccount(pubKey []byte) (Account, error) {
	v, err := tx.dbtx.Get(append(tx.prefix, pubKey...))
	if errors.Cause(err) == db.ErrNotFound {
		return Account{}, nil
	}

	if err != nil {
		return Account{}, err
	}

	return DecodeAccount(v)
}

func (tx dbTx) UpdateAccount(pubKey []byte, value Account) error {
	return tx.dbtx.Put(append(tx.prefix, pubKey...), value.Encode())
}

func (tx dbTx) Transfer(fromPubKey, toPubKey []byte, amount, nonce uint64) error {
	if err := tx.incBalance(fromPubKey, false, true, amount, nonce); err != nil {
		return err
	}

	if err := tx.incBalance(toPubKey, true, false, amount, 0); err != nil {
		return err
	}

	return nil
}

// incBalance adds or subtracts coins from an account in a transaction.
func (tx dbTx) incBalance(
	pubKey []byte,
	add, updateNonce bool,
	amount, nonce uint64,
) error {
	// Get current value.
	a, err := tx.GetAccount(pubKey)
	if err != nil {
		return err
	}

	// Update value.
	if add {
		a.Balance += amount
	} else {
		if amount > a.Balance {
			return ErrAmountTooBig
		}

		a.Balance -= amount
	}

	if updateNonce {
		a.Nonce = nonce
	}

	return tx.UpdateAccount(pubKey, a)
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
