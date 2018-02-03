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

	"github.com/gogo/protobuf/proto"
	"github.com/pkg/errors"
	db "github.com/stratumn/alice/core/protocol/coin/db"
	pb "github.com/stratumn/alice/pb/coin"
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

	// ErrInvalidJobID is returned when a job ID is invalid.
	ErrInvalidJobID = errors.New("job ID is invalid")
)

// State stores users' account balances.
type State interface {
	Reader
	Writer
}

// Reader gives read access to users' account balances.
type Reader interface {
	// GetAccount gets the account details of a user identified
	// by his public key. It returns Account{} if the account is not found.
	GetAccount(pubKey []byte) (*pb.Account, error)
}

// Writer gives write access to users' account balances.
//
// NOTE: we could expose methods that take an existing database transactions
// so that you can run a transaction on both the state and the chain.
type Writer interface {
	// UpdateAccount sets or updates the account of a user identified by
	// his public key. It should only be used for testing as it cannot be
	// rolled back.
	UpdateAccount(pubKey []byte, account *pb.Account) error

	// ProcessTransactions processes all the given transactions and updates
	// the state accordingly. It should be given a unique job ID, for
	// instance the hash of the block containing the transactions.
	ProcessTransactions(jobID []byte, txs []*pb.Transaction) error

	// RollbackTransactions rolls back transactions. The parameters are
	// expected to be identical to the ones that were given to the
	// corresponding call to ProcessTransactions().
	// You should only rollback the job corresponding to the current state.
	RollbackTransactions(jobID []byte, txs []*pb.Transaction) error
}

// NOTE: prefixes should have be the same bytesize to prevent unexpected
// behaviors when iterating over key prefixes. Smaller means less space used.
var (
	// accountPrefix is the prefix for account keys.
	accountPrefix = []byte{0}

	// prevNoncePrefix is the prefix for previous nonce keys.
	prevNoncePrefix = []byte{1}
)

// stateDB implements the State interface using a key-value database.
type stateDB struct {
	db     db.DB
	prefix []byte

	// jobIDLen is the bytesize of a job ID.
	//
	// It is required because we iterate over keys prefixed by the jobID.
	// If the keys were not always the same size they could conflict.
	//
	// For example:
	//
	//	jobA  + BC -> jobABC
	//	jobAB + C  -> jobABC
	//
	// Here the key jobABC would be included whether iterating over jobA or
	// jobAB.
	jobIDSize int
}

// NewState creates a new state from a DB instance.
//
// Prefix is used to prefix keys in the database.
//
// VERY IMPORTANT NOTE: ALL THE DIFFERENT MODULES THAT SHARE THE SAME INSTANCE
// OF THE DATABASE MUST USE A UNIQUE PREFIX OF THE SAME BYTESIZE.
//
// All job IDs must have the given bytesize.
func NewState(db db.DB, prefix []byte, jobIDSize int) State {
	return &stateDB{db: db, prefix: prefix, jobIDSize: jobIDSize}
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
	err = proto.Unmarshal(buf, &account)

	return &account, errors.WithStack(err)
}

func (s *stateDB) UpdateAccount(pubKey []byte, account *pb.Account) error {
	return s.doUpdateAccount(s.db, pubKey, account)
}

// doGetAccount updates an account using anything that implements db.Writer.
func (s *stateDB) doUpdateAccount(dbw db.Writer, pubKey []byte, account *pb.Account) error {
	// Delete empty accounts to save space.
	if account.Balance == 0 && account.Nonce == 0 {
		return dbw.Delete(s.accountKey(pubKey))
	}

	buf, err := proto.Marshal(account)
	if err != nil {
		return errors.WithStack(err)
	}

	return dbw.Put(s.accountKey(pubKey), buf)
}

func (s *stateDB) ProcessTransactions(jobID []byte, txs []*pb.Transaction) error {
	if s.jobIDSize != len(jobID) {
		return ErrInvalidJobID
	}

	dbtx, err := s.db.Transaction()
	if err != nil {
		return err
	}

	// Save the current nonces.
	if err := s.doSaveNonces(dbtx, jobID, txs); err != nil {
		dbtx.Discard()
		return err
	}

	// Update balances and nonces.
	if err := s.doTransactions(dbtx, jobID, txs); err != nil {
		dbtx.Discard()
		return err
	}

	return dbtx.Commit()
}

// doSaveNonces saves the current nonces of all the senders using anything that
// implements db.ReadWriter.
func (s *stateDB) doSaveNonces(dbrw db.ReadWriter, jobID []byte, txs []*pb.Transaction) error {
	if s.jobIDSize != len(jobID) {
		return ErrInvalidJobID
	}

	// Note: it might save the nonce of the same account multiple times,
	// but it has no impact on the integrity of the state (and probably
	// very little on performance).

	for _, tx := range txs {
		account, err := s.doGetAccount(dbrw, tx.From)
		if err != nil {
			return err
		}

		buf := make([]byte, 8)
		binary.LittleEndian.PutUint64(buf, account.Nonce)

		if err := dbrw.Put(s.prevNonceKey(jobID, tx.From), buf); err != nil {
			return err
		}
	}

	return nil
}

// doTransactions updates the accounts given a slice of transactions using
// anything that implements db.ReadWriter.
func (s *stateDB) doTransactions(dbrw db.ReadWriter, jobID []byte, txs []*pb.Transaction) error {
	for _, tx := range txs {
		if err := s.doTransaction(dbrw, jobID, tx); err != nil {
			return err
		}
	}

	return nil
}

// doTransaction updates a pair of accounts given a transaction using anything
// that implements db.ReadWriter.
func (s *stateDB) doTransaction(dbrw db.ReadWriter, jobID []byte, tx *pb.Transaction) error {
	// Subtract amount from sender.
	err := s.incBalance(dbrw, tx.From, false, true, tx.Value, tx.Nonce)
	if err != nil {
		return err
	}

	// Add amount to receiver.
	return s.incBalance(dbrw, tx.To, true, false, tx.Value, 0)
}

// incBalance adds or subtracts coins from an account using anything that
// implements db.ReadWriter.
func (s *stateDB) incBalance(
	dbrw db.ReadWriter,
	pubKey []byte,
	add, updateNonce bool,
	amount, nonce uint64,
) error {
	account, err := s.doGetAccount(dbrw, pubKey)
	if err != nil {
		return nil
	}

	// Update value.
	if add {
		account.Balance += amount
	} else {
		if amount > account.Balance {
			return ErrAmountTooBig
		}

		account.Balance -= amount
	}

	if updateNonce {
		account.Nonce = nonce
	}

	return s.doUpdateAccount(dbrw, pubKey, account)
}

func (s *stateDB) RollbackTransactions(jobID []byte, txs []*pb.Transaction) error {
	if s.jobIDSize != len(jobID) {
		return ErrInvalidJobID
	}

	dbtx, err := s.db.Transaction()
	if err != nil {
		return err
	}

	// Restore balances.
	if err := s.doRollbackTransactions(dbtx, jobID, txs); err != nil {
		dbtx.Discard()
		return err
	}

	// Restore nonces.
	if err := s.doRestoreNonces(dbtx, jobID, txs); err != nil {
		dbtx.Discard()
		return err
	}

	return dbtx.Commit()
}

// doRollbackTransactions reverts transactions using anything that implements
// db.ReadWriter. It does not revert nonces.
func (s *stateDB) doRollbackTransactions(dbrw db.ReadWriter, jobID []byte, txs []*pb.Transaction) error {
	for i := len(txs) - 1; i >= 0; i-- {
		tx := txs[i]

		if err := s.doRollbackTransaction(dbrw, jobID, tx); err != nil {
			return err
		}
	}

	return nil
}

// doRollbackTransaction reverts a transaction using anything that implements
// db.ReadWriter. It does not revert the nonces.
func (s *stateDB) doRollbackTransaction(dbrw db.ReadWriter, jobID []byte, tx *pb.Transaction) error {
	// Add amount to sender.
	err := s.incBalance(dbrw, tx.From, true, false, tx.Value, 0)
	if err != nil {
		return err
	}

	// Subtract amount from receiver.
	return s.incBalance(dbrw, tx.To, false, false, tx.Value, 0)
}

// doRestoreNonces restores the nonces to what they were before the given
// transactions using anything that implements db.ReadWriter.
func (s *stateDB) doRestoreNonces(dbrw db.ReadWriter, jobID []byte, txs []*pb.Transaction) error {
	iter := dbrw.IteratePrefix(s.prevNoncePrefix(jobID))
	defer iter.Release()

	for iter.Next() {
		// Get the public key part from the key.
		from := iter.Key()[len(s.prevNoncePrefix(jobID)):]

		// Update the account nonce.
		account, err := s.doGetAccount(dbrw, from)
		if err != nil {
			return err
		}

		account.Nonce = binary.LittleEndian.Uint64(iter.Value())

		if err := s.doUpdateAccount(dbrw, from, account); err != nil {
			return err
		}

		// We can delete the nonce now to save space.
		if err := dbrw.Delete(iter.Key()); err != nil {
			return err
		}
	}

	return nil
}

// accountKey returns the key corresponding to an account given its public key.
func (s *stateDB) accountKey(pubKey []byte) []byte {
	return append(append(s.prefix, accountPrefix...), pubKey...)
}

// prevNoncePrefix returns the previous nonce key prefix for the given job ID.
func (s *stateDB) prevNoncePrefix(jobID []byte) []byte {
	return append(append(s.prefix, prevNoncePrefix...), jobID...)
}

// prevNonceKey returns the key corresponding to a previous nonce given a job
// ID and the public key of the sender.
func (s *stateDB) prevNonceKey(jobID, pubKey []byte) []byte {
	return append(s.prevNoncePrefix(jobID), pubKey...)
}
