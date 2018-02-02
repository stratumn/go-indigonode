// Copyright © 2017-2018  Stratumn SAS
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

// Reader reads values from a key-value database.
type Reader interface {
	// Get returns the value of a key. If the key doesn't exist, it returns
	// ErrNotFound (which feels a bit safer than returning nil).
	Get(key []byte) ([]byte, error)

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

	// Batch creates a batch which can be used to execute multiple write
	// operations efficiently.
	Batch() Batch

	// Write executes all the operations in a batch. It may write
	// concurrently.
	Write(Batch) error
}

// Iterator iterates over a range of keys in the key-value database.
type Iterator interface {
	// Next returns whether there are keys left.
	Next() bool

	// Key returns the key of the current entry.
	Key() []byte

	// Value returns the value of the current entry.
	Value() []byte

	// Release needs to be called to free the iterator.
	Release()
}

// DB can read and write values to a key-value database.
//
// Implementations should be concurrently safe.
type DB interface {
	Reader
	Writer

	// Transaction creates a transaction which can be used to execute
	// multiple operations atomically.
	Transaction() (Transaction, error)

	// Close may close the underlying storage.
	Close() error
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
