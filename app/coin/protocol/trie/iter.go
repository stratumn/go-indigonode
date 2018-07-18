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

package trie

import (
	"bytes"
	"sync"

	"github.com/pkg/errors"
)

var (
	// ErrIteratorReleased is returned when trying to use the iterator
	// after it was released.
	ErrIteratorReleased = errors.New("the iterator was released")
)

// iter is used to iterate ranges of a trie.
type iter struct {
	trie       *Trie
	subtrieKey []uint8
	start      []uint8
	stop       []uint8

	mu  sync.RWMutex
	key []byte
	val []byte
}

// newIter creates a new iterator.
func newIter(trie *Trie, subtrieKey []uint8, start, stop []byte) *iter {
	return &iter{
		trie:       trie,
		subtrieKey: subtrieKey,
		start:      newNibs(start, false).Expand(),
		stop:       newNibs(stop, false).Expand(),
	}
}

// Next returns whether there are keys left.
func (i *iter) Next() (bool, error) {
	i.mu.Lock()
	defer i.mu.Unlock()

	if i.trie == nil {
		return false, errors.WithStack(ErrIteratorReleased)
	}

	if len(i.stop) > 0 && bytes.Compare(i.start, i.stop) >= 0 {
		// Might as well stop early.
		return false, nil
	}

	// We also need to read-lock the trie.
	i.trie.mu.RLock()
	defer i.trie.mu.RUnlock()

	root, err := i.trie.cache.Get(i.subtrieKey)
	if err != nil {
		return false, err
	}

	return i.recNext(root, nil)
}

// Key returns the key of the current entry.
func (i *iter) Key() []byte {
	i.mu.RLock()
	defer i.mu.RUnlock()

	return i.key
}

// Value returns the value of the current entry.
func (i *iter) Value() []byte {
	i.mu.RLock()
	defer i.mu.RUnlock()

	return i.val
}

// Release needs to be called to free the iterator.
func (i *iter) Release() {
	i.mu.Lock()
	defer i.mu.Unlock()

	i.trie = nil
}

// recNext recursively finds the next pair.
func (i *iter) recNext(n node, prefix []uint8) (bool, error) {
	if e, ok := n.(*edge); ok {
		// Follow edge.
		prefix = append(prefix, e.Path...)

		var err error
		n, err = i.trie.cache.Get(prefix)
		if err != nil {
			return false, err
		}
	}

	switch n := n.(type) {
	case null:
		// Only happens if root is nil.
		return false, nil

	case *branch:
		if len(n.Value) > 0 && i.cmp(prefix) {
			// Branch has a value is in within range.
			i.set(prefix, n.Value)

			return true, nil
		}

		for _, e := range n.EmbeddedNodes {
			switch e := e.(type) {
			case null:
				// Pass.
			case *edge:
				key := append(prefix, e.Path...)

				if len(i.stop) > 0 && bytes.Compare(key, i.stop) >= 0 {
					// All children are after the range.
					return false, nil
				}

				start := i.start
				if len(key) < len(start) {
					start = start[:len(key)]
				}

				if bytes.Compare(key, start) >= 0 {
					// Children could be within range.
					next, err := i.recNext(e, prefix)
					if err != nil {
						return false, err
					}

					if next {
						return true, nil
					}
				}
			}
		}

		return false, nil

	case *leaf:
		if !i.cmp(prefix) {
			return false, nil
		}

		i.set(prefix, n.Value)

		return true, nil
	}

	return false, errors.WithStack(ErrInvalidNodeType)
}

// set sets the current key and value.
func (i *iter) set(key []uint8, val []byte) {
	// Set start to the next possible key.
	i.start = append(key, 0)

	i.key = newNibsFromNibs(key...).buf
	i.val = make([]byte, len(val))

	copy(i.val, val)
}

// cmp returns whether the key is within the iterator's range.
func (i *iter) cmp(key []uint8) bool {
	if bytes.Compare(key, i.start) < 0 {
		return false
	}

	if len(i.stop) < 1 {
		return true
	}

	return bytes.Compare(key, i.stop) < 0
}
