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
	mu     sync.RWMutex
	values map[string][]byte
}

// NewMemDB creates a new in-memory key-value database.
//
// It can be used for testing, but it will run out of memory quickly under
// normal circumstances.
func NewMemDB() DB {
	return &memDB{values: map[string][]byte{}}
}

func (db *memDB) Get(key []byte) ([]byte, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	v, ok := db.values[hex.EncodeToString(key)]
	if !ok {
		return nil, errors.WithStack(ErrNotFound)
	}

	return v, nil
}

func (db *memDB) Put(key, value []byte) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	db.values[hex.EncodeToString(key)] = value

	return nil
}

func (db *memDB) Delete(key []byte) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	k := hex.EncodeToString(key)

	if _, ok := db.values[k]; !ok {
		return errors.WithStack(ErrNotFound)
	}

	delete(db.values, k)

	return nil
}

// Close doesn't do anything in this implementation.
func (db *memDB) Close() error {
	return nil
}
