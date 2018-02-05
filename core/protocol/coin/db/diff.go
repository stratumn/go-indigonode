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

import (
	"sync"
)

// Diff implements ReadWriter and records changes that need to be applied to an
// underlying database. The recorded changes can then be applied atomically
// using a batch. It ensures only one write operation per key changed.
type Diff struct {
	db ReadWriteBatcher

	mu      sync.RWMutex
	updated map[string][]byte
	deleted map[string]struct{}
}

// NewDiff creates a new Diff with the given underlying database.
func NewDiff(db ReadWriteBatcher) *Diff {
	return &Diff{
		db:      db,
		updated: map[string][]byte{},
		deleted: map[string]struct{}{},
	}
}

// Get returns the recorded value or the value of the underlying database.
func (d *Diff) Get(key []byte) ([]byte, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	// NOTE: Go has optimizations for `map[(string(key)]` so it's better
	// not to create a variable for the key.
	if _, ok := d.deleted[string(key)]; ok {
		return nil, ErrNotFound
	}

	v, ok := d.updated[string(key)]
	if ok {
		return v, nil
	}

	return d.db.Get(key)
}

// Put records an updated value.
func (d *Diff) Put(key, value []byte) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.updated[string(key)] = value
	delete(d.deleted, string(key))

	return nil
}

// Delete records a deleted value.
func (d *Diff) Delete(key []byte) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.deleted[string(key)] = struct{}{}
	delete(d.updated, string(key))

	return nil
}

// Apply applies all the recorded changes to the underlying database. It also
// resets the recorded changes.
func (d *Diff) Apply() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	batch := d.db.Batch()

	for k, v := range d.updated {
		batch.Put([]byte(k), v)
	}

	for k := range d.deleted {
		batch.Delete([]byte(k))
	}

	if err := d.db.Write(batch); err != nil {
		return err
	}

	d.doReset()

	return nil
}

// Reset resets all the recorded changes.
func (d *Diff) Reset() {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.doReset()
}

func (d *Diff) doReset() {
	d.updated = map[string][]byte{}
	d.deleted = map[string]struct{}{}
}
