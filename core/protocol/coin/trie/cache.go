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
	multihash "github.com/multiformats/go-multihash"
	"github.com/pkg/errors"
	"github.com/stratumn/alice/core/protocol/coin/db"
)

// nodeGetter gets a node from a database.
type nodeGetter func(key []uint8) (node, error)

// nodePutter puts a node in a database.
type nodePutter func(key []uint8, n node, buffer []byte) error

// nodeDeleter deletes a node from a database.
type nodeDeleter func(key []uint8) error

// hasher hashes bytes.
type hasher func(key []byte) (multihash.Multihash, error)

// cacheEntry represents an entry in the cash.
type cacheEntry struct {
	Node    node
	Dirty   bool
	Deleted bool
	Encoded []byte
	Hash    multihash.Multihash
}

// cache stores unencoded nodes in memory, and also handles lazy-hashing and
// lazy-encoding of nodes.
type cache struct {
	entries map[string]*cacheEntry

	get nodeGetter
	put nodePutter
	del nodeDeleter

	hash hasher
}

// newCache creates a new cache.
func newCache(get nodeGetter, put nodePutter, del nodeDeleter, hash hasher) *cache {
	return &cache{
		entries: map[string]*cacheEntry{},
		get:     get,
		put:     put,
		del:     del,
		hash:    hash,
	}
}

// Get gets a node from the cache. It will load the value from the database if
// it isn't in the cache.
func (c *cache) Get(key []uint8) (node, error) {
	if entry, ok := c.entries[string(key)]; ok {
		if entry.Deleted {
			return null{}, nil
		}

		return entry.Node, nil
	}

	n, err := c.get(key)
	if err != nil {
		return null{}, err
	}

	c.entries[string(key)] = &cacheEntry{Node: n}

	return n.Clone(), nil
}

// Put sets a node in the cache.
func (c *cache) Put(key []uint8, n node) error {
	c.entries[string(key)] = &cacheEntry{Node: n, Dirty: true}

	return nil
}

// Delete marks a node as deleted.
func (c *cache) Delete(key []uint8) error {
	c.entries[string(key)] = &cacheEntry{Dirty: true, Deleted: true}

	return nil
}

// Hash retrieves the hash of a node.
func (c *cache) Hash(key []uint8) (multihash.Multihash, error) {
	if c.hash == nil {
		return nil, nil
	}

	// Make sure the node is in the cache.
	if _, err := c.Get(key); err != nil {
		return nil, err
	}

	entry, ok := c.entries[string(key)]
	if !ok || entry.Deleted {
		return nil, errors.WithStack(db.ErrNotFound)
	}

	if entry.Hash != nil && !entry.Dirty {
		return entry.Hash, nil
	}

	if b, ok := entry.Node.(*branch); ok {
		for _, e := range b.EmbeddedNodes {
			e, ok := e.(*edge)
			if !ok {
				continue
			}

			// Already hashed.
			if len(e.Hash) > 0 {
				continue
			}

			path := append(key, e.Path...)

			hash, err := c.Hash(path)
			if err != nil {
				return nil, err
			}

			e.Hash = hash
		}
	}

	buf, err := entry.Node.MarshalBinary()
	if err != nil {
		return nil, err
	}

	hash, err := c.hash(buf)
	if err != nil {
		return nil, err
	}

	entry.Encoded = buf
	entry.Hash = hash

	return hash, nil
}

// Commit writes all the dirty nodes to the database. It does not reset the
// cache.
//
// If a hasher was given, the hashes will be updated, and the putter will
// receive the encoded buffer of the node in addition to the node.
func (c *cache) Commit() error {
	if c.hash != nil {
		// Compute Merkle Root.
		if root, ok := c.entries[string([]uint8(nil))]; ok {
			if root.Dirty && !root.Deleted {
				if _, err := c.Hash(nil); err != nil {
					return err
				}
			}
		}
	}

	for k, entry := range c.entries {
		key := []uint8(k)

		if entry.Dirty {
			if entry.Deleted {
				if err := c.del(key); err != nil {
					return err
				}
			} else {
				if err := c.put(key, entry.Node, entry.Encoded); err != nil {
					return err
				}
			}

			entry.Dirty = false
		}
	}

	return nil
}

// Reset deletes all entries.
func (c *cache) Reset() {
	c.entries = map[string]*cacheEntry{}
}
