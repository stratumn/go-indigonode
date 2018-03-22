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
	"context"
	"sync"

	"github.com/pkg/errors"
	"github.com/stratumn/alice/core/db"

	"gx/ipfs/QmZyZDi491cCNTLfAhwcaDii2Kg4pwKRkhqQzURGDvY6ua/go-multihash"
)

var (
	// ErrInvalidHash is returned when a node has an unexpected hash.
	ErrInvalidHash = errors.New("the node's hash is invalid")
)

// Opt is an option for a Patricia Merkle Trie.
type Opt func(*Trie)

// OptDB sets the database. By default it uses an in-memory map.
func OptDB(db db.ReadWriter) Opt {
	return func(t *Trie) {
		t.dbrw = db
	}
}

// OptPrefix sets a prefix for all the database keys.
func OptPrefix(prefix []byte) Opt {
	// To make sure append creates a new array we set a cap.
	l := len(prefix)

	return func(t *Trie) {
		t.prefix = prefix[:l:l]
	}
}

// OptHashCode sets the hash algorithm (see Multihash for codes).
//
// The default value is SHA2_256.
func OptHashCode(code uint64) Opt {
	return func(t *Trie) {
		t.hashCode = code
	}
}

// OptCacheSize sets the number of nodes kept in the cache.
//
// The default value is 1000.
func OptCacheSize(size int) Opt {
	return func(t *Trie) {
		t.cacheSize = size
	}
}

// Trie represents a Patricia Merkle Trie.
//
// It implements db.ReadWriter and db.Ranger, though currently it doesn't
// support empty values, so calling Put(key, nil) is equivalent to calling
// Delete(key).
//
// TODO: this could be changed by adding a null flag to nodes.
//
// The underlying database is not modified until Commit() is called.
type Trie struct {
	mu    sync.RWMutex
	dbrw  db.ReadWriter
	cache *cache

	prefix    []byte
	hashCode  uint64
	cacheSize int
}

// New create a new Patricia Merkle Trie.
func New(opts ...Opt) *Trie {
	t := &Trie{
		hashCode:  multihash.SHA2_256,
		cacheSize: 1000,
	}

	for _, o := range opts {
		o(t)
	}

	if t.dbrw == nil {
		t.cache = newCache(newMapNodeDB(), t.hashCode, t.cacheSize)
	} else {
		t.cache = newCache(newDbrwNodeDB(t.dbrw, t.prefix), t.hashCode, t.cacheSize)
	}

	return t
}

// Commit applies all the changes to the database since the last call to
// Commit() or Reset().
func (t *Trie) Commit() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	return t.cache.Commit()
}

// Reset resets all the uncommited changes.
func (t *Trie) Reset() {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.cache.Reset()
}

// Get gets the value of the given key.
//
// The value is not copied, so the caller should copy it if it plans on using
// it elsewhere.
func (t *Trie) Get(key []byte) ([]byte, error) {
	return t.SubtrieGet(nil, key)
}

// SubtrieGet gets the value of the given key relative to the given subtrie
// key.
func (t *Trie) SubtrieGet(subtrieKey, key []byte) ([]byte, error) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	subtrieKey = newNibsWithoutCopy(subtrieKey, false).Expand()
	key = newNibsWithoutCopy(key, false).Expand()

	n, err := t.cache.Get(append(subtrieKey, key...))
	if err != nil {
		return nil, err
	}

	switch n := n.(type) {
	case null:
		return nil, errors.WithStack(db.ErrNotFound)

	case *branch:
		if len(n.Value) < 1 {
			return nil, errors.WithStack(db.ErrNotFound)
		}

		return n.Value, nil

	case *leaf:
		if len(n.Value) < 1 {
			return nil, errors.WithStack(db.ErrNotFound)
		}

		return n.Value, nil
	}

	return nil, errors.WithStack(ErrInvalidNodeType)
}

// IterateRange creates an iterator that iterates from the given start
// key (inclusive) up to the given stop key (exclusive). Remember to call
// Release() on the iterator.
func (t *Trie) IterateRange(start, stop []byte) db.Iterator {
	return t.SubtrieIterateRange(nil, start, stop)
}

// SubtrieIterateRange creates an iterator that iterates from the given start
// key (inclusive) up to the given stop key (exclusive) relative to the given
// subtrie key. Remember to call Release() on the iterator.
func (t *Trie) SubtrieIterateRange(subtrieKey, start, stop []byte) db.Iterator {
	subtrieKey = newNibsWithoutCopy(subtrieKey, false).Expand()

	return newIter(t, subtrieKey, start, stop)
}

// IteratePrefix creates an iterator that iterates over all the keys
// that begin with the given prefix. Remember to call Release() on the
// iterator.
func (t *Trie) IteratePrefix(prefix []byte) db.Iterator {
	return t.SubtrieIteratePrefix(nil, prefix)
}

// SubtrieIteratePrefix creates an iterator that iterates over all the keys
// that begin with the given prefix relative to the given subtrie. Remember to
// call Release() on the iterator.
func (t *Trie) SubtrieIteratePrefix(subtrieKey, prefix []byte) db.Iterator {
	subtrieKey = newNibsWithoutCopy(subtrieKey, false).Expand()

	// Taken from goleveldb.
	var stop []byte

	for i := len(prefix) - 1; i >= 0; i-- {
		c := prefix[i]
		if c < 0xff {
			stop = make([]byte, i+1)
			copy(stop, prefix)
			stop[i] = c + 1
			break
		}
	}

	return newIter(t, subtrieKey, prefix, stop)
}

// MerkleRoot returns the hash of the root node.
func (t *Trie) MerkleRoot() (multihash.Multihash, error) {
	return t.SubtrieMerkleRoot(nil)
}

// SubtrieMerkleRoot returns the hash of the given subtrie key.
func (t *Trie) SubtrieMerkleRoot(subtrieKey []byte) (multihash.Multihash, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	return t.cache.Hash(newNibsWithoutCopy(subtrieKey, false).Expand())
}

// Proof returns a proof of the value for the given key.
func (t *Trie) Proof(key []byte) (Proof, error) {
	return t.SubtrieProof(nil, key)
}

// SubtrieProof returns a proof of the value for the given key relative to
// the given subtrie key.
func (t *Trie) SubtrieProof(subtrieKey, key []byte) (Proof, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	subtrieKey = newNibsWithoutCopy(subtrieKey, false).Expand()
	key = newNibsWithoutCopy(key, false).Expand()

	// Compute hashes if needed.
	if _, err := t.cache.Hash(subtrieKey); err != nil {
		return nil, err
	}

	root, err := t.cache.Get(subtrieKey)
	if err != nil {
		return nil, err
	}

	proof, found, err := t.recProof(root, subtrieKey, key)
	if err != nil {
		return nil, err
	}

	if !found {
		return nil, errors.WithStack(db.ErrNotFound)
	}

	return proof, nil
}

// recProof goes down the trie to the given key and returns all the nodes
// visited bottom up.
//
//	- prefix is the part of the key visited so far
//	- key is the part of the key left to visit
//	- node is the node corresponding to the prefix
func (t *Trie) recProof(n node, prefix, key []uint8) ([]ProofNode, bool, error) {
	l := len(prefix)
	prefix = prefix[:l:l]

	if e, ok := n.(*edge); ok {
		// Follow edge but don't include it in the proof.
		if !bytes.HasPrefix(key, e.Path) {
			return nil, false, nil
		}

		key = key[len(e.Path):]
		prefix = append(prefix, e.Path...)

		var err error

		n, err = t.cache.Get(prefix)
		if err != nil {
			return nil, false, err
		}
	}

	if len(key) < 1 {
		// The end of the key was reached, so the current node is
		// the last one.
		return []ProofNode{nodeToProof(prefix, n)}, true, nil
	}

	switch n := n.(type) {
	case null:
		if len(prefix) > 0 {
			return nil, false, errors.WithStack(ErrInvalidNodeType)
		}

		return nil, false, nil

	case *branch:
		// The embedded node will be an edge node if a child already
		// exists, null otherwise.
		e := n.EmbeddedNodes[key[0]]

		switch e.(type) {
		case null:
			return nil, false, nil

		case *edge:
			nodes, found, err := t.recProof(e, prefix, key)
			if err != nil {
				return nil, false, err
			}

			return append(nodes, nodeToProof(prefix, n)), found, nil
		}

	case *leaf:
		return nil, false, nil
	}

	return nil, false, errors.WithStack(ErrInvalidNodeType)
}

// Put sets the value of the given key. Putting a nil or empty value is the
// same as deleting it.
//
// The value is not copied, so the caller should copy it beforehand if it plans
// on using it elsewhere.
func (t *Trie) Put(key, value []byte) error {
	return t.SubtriePut(nil, key, value)
}

// SubtriePut sets the value of the given key relative to the given subtrie
// key. Putting a nil or empty value is the same as deleting it.
//
// The value is not copied, so the caller should copy it beforehand if it plans
// on using it elsewhere.
func (t *Trie) SubtriePut(subtrieKey, key, value []byte) error {
	if len(value) < 1 {
		return t.Delete(key)
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	subtrieKey = newNibsWithoutCopy(subtrieKey, false).Expand()
	key = newNibsWithoutCopy(key, false).Expand()

	c := t.createTmpCache()
	defer c.Reset()

	root, err := c.Get(nil)
	if err != nil {
		return err
	}

	_, _, err = t.recPut(c, root, nil, append(subtrieKey, key...), value)

	if err != nil {
		return err
	}

	// Commit the atomic cache (an error should never occur).
	return c.Commit()
}

// recPut recursively puts a value in the trie, creating new nodes if
// necessary.
//
//	- prefix is the part of the key visited so far
//	- key is the part of the key left to visit
//	- node is the node corresponding to the prefix
//
// It returns the updated node and its key.
func (t *Trie) recPut(c *cache, n node, prefix, key []uint8, value []byte) (node, []uint8, error) {
	l := len(prefix)
	prefix = prefix[:l:l]

	if e, ok := n.(*edge); ok {
		// Handle edge cases.
		i := 0
		for ; i < len(key) && i < len(e.Path); i++ {
			if key[i] != e.Path[i] {
				break
			}
		}

		if i < len(key) && i < len(e.Path) {
			// Fork.
			//
			// \        \
			//  \        C
			//   \      / \
			//    A    A   B
			return t.fork(c, e, prefix, key, value, i)
		}

		if i < len(e.Path) {
			// Split.
			//
			// \       \
			//  \       B
			//   \       \
			//    A       A
			return t.split(c, e, prefix, key, value, i)
		}

		// Other cases can be handled normally.
		prefix = append(prefix, e.Path...)
		key = key[len(e.Path):]

		var err error

		n, err = c.Get(prefix)
		if err != nil {
			return null{}, nil, err
		}
	}

	if len(key) < 1 {
		// The end of the key was reached, so the current node is
		// the one that needs to be updated.
		return t.putNodeValue(c, n, prefix, value)
	}

	// Otherwise this node has to be turned into a branch if not already
	// one or updated.
	return t.putChildValue(c, n, prefix, key, value)
}

// putNodeValue sets the value of a node.
func (t *Trie) putNodeValue(c *cache, n node, prefix []uint8, value []byte) (node, []uint8, error) {
	switch v := n.(type) {
	case null:
		n = &leaf{Value: value}
	case *branch:
		v = v.Clone().(*branch)
		v.Value = value
		n = v
	case *leaf:
		v = v.Clone().(*leaf)
		v.Value = value
		n = v
	default:
		return null{}, nil, errors.WithStack(ErrInvalidNodeType)
	}

	if err := c.Put(prefix, n); err != nil {
		return null{}, nil, err
	}

	return n, prefix, nil
}

// putChildValue upgrades a node to a branch if needed and sets the value a
// child.
func (t *Trie) putChildValue(c *cache, n node, prefix, key []uint8, value []byte) (node, []uint8, error) {
	var b *branch

	switch n := n.(type) {
	case null:
		if len(prefix) > 0 {
			return null{}, nil, errors.WithStack(ErrInvalidNodeType)
		}

		b = newEmptyBranch()
	case *branch:
		b = n.Clone().(*branch)
	case *leaf:
		// Upgrade leaf to a branch.
		b = newEmptyBranch()
		b.Value = n.Value
	default:
		return null{}, nil, errors.WithStack(ErrInvalidNodeType)
	}

	// The embedded node will be an edge node if a child already exists,
	// null otherwise.
	e := b.EmbeddedNodes[key[0]]
	if _, ok := e.(null); ok {
		e = &edge{Path: key}
	}

	_, path, err := t.recPut(c, e, prefix, key, value)
	if err != nil {
		return null{}, nil, err
	}

	// Hash will be lazily computed.
	b.EmbeddedNodes[key[0]] = &edge{Path: path[len(prefix):]}

	if err := c.Put(prefix, b); err != nil {
		return null{}, nil, err
	}

	return b, prefix, nil
}

// fork forks an edge, creating two new nodes.
func (t *Trie) fork(c *cache, e *edge, prefix, key []uint8, value []byte, i int) (node, []uint8, error) {
	b := newEmptyBranch()

	// Existing node.
	b.EmbeddedNodes[e.Path[i]] = &edge{
		Path: e.Path[i:],
		Hash: e.Hash,
	}

	// New node.
	if err := c.Put(append(prefix, key...), &leaf{Value: value}); err != nil {
		return null{}, nil, err
	}

	// Hash will be lazily computed.
	b.EmbeddedNodes[key[i]] = &edge{Path: key[i:]}

	// Save forking branch.
	path := append(prefix, key[:i]...)
	if err := c.Put(path, b); err != nil {
		return null{}, nil, err
	}

	return b, path, nil
}

// split splits an edge in two.
func (t *Trie) split(c *cache, e *edge, prefix, key []uint8, value []byte, i int) (node, []uint8, error) {
	b := newEmptyBranch()
	b.Value = value

	b.EmbeddedNodes[e.Path[i]] = &edge{
		Path: e.Path[i:],
		Hash: e.Hash,
	}

	path := append(prefix, key[:i]...)
	if err := c.Put(path, b); err != nil {
		return null{}, nil, err
	}

	return b, path, nil
}

// Delete removes the value for the given key. Deleting a non-existing key is
// a NOP.
func (t *Trie) Delete(key []byte) error {
	return t.SubtrieDelete(nil, key)
}

// SubtrieDelete removes the value for the given key relative to the given
// subtrie key. Deleting a non-existing key is a NOP.
func (t *Trie) SubtrieDelete(subtrieKey, key []byte) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	subtrieKey = newNibsWithoutCopy(subtrieKey, false).Expand()
	key = newNibsWithoutCopy(key, false).Expand()

	c := t.createTmpCache()
	defer c.Reset()

	root, err := c.Get(nil)
	if err != nil {
		return err
	}

	_, _, _, err = t.recDelete(c, root, nil, append(subtrieKey, key...))

	if err != nil {
		return err
	}

	// Commit the atomic cache (an error should never occur).
	return c.Commit()
}

// recDelete recursively removes a value in the trie, deleting nodes if
// needed.
//
//	- prefix is the part of the key visited so far
//	- key is the part of the key left to visit
//	- node is the node corresponding to the prefix
//
// It returns the updated node, its path, and whether the key was found.
func (t *Trie) recDelete(c *cache, n node, prefix, key []uint8) (node, []uint8, bool, error) {
	l := len(prefix)
	prefix = prefix[:l:l]

	var err error

	if e, ok := n.(*edge); ok {
		key = key[len(e.Path):]
		prefix = append(prefix, e.Path...)

		n, err = c.Get(prefix)
		if err != nil {
			return null{}, nil, false, err
		}
	}

	if len(key) < 1 {
		return t.deleteNodeVal(c, n, prefix)
	}

	// We're not at the end of the key yet, so we have to update the
	// branch.
	return t.deleteChildVal(c, n, prefix, key)
}

// deleteNodeVal deletes a value from a node, which may delete nodes along the
// ways.
func (t *Trie) deleteNodeVal(c *cache, n node, prefix []uint8) (node, []uint8, bool, error) {
	switch n := n.(type) {
	case null:
		// NOP.
		return null{}, nil, false, nil

	case *branch:
		if len(prefix) > 0 {
			var lastEdge *edge
			numEdges := 0

			for _, n := range n.EmbeddedNodes {
				if e, ok := n.(*edge); ok {
					lastEdge = e
					numEdges++
				}
			}

			if numEdges == 1 {
				// Join edges.
				//
				// \      \
				// [x]     \
				//   \      \
				//    A      A
				if err := c.Delete(prefix); err != nil {
					return null{}, nil, false, err
				}

				childPrefix := append(prefix, lastEdge.Path...)

				child, err := c.Get(childPrefix)
				if err != nil {
					return null{}, nil, false, err
				}

				return child, childPrefix, true, nil
			}
		}

		n = n.Clone().(*branch)
		n.Value = nil
		if err := c.Put(prefix, n); err != nil {
			return null{}, nil, false, err
		}

		return n, prefix, true, nil

	case *leaf:
		if err := c.Delete(prefix); err != nil {
			return null{}, nil, false, err
		}

		return null{}, nil, true, nil
	}

	return null{}, nil, false, errors.WithStack(ErrInvalidNodeType)
}

// deleteChildVal removes a value from a child node.
func (t *Trie) deleteChildVal(c *cache, n node, prefix, key []uint8) (node, []uint8, bool, error) {
	switch n := n.(type) {
	case null, *leaf:
		// NOP.
		return null{}, nil, false, nil

	case *branch:
		e := n.EmbeddedNodes[key[0]]

		switch e := e.(type) {
		case null:
			// NOP.
			return null{}, nil, false, nil

		case *edge:
			if !bytes.HasPrefix(key, e.Path) {
				// NOP.
				return null{}, nil, false, nil
			}

			child, childPrefix, found, err := t.recDelete(c, e, prefix, key)
			if err != nil {
				return null{}, nil, false, err
			}

			if !found {
				// NOP.
				return null{}, nil, false, nil
			}

			n = n.Clone().(*branch)

			// Recompute edge and hash children.
			switch child.(type) {
			case null:
				n.EmbeddedNodes[key[0]] = null{}

			case *branch, *leaf:
				// Hash will be lazily computed.
				n.EmbeddedNodes[key[0]] = &edge{
					Path: childPrefix[len(prefix):],
				}

			default:
				return null{}, nil, false, errors.WithStack(ErrInvalidNodeType)
			}

			return t.restructureBranch(c, n, prefix, key)

		default:
			return null{}, nil, false, errors.WithStack(ErrInvalidNodeType)
		}
	}

	return null{}, nil, false, errors.WithStack(ErrInvalidNodeType)
}

// restructureBranch restructures a branch after one of its child node was
// deleted.
func (t *Trie) restructureBranch(c *cache, b *branch, prefix, key []uint8) (node, []uint8, bool, error) {
	var lastEdge *edge
	numEdges := 0

	for _, n := range b.EmbeddedNodes {
		if e, ok := n.(*edge); ok {
			lastEdge = e
			numEdges++
		}
	}

	if numEdges == 0 {
		// No more children.

		if len(b.Value) > 0 {
			// The node has a value, so we can't delete it but we
			// can downgrade it to a leaf.
			l := &leaf{Value: b.Value}

			if err := c.Put(prefix, l); err != nil {
				return null{}, nil, false, err
			}

			return l, prefix, true, nil
		}

		// No value stored so we can delete the node.
		if err := c.Delete(prefix); err != nil {
			return null{}, nil, false, err
		}

		return null{}, nil, true, nil
	}

	if numEdges == 1 && len(b.Value) <= 0 && len(prefix) > 0 {
		// One child and no value so we can delete the
		// node.
		//
		// \      \
		// [x]     \
		//   \      \
		//    B      B
		//
		// Note that we can't collapse the root node, which is why we
		// checked the length of the prefix.
		if err := c.Delete(prefix); err != nil {
			return null{}, nil, false, err
		}

		childPrefix := append(prefix, lastEdge.Path...)

		child, err := c.Get(childPrefix)
		if err != nil {
			return null{}, nil, false, err
		}

		return child, childPrefix, true, nil
	}

	if err := c.Put(prefix, b); err != nil {
		return null{}, nil, false, err
	}

	return b, prefix, true, nil
}

// Check makes sure that a trie is properly structured. This can take a very
// long time.
func (t *Trie) Check(ctx context.Context) error {
	return t.SubtrieCheck(ctx, nil)
}

// SubtrieCheck makes sure that a subtrie is properly structured. This can
// take a very long time.
func (t *Trie) SubtrieCheck(ctx context.Context, subtrieKey []byte) error {
	t.mu.RLock()
	defer t.mu.RUnlock()

	subtrieKey = newNibsWithoutCopy(subtrieKey, false).Expand()

	hash, err := t.cache.Hash(subtrieKey)
	if err != nil {
		return err
	}

	root, err := t.cache.Get(subtrieKey)
	if err != nil {
		return err
	}

	return t.recCheck(ctx, root, subtrieKey, hash)
}

// recCheck recursively checks the nodes of the trie.
//
//	- prefix is the part of the key visited so far
//	- node is the node corresponding to the prefix
//	- hash is the expected hash of the node
func (t *Trie) recCheck(ctx context.Context, n node, prefix []uint8, hash multihash.Multihash) error {
	l := len(prefix)
	prefix = prefix[:l:l]

	select {
	case <-ctx.Done():
		return errors.WithStack(ctx.Err())
	default:
	}

	if e, ok := n.(*edge); ok {
		prefix = append(prefix, e.Path...)
		hash = e.Hash

		var err error

		n, err = t.cache.Get(prefix)
		if err != nil {
			return err
		}
	}

	h, err := nodeToProof(prefix, n).Hash(t.hashCode)
	if err != nil {
		return nil
	}

	if !bytes.Equal(hash, h) {
		return errors.WithStack(ErrInvalidHash)
	}

	switch n := n.(type) {
	case null:
		if len(prefix) > 0 {
			return errors.WithStack(ErrInvalidNodeType)
		}

		return nil

	case *branch:
		for _, e := range n.EmbeddedNodes {
			switch e.(type) {
			case null:
				continue

			case *edge:
				err := t.recCheck(ctx, e, prefix, nil)
				if err != nil {
					return err
				}

				continue
			}

			return errors.WithStack(ErrInvalidNodeType)
		}

		return nil

	case *leaf:
		return nil
	}

	return errors.WithStack(ErrInvalidNodeType)
}

// underlying database. It can be used to create caches for atomic operations
// to delay committing changes to the global cache until the operation has
// succeeded.
func (t *Trie) createTmpCache() *cache {
	// Use LRU size zero so it only keeps modified nodes. Hashing is
	// always done at the last minute on the global cache.
	return newCache(t.cache, 0, 0)
}
