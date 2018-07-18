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

package db

import "errors"

var (
	// ErrNotFound is returned when a key doesn't exist.
	ErrNotFound = errors.New("key does not exist")

	// ErrInvalidBatch is returned when a batch is invalid.
	ErrInvalidBatch = errors.New("batch is invalid")

	// ErrInvalidTransaction is returned when a transaction is invalid.
	ErrInvalidTransaction = errors.New("transaction is invalid")
)

// DB can read and write values to a key-value database.
//
// Implementations should be concurrently safe.
type DB interface {
	Reader
	Ranger
	Writer
	Batcher
	Transactor

	// Close may close the underlying storage.
	Close() error
}

// Reader reads values from a key-value database.
type Reader interface {
	// Get returns the value of a key. If the key doesn't exist, it returns
	// ErrNotFound (which feels a bit safer than returning nil).
	Get(key []byte) ([]byte, error)
}

// Ranger can iterate over ranges.
type Ranger interface {
	// IterateRange creates an iterator that iterates from the given start
	// key (inclusive) up to the given stop key (exclusive). Remember to
	// call Release() on the iterator.
	IterateRange(start, stop []byte) Iterator

	// IteratePrefix creates an iterator that iterates over all the keys
	// that begin with the given prefix. Remember to call Release() on the
	// iterator.
	IteratePrefix(prefix []byte) Iterator
}

// Writer writes values to a key-value database.
type Writer interface {
	// Put sets or overwrites the value of a key.
	Put(key, value []byte) error

	// Delete removes a key. If the key doesn't exist, it is a NOP.
	Delete(key []byte) error
}

// Batcher can batch write.
type Batcher interface {
	// Batch creates a batch which can be used to execute multiple write
	// operations atomically and efficiently.
	Batch() Batch

	// Write executes all the operations in a batch atomically. It may
	// write concurrently.
	Write(Batch) error
}

// ReadWriter can both read and write values to a key-value database.
type ReadWriter interface {
	Reader
	Writer
}

// ReadWriteBatcher can both read, write, and batch write values to a key-value
// database.
type ReadWriteBatcher interface {
	ReadWriter
	Batcher
}

// Iterator iterates over a range of keys in the key-value database.
type Iterator interface {
	// Next returns whether there are keys left.
	Next() (bool, error)

	// Key returns the key of the current entry.
	Key() []byte

	// Value returns the value of the current entry.
	Value() []byte

	// Release needs to be called to free the iterator.
	Release()
}

// Batch can execute multiple write operations efficiently. It is only intended
// to be used once. When writing the batch to the database, the operations may
// execute concurrently, so order is not guaranteed.
type Batch interface {
	// Put sets or overwrites the value of a key.
	Put(key, value []byte)

	// Delete removes a key.
	Delete(key []byte)
}

// Transactor can create database transactions.
type Transactor interface {
	// Transaction creates a transaction which can be used to execute
	// multiple operations atomically.
	Transaction() (Transaction, error)
}

// Transaction can be used to execute multiple operations atomically. It should
// be closed exactly once by calling either Commit or Discard.
type Transaction interface {
	Reader
	Ranger
	Writer
	Batcher

	// Commit should be called to commit the transaction to the database.
	Commit() error

	// Discard should be called to discard the transaction.
	Discard()
}
