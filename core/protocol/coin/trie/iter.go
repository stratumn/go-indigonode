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
	trie  *Trie
	start []uint8
	stop  []uint8

	mu  sync.RWMutex
	key []byte
	val []byte
}

// newIter creates a new iterator.
func newIter(trie *Trie, start, stop []byte) *iter {
	return &iter{
		trie:  trie,
		start: newNibs(start, false).Expand(),
		stop:  newNibs(stop, false).Expand(),
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

	// We also need to read-lock the tree.
	i.trie.mu.RLock()
	defer i.trie.mu.RUnlock()

	defer i.trie.atomicCache.Reset()

	root, err := i.trie.rootNode()
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
		n, err = i.trie.getNode(prefix)
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
