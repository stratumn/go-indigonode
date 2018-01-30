// Copyright © 2017-2017  Stratumn SAS
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

import (
	"encoding/hex"
	"sync"

	"github.com/pkg/errors"
)

type memDB struct {
	// If false, don't use the mutex (used by batch commits since locking
	// is unnecessary in this case).
	lock bool

	mu     sync.RWMutex
	values map[string][]byte
}

// NewMemDB creates a new in-memory key-value database.
//
// It can be used for testing or for storing small amounts of data, but it
// will run out of memory quickly otherwise.
func NewMemDB() DB {
	return &memDB{lock: true, values: map[string][]byte{}}
}

func (db *memDB) Get(key []byte) ([]byte, error) {
	if db.lock {
		db.mu.RLock()
		defer db.mu.RUnlock()
	}

	v, ok := db.values[hex.EncodeToString(key)]
	if !ok {
		return nil, errors.WithStack(ErrNotFound)
	}

	return v, nil
}

func (db *memDB) Put(key, value []byte) error {
	if db.lock {
		db.mu.Lock()
		defer db.mu.Unlock()
	}

	db.values[hex.EncodeToString(key)] = value

	return nil
}

func (db *memDB) Delete(key []byte) error {
	if db.lock {
		db.mu.Lock()
		defer db.mu.Unlock()
	}

	k := hex.EncodeToString(key)

	if _, ok := db.values[k]; !ok {
		return ErrNotFound
	}

	delete(db.values, k)

	return nil
}

func (db *memDB) Close() error {
	return nil
}

func (db *memDB) Batch() Batch {
	return newBatch(db.commit)
}

// commit commits operations from a batch.
func (db *memDB) commit(ops []interface{}) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	// Clone the values in case something goes wrong, in which
	// case the state of the DB will remain untouched.
	clone := memDB{values: map[string][]byte{}}
	for k, v := range db.values {
		clone.values[k] = v
	}

	for _, op := range ops {
		switch v := op.(type) {
		case batchPut:
			if err := clone.Put(v.key, v.value); err != nil {
				return err
			}
		case batchDelete:
			if err := clone.Delete(v.key); err != nil {
				return err
			}
		}
	}

	// Now we can safely update the values.
	db.values = clone.values

	return nil
}
