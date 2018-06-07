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
	"container/list"

	"gx/ipfs/QmZyZDi491cCNTLfAhwcaDii2Kg4pwKRkhqQzURGDvY6ua/go-multihash"
)

// cacheEntry represents an entry in the cash.
type cacheEntry struct {
	Key     []uint8
	Node    node
	Dirty   bool
	Deleted bool
	InLRU   bool
	Hash    multihash.Multihash
}

// cache stores unencoded nodes in memory, and also handles lazy-hashing of
// nodes. It uses an underlying database to get nodes that are not in the
// cache. It keeps track of dirty nodes, and can commit the changes to the
// underlying database.
type cache struct {
	ndb nodeDB

	elements map[string]*list.Element

	lru     *list.List
	lruSize int

	hashCode uint64
}

// newCache creates a new cache. If hashCode is non-zero, it will update the
// hashes of nodes when committing the dirty nodes to the underlying database.
// The LRU size is the maximum number of non-modified nodes to keep in memory.
// All dirty nodes are kept in memory.
func newCache(ndb nodeDB, hashCode uint64, lruSize int) *cache {
	c := &cache{
		ndb:      ndb,
		elements: map[string]*list.Element{},
		lruSize:  lruSize,
		hashCode: hashCode,
	}

	if lruSize > 0 {
		c.lru = list.New()
	}

	return c
}

// Get gets a node from the cache. It will load the value from the database if
// it isn't in the cache.
func (c *cache) Get(key []uint8) (node, error) {
	if elem, ok := c.elements[string(key)]; ok {
		entry := elem.Value.(*cacheEntry)

		c.touch(entry)

		if entry.Deleted {
			return null{}, nil
		}

		return entry.Node, nil
	}

	n, err := c.ndb.Get(key)
	if err != nil {
		return null{}, err
	}

	c.touch(&cacheEntry{Key: key, Node: n})

	return n, nil
}

// Put sets a node in the cache.
func (c *cache) Put(key []uint8, n node) error {
	c.touch(&cacheEntry{Key: key, Node: n, Dirty: true})

	return nil
}

// Delete marks a node as deleted.
func (c *cache) Delete(key []uint8) error {
	c.touch(&cacheEntry{Key: key, Dirty: true, Deleted: true})

	return nil
}

// Commit writes all the dirty nodes to the database. It does not reset the
// cache.
//
// If a hashCode was given, the hashes will be updated before the nodes are
// sent to the underlying database.
func (c *cache) Commit() error {
	// Update hashes if needed.
	if elem, ok := c.elements[string([]uint8(nil))]; ok {
		root := elem.Value.(*cacheEntry)

		if root.Dirty && !root.Deleted {
			if _, err := c.Hash(nil); err != nil {
				return err
			}
		}
	}

	for k, elem := range c.elements {
		entry := elem.Value.(*cacheEntry)
		key := []uint8(k)

		if entry.Dirty {
			if entry.Deleted {
				if err := c.ndb.Delete(key); err != nil {
					return err
				}
			} else {
				if err := c.ndb.Put(key, entry.Node); err != nil {
					return err
				}
			}

			entry.Dirty = false

			// If the node is not in the LRU, we no longer need to
			// keep it in memory.
			if !entry.InLRU {
				delete(c.elements, string(key))
			}
		}
	}

	return nil
}

// Reset deletes all entries.
func (c *cache) Reset() {
	c.elements = map[string]*list.Element{}

	if c.lruSize > 0 {
		c.lru.Init()
	}
}

// Hash returns the hash of a node.
func (c *cache) Hash(key []uint8) (multihash.Multihash, error) {
	if c.hashCode == 0 {
		return nil, nil
	}

	l := len(key)
	key = key[:l:l]

	n, err := c.Get(key)
	if err != nil {
		return nil, err
	}

	var entry *cacheEntry

	elem, ok := c.elements[string(key)]
	if ok {
		entry = elem.Value.(*cacheEntry)
	} else {
		// If the LRU size is zero, the element could be nil because
		// c.touch() doesn't save the node unless it's been modified.
		entry = &cacheEntry{Node: n}
	}

	if entry.Hash != nil {
		// This is safe because calling c.Put() resets it, so it's
		// either nil or up-to-date.
		return entry.Hash, nil
	}

	if b, ok := entry.Node.(*branch); ok {
		// Note: we don't need to clone the node before changing it
		// because this only happens in the global cache and there
		// shouldn't be other refs to it. This relies on an
		// internal implementation detail, but cloning nodes is
		// expansive.
		for _, e := range b.EmbeddedNodes {
			e, ok := e.(*edge)
			if !ok {
				continue
			}

			// Already up-to-date.
			if len(e.Hash) > 0 {
				continue
			}

			path := append(key, e.Path...)

			hash, err := c.Hash(path)
			if err != nil {
				return nil, err
			}

			e.Hash = hash
			entry.Dirty = true
		}
	}

	// Make sure it is in c.elements if the node was modified even if the
	// LRU size is zero.
	c.touch(entry)

	hash, err := nodeToProof(key, entry.Node).Hash(c.hashCode)
	if err != nil {
		return nil, err
	}

	entry.Hash = hash

	return hash, nil
}

// touch puts an entry at the front of the LRU list. It will also remove the
// oldest entry if needed.
func (c *cache) touch(entry *cacheEntry) {
	if c.lruSize < 1 {
		if entry.Dirty {
			c.elements[string(entry.Key)] = &list.Element{Value: entry}
		}

		return
	}

	if elem, ok := c.elements[string(entry.Key)]; ok {
		entry.InLRU = elem.Value.(*cacheEntry).InLRU
		elem.Value = entry

		if entry.InLRU {
			c.lru.MoveToFront(elem)
			return
		}
	}

	entry.InLRU = true
	elem := c.lru.PushFront(entry)
	c.elements[string(entry.Key)] = elem

	if c.lru.Len() <= c.lruSize {
		return
	}

	back := c.lru.Back()
	if back == nil {
		return
	}

	c.lru.Remove(back)

	oldest := back.Value.(*cacheEntry)
	oldest.InLRU = false

	if !oldest.Dirty {
		delete(c.elements, string(oldest.Key))
	}
}
