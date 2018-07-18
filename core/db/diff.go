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

import (
	"sync"
)

// diffEntry describes an updated value.
type diffEntry struct {
	Value   []byte
	Deleted bool
}

// Diff implements ReadWriter and records changes that need to be applied to an
// underlying database. The recorded changes can then be applied atomically
// using a batch. It ensures only one write operation per key changed.
type Diff struct {
	db ReadWriteBatcher

	mu      sync.RWMutex
	entries map[string]diffEntry
}

// NewDiff creates a new Diff with the given underlying database.
func NewDiff(db ReadWriteBatcher) *Diff {
	return &Diff{
		db:      db,
		entries: map[string]diffEntry{},
	}
}

// Get returns the recorded value or the value of the underlying database.
func (d *Diff) Get(key []byte) ([]byte, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	// NOTE: Go has optimizations for `map[(string(key)]` so it's better
	// not to create a variable for the key.
	entry, ok := d.entries[string(key)]

	if !ok {
		return d.db.Get(key)
	}

	if entry.Deleted {
		return nil, ErrNotFound
	}

	return entry.Value, nil
}

// Put records an updated value.
func (d *Diff) Put(key, value []byte) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.entries[string(key)] = diffEntry{Value: value}

	return nil
}

// Delete records a deleted value.
func (d *Diff) Delete(key []byte) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.entries[string(key)] = diffEntry{Deleted: true}

	return nil
}

// Apply applies all the recorded changes to the underlying database. It also
// resets the recorded changes.
func (d *Diff) Apply() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	batch := d.db.Batch()

	for key, entry := range d.entries {
		if entry.Deleted {
			batch.Delete([]byte(key))
		} else {
			batch.Put([]byte(key), entry.Value)
		}
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
	d.entries = map[string]diffEntry{}
}
