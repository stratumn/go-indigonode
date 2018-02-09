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
	"bytes"
	"encoding/binary"
	"sync"

	"github.com/pkg/errors"
	"github.com/stratumn/alice/core/protocol/coin/coinutil"
	db "github.com/stratumn/alice/core/protocol/coin/db"
	pb "github.com/stratumn/alice/pb/coin"
)

var (
	// ErrInconsistentTransactions is returned when the transactions to
	// roll back are inconsistent with the saved nonces.
	ErrInconsistentTransactions = errors.New("transactions are inconsistent")
)

// State stores users' account balances. It doesn't handle validation.
type State interface {
	Reader
	Writer
}

// Reader gives read access to users' account balances.
type Reader interface {
	// GetAccount gets the account details of a user identified
	// by his public key. It returns &pb.Account{} if the account is not
	// found.
	GetAccount(pubKey []byte) (*pb.Account, error)
}

// Writer gives write access to users' account balances.
type Writer interface {
	// UpdateAccount sets or updates the account of a user identified by
	// his public key. It should only be used for testing as it cannot be
	// rolled back.
	UpdateAccount(pubKey []byte, account *pb.Account) error

	// ProcessTransactions processes all the given transactions and updates
	// the state accordingly. It should be given a unique state ID, for
	// instance the hash of the block containing the transactions.
	ProcessTransactions(stateID []byte, txs []*pb.Transaction) error

	// RollbackTransactions rolls back transactions. The parameters are
	// expected to be identical to the ones that were given to the
	// corresponding call to ProcessTransactions().
	// You should only rollback the current state to the previous one.
	RollbackTransactions(stateID []byte, txs []*pb.Transaction) error
}

const (
	// accountPrefix is the prefix for account keys.
	accountPrefix byte = iota

	// prevNoncesPrefix is the prefix for previous nonces keys.
	prevNoncesPrefix = iota + 1
)

type stateDB struct {
	mu     sync.RWMutex
	db     db.DB
	diff   *db.Diff
	prefix []byte
}

// Opt is an option for state.
type Opt func(*stateDB)

// OptPrefix sets a prefix for all the database keys.
var OptPrefix = func(prefix []byte) Opt {
	return func(s *stateDB) {
		s.prefix = prefix
	}
}

// NewState creates a new state from a DB instance.
//
// Prefix is used to prefix keys in the database.
//
// VERY IMPORTANT NOTE: ALL THE DIFFERENT MODULES THAT SHARE THE SAME INSTANCE
// OF THE DATABASE MUST USE A UNIQUE PREFIX OF THE SAME BYTESIZE.
func NewState(database db.DB, opts ...Opt) State {
	s := &stateDB{
		db:   database,
		diff: db.NewDiff(database),
	}

	for _, o := range opts {
		o(s)
	}

	return s
}

func (s *stateDB) GetAccount(pubKey []byte) (*pb.Account, error) {
	return s.doGetAccount(s.db, pubKey)
}

// doGetAccount is able to get an account using anything that implements the
// db.Reader interface.
func (s *stateDB) doGetAccount(dbr db.Reader, pubKey []byte) (*pb.Account, error) {
	buf, err := dbr.Get(s.accountKey(pubKey))
	if errors.Cause(err) == db.ErrNotFound {
		return &pb.Account{}, nil
	}
	if err != nil {
		return nil, err
	}

	var account pb.Account
	err = account.Unmarshal(buf)

	return &account, errors.WithStack(err)
}

func (s *stateDB) UpdateAccount(pubKey []byte, account *pb.Account) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.doUpdateAccount(s.db, pubKey, account)
}

// doGetAccount updates an account using anything that implements db.Writer.
func (s *stateDB) doUpdateAccount(dbw db.Writer, pubKey []byte, account *pb.Account) error {
	// Delete empty accounts to save space.
	if account.Balance == 0 && account.Nonce == 0 {
		return dbw.Delete(s.accountKey(pubKey))
	}

	buf, err := account.Marshal()
	if err != nil {
		return errors.WithStack(err)
	}

	return dbw.Put(s.accountKey(pubKey), buf)
}

func (s *stateDB) ProcessTransactions(stateID []byte, txs []*pb.Transaction) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.diff.Reset()

	// Eight bytes per transaction.
	nonces := make([]byte, len(txs)*8)

	for i, tx := range txs {
		txHash, err := coinutil.HashTransaction(tx)
		if err != nil {
			return err
		}
		// Substract amount for sender, except for coinbase transaction.
		if tx.From != nil {
			from, err := s.doGetAccount(s.diff, tx.From)
			if err != nil {
				return err
			}

			binary.LittleEndian.PutUint64(nonces[i*8:], from.Nonce)

			// Subtract amount from sender and update nonce.
			from.Balance -= tx.Value
			from.Nonce = tx.Nonce
			from.TransactionHashes = append(from.TransactionHashes, txHash)

			err = s.doUpdateAccount(s.diff, tx.From, from)
			if err != nil {
				return err
			}
		}

		// Add amount to receiver.
		to, err := s.doGetAccount(s.diff, tx.To)
		if err != nil {
			return err
		}

		to.Balance += tx.Value
		to.TransactionHashes = append(to.TransactionHashes, txHash)

		err = s.doUpdateAccount(s.diff, tx.To, to)
		if err != nil {
			return err
		}

	}

	if err := s.diff.Put(s.prevNoncesKey(stateID), nonces); err != nil {
		return err
	}

	return s.diff.Apply()
}

func (s *stateDB) RollbackTransactions(stateID []byte, txs []*pb.Transaction) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	defer s.diff.Reset()

	nonces, err := s.db.Get(s.prevNoncesKey(stateID))
	if err != nil {
		return err
	}

	if len(nonces) != len(txs)*8 {
		return ErrInconsistentTransactions
	}

	for i := len(txs) - 1; i >= 0; i-- {
		tx := txs[i]
		txHash, err := coinutil.HashTransaction(tx)
		if err != nil {
			return err
		}

		// Add amount to sender and restore nonce.
		from, err := s.doGetAccount(s.diff, tx.From)
		if err != nil {
			return nil
		}

		from.Balance += tx.Value
		from.Nonce = binary.LittleEndian.Uint64(nonces[i*8:])
		from.TransactionHashes = removeTxHash(from.TransactionHashes, txHash)

		err = s.doUpdateAccount(s.diff, tx.From, from)
		if err != nil {
			return err
		}

		// Subtract amount from received.
		to, err := s.doGetAccount(s.diff, tx.To)
		if err != nil {
			return nil
		}

		to.Balance -= tx.Value
		to.TransactionHashes = removeTxHash(to.TransactionHashes, txHash)

		err = s.doUpdateAccount(s.diff, tx.To, to)
		if err != nil {
			return err
		}
	}

	if err := s.diff.Delete(s.prevNoncesKey(stateID)); err != nil {
		return err
	}

	return s.diff.Apply()
}

func removeTxHash(history [][]byte, txHash []byte) [][]byte {
	var index int
	for i, h := range history {
		if bytes.Compare(h, txHash) == 0 {
			index = i
			break
		}
	}
	return append(history[:index], history[index+1:]...)
}

// accountKey returns the key corresponding to an account given its public key.
func (s *stateDB) accountKey(pubKey []byte) []byte {
	return append(append(s.prefix, accountPrefix), pubKey...)
}

// prevNoncesKey returns the previous nonces key for the given state ID.
func (s *stateDB) prevNoncesKey(stateID []byte) []byte {
	return append(append(s.prefix, prevNoncesPrefix), stateID...)
}
