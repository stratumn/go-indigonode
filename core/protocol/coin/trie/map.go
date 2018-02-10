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

package trie

import (
	"github.com/pkg/errors"
	"github.com/stratumn/alice/core/protocol/coin/db"
)

// mapDB implements a simple db.ReadWritter.
type mapDB struct {
	vals map[string][]byte
}

// newMapDB creates a new map.
func newMapDB() *mapDB {
	return &mapDB{
		vals: map[string][]byte{},
	}
}

// Get reads a value from the map.
func (m *mapDB) Get(key []byte) ([]byte, error) {
	v, ok := m.vals[string(key)]
	if !ok {
		return nil, errors.WithStack(db.ErrNotFound)
	}

	return v, nil
}

// Put puts a value in the map.
func (m *mapDB) Put(key, val []byte) error {
	m.vals[string(key)] = val

	return nil
}

// Delete removes a value from the map. Deleting a non-existing value is a NOP.
func (m *mapDB) Delete(key []byte) error {
	delete(m.vals, string(key))

	return nil
}
