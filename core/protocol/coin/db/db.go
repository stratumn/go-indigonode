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
)

// Reader reads values from a key-value database.
type Reader interface {
	// Get returns the value of a key. If the key doesn't exist, it returns
	// ErrNotFound (which feels a bit safer than returning nil).
	Get(key []byte) ([]byte, error)
}

// Writer writes values to a key-value database.
type Writer interface {
	// Put sets or overwrites the value of a key.
	Put(key, value []byte) error

	// Delete removes a key. If the key doesn't exist, it returns
	// ErrNotFound (which feels a bit safer than a NOP).
	Delete(key []byte) error
}

// DB can read and write values to a key-value database.
//
// Implementations should be concurrently safe.
type DB interface {
	Reader
	Writer

	// Batch returns a type that makes it possible to commit multiple write
	// operations at once. A batch is only intended to be commited once.
	Batch() Batch

	// Close may close the underlying storage.
	Close() error
}
