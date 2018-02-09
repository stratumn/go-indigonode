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
	"encoding/json"
	"sync"

	"github.com/pkg/errors"
	"github.com/stratumn/alice/core/protocol/coin/coinutil"
	"github.com/stratumn/alice/core/protocol/coin/db"
	"github.com/stratumn/alice/core/protocol/coin/trie"
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
	TxReader
	Writer
}

// Reader gives read access to users' account balances.
type Reader interface {
	// GetAccount gets the account details of a user identified
	// by his public key. It returns &pb.Account{} if the account is not
	// found.
	GetAccount(pubKey []byte) (*pb.Account, error)
}

// TxReader gives read access to a user's transactions.
type TxReader interface {
	// GetAccountTxHashes gets the transaction history of a user identified
	// by his public key.
	GetAccountTxKeys(pubKey []byte) ([]*TxKey, error)
}

// Writer gives write access to users' account balances.
type Writer interface {
	// UpdateAccount sets or updates the account of a user identified by
	// his public key. It should only be used for testing as it cannot be
	// rolled back.
	UpdateAccount(pubKey []byte, account *pb.Account) error

	// ProcessTransactions processes all the transactions of the given
	// block and updates the state accordingly. It should be given a unique
	// state ID, for instance the hash of the block containing the
	// transactions.
	ProcessTransactions(stateID []byte, blk *pb.Block) error

	// RollbackTransactions rolls back transactions. The parameters are
	// expected to be identical to the ones that were given to the
	// corresponding call to ProcessTransactions().
	// You should only rollback the current state to the previous one.
	RollbackTransactions(stateID []byte, blk *pb.Block) error
}

const (
	// accountPrefix is the prefix for account keys.
	accountPrefix byte = iota

	// prevNoncesPrefix is the prefix for previous nonces keys.
	prevNoncesPrefix = iota + 1
)

// TxKey is used to save a user transactions.
type TxKey struct {
	TxIdx   int
	BlkHash []byte
}

type stateDB struct {
	// The general pattern for updating accounts is:
	//
	//	1. Modify the accountsTrie
	//	2. Apply accountsTrie changes to diff
	//	3. Do other modifications to diff (like saving nonces)
	//	4. Atomically apply changes from diff to underlying DB
	//	5. Reset accountsTrie and diff changes in case an error occured
	mu           sync.RWMutex
	diff         *db.Diff
	accountsTrie *trie.Trie
	prefix       []byte
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
func NewState(database db.ReadWriteBatcher, opts ...Opt) State {
	diff := db.NewDiff(database)

	s := &stateDB{diff: db.NewDiff(database)}

	for _, o := range opts {
		o(s)
	}

	s.accountsTrie = trie.New(
		trie.OptDB(diff),
		trie.OptPrefix(append(s.prefix, accountPrefix)),
	)

	return s
}

func (s *stateDB) GetAccount(pubKey []byte) (*pb.Account, error) {
	return s.doGetAccount(pubKey)
}

// doGetAccount returns an account.
func (s *stateDB) doGetAccount(pubKey []byte) (*pb.Account, error) {
	buf, err := s.accountsTrie.Get(pubKey)
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

func (s *stateDB) GetAccountTxKeys(pubKey []byte) ([]*TxKey, error) {
	var txKeys []*TxKey
	iter := s.accountsTrie.IteratePrefix(pubKey)
	defer iter.Release()

	for {
		next, err := iter.Next()
		if err != nil {
			return nil, err
		}
		if !next {
			break
		}
		if bytes.Compare(iter.Key(), pubKey) != 0 {
			txK := &TxKey{}
			err := json.Unmarshal(iter.Value(), txK)
			if err != nil {
				return nil, err
			}

			txKeys = append(txKeys, txK)
		}
	}

	return txKeys, nil
}

func (s *stateDB) UpdateAccount(pubKey []byte, account *pb.Account) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	defer s.diff.Reset()
	defer s.accountsTrie.Reset()

	err := s.doUpdateAccount(pubKey, account)
	if err != nil {
		s.accountsTrie.Reset()
	}

	// Commit trie changes to diff.
	if err := s.accountsTrie.Commit(); err != nil {
		return err
	}

	// Commit diff changes to database.
	return s.diff.Apply()
}

// doGetAccount updates an account.
func (s *stateDB) doUpdateAccount(pubKey []byte, account *pb.Account) error {
	// Delete empty accounts to save space.
	if account.Balance == 0 && account.Nonce == 0 {
		return s.accountsTrie.Delete(pubKey)
	}

	buf, err := account.Marshal()
	if err != nil {
		return errors.WithStack(err)
	}

	return s.accountsTrie.Put(pubKey, buf)
}

func (s *stateDB) ProcessTransactions(stateID []byte, blk *pb.Block) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	defer s.diff.Reset()
	defer s.accountsTrie.Reset()

	txs := blk.GetTransactions()

	// Eight bytes per transaction.
	nonces := make([]byte, len(txs)*8)

	for i, tx := range txs {
		// Substract amount for sender, except for reward transaction.
		if tx.From != nil {
			from, err := s.doGetAccount(tx.From)
			if err != nil {
				return err
			}

			// Save current nonce.
			binary.LittleEndian.PutUint64(nonces[i*8:], from.Nonce)

			// Subtract amount from sender.
			from.Balance -= tx.Value + tx.Fee

			// Update nonce. Since transactions are batches, nonces might be
			// out-of-order in the batch. If the batch contains Nonces [3,4,1]
			// then the final nonce should be 4.
			if from.Nonce < tx.Nonce {
				from.Nonce = tx.Nonce
			}

			err = s.doUpdateAccount(tx.From, from)
			if err != nil {
				return err
			}

			err = s.addTxKey(tx.From, i, blk)
			if err != nil {
				return err
			}
		}

		// Add amount to receiver.
		to, err := s.doGetAccount(tx.To)
		if err != nil {
			return err
		}

		to.Balance += tx.Value

		err = s.doUpdateAccount(tx.To, to)
		if err != nil {
			return err
		}

		err = s.addTxKey(tx.To, i, blk)
		if err != nil {
			return err
		}
	}

	// We do not put nonces in the trie since they are implementation
	// specific and should not affect the Merkle Root.
	if err := s.diff.Put(s.prevNoncesKey(stateID), nonces); err != nil {
		return err
	}

	// Commit trie changes to diff.
	if err := s.accountsTrie.Commit(); err != nil {
		return err
	}

	// Commit diff changes to database.
	return s.diff.Apply()
}

func (s *stateDB) RollbackTransactions(stateID []byte, blk *pb.Block) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	defer s.diff.Reset()
	defer s.accountsTrie.Reset()

	// Nonces are in the database, not the trie.
	nonces, err := s.diff.Get(s.prevNoncesKey(stateID))
	if err != nil {
		return err
	}

	txs := blk.GetTransactions()

	if len(nonces) != len(txs)*8 {
		return ErrInconsistentTransactions
	}

	for i := len(txs) - 1; i >= 0; i-- {
		tx := txs[i]
		if err != nil {
			return err
		}

		// Add amount to sender, except for reward transaction.
		if tx.From != nil {
			from, err := s.doGetAccount(tx.From)
			if err != nil {
				return err
			}

			from.Balance += tx.Value + tx.Fee

			// Restore nonce.
			from.Nonce = binary.LittleEndian.Uint64(nonces[i*8:])

			err = s.doUpdateAccount(tx.From, from)
			if err != nil {
				return err
			}

			err = s.removeTxKey(tx.From, i, blk)
			if err != nil {
				return err
			}
		}

		// Subtract amount from received.
		to, err := s.doGetAccount(tx.To)
		if err != nil {
			return err
		}

		to.Balance -= tx.Value

		err = s.doUpdateAccount(tx.To, to)
		if err != nil {
			return err
		}

		err = s.removeTxKey(tx.To, i, blk)
		if err != nil {
			return err
		}
	}

	// Don't need nonces anymore.
	if err := s.diff.Delete(s.prevNoncesKey(stateID)); err != nil {
		return err
	}

	// Commit trie changes to diff.
	if err := s.accountsTrie.Commit(); err != nil {
		return err
	}

	// Commit diff changes to database.
	return s.diff.Apply()
}

func (s *stateDB) addTxKey(pubKey []byte, txIdx int, blk *pb.Block) error {
	h, err := coinutil.HashHeader(blk.Header)
	if err != nil {
		return err
	}
	txKey := &TxKey{TxIdx: txIdx, BlkHash: h}
	txKeyB, err := json.Marshal(txKey)
	if err != nil {
		return err
	}
	return s.accountsTrie.Put(accountTxKeysKey(pubKey, txIdx, blk.BlockNumber()), txKeyB)
}

func (s *stateDB) removeTxKey(pubKey []byte, txIdx int, blk *pb.Block) error {
	return s.accountsTrie.Delete(accountTxKeysKey(pubKey, txIdx, blk.BlockNumber()))
}

// accountTxHashesKey returns the key used to store a transaction related
// to an account given its public key. We use the pubKey of the account so that
// all the transactions of an account are stored in the account subtree.
// The block height is used so that transactions are stored chronologically.
func accountTxKeysKey(pubKey []byte, txIdx int, blkHeight uint64) []byte {
	return append(append(pubKey, encodeUint64(blkHeight)...), encodeUint64(uint64(txIdx))...)
}

// prevNoncesKey returns the previous nonces key for the given state ID.
func (s *stateDB) prevNoncesKey(stateID []byte) []byte {
	return append(append(s.prefix, prevNoncesPrefix), stateID...)
}

// encodeUint64 encodes an uint64 to a buffer.
func encodeUint64(value uint64) []byte {
	v := make([]byte, 8)
	binary.LittleEndian.PutUint64(v, value)

	return v
}
