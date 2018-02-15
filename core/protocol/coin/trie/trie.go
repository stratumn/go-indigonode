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

	"github.com/multiformats/go-multihash"
	"github.com/pkg/errors"
	"github.com/stratumn/alice/core/protocol/coin/db"
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
	return func(t *Trie) {
		t.prefix = prefix
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
	mu          sync.RWMutex
	dbrw        db.ReadWriter
	cache       *cache
	atomicCache *cache // atomicCache is used to make operations atomic.

	prefix   []byte
	hashCode uint64
}

// New create a new Patricia Merkle Trie.
func New(opts ...Opt) *Trie {
	t := &Trie{
		dbrw:     newMapDB(),
		hashCode: multihash.SHA2_256,
	}

	for _, o := range opts {
		o(t)
	}

	t.cache = newCache(t.doGetNode, t.doPutNode, t.doDeleteNode, t.hashCode)
	t.atomicCache = newCache(t.cache.Get, t.cache.Put, t.cache.Delete, 0)

	return t
}

// Commit applies all the changes to the database since the last call to
// Commit() or Reset().
func (t *Trie) Commit() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if err := t.cache.Commit(); err != nil {
		return err
	}

	t.cache.Reset()

	return nil
}

// Reset resets all the uncommited changes.
func (t *Trie) Reset() {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.cache.Reset()
}

// Get gets the value of the given key.
func (t *Trie) Get(key []byte) ([]byte, error) {
	t.mu.RLock()
	defer t.mu.RUnlock()
	defer t.atomicCache.Reset()

	n, err := t.getNode(newNibs(key, false).Expand())
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
	return newIter(t, start, stop)
}

// IteratePrefix creates an iterator that iterates over all the keys
// that begin with the given prefix. Remember to call Release() on the
// iterator.
func (t *Trie) IteratePrefix(prefix []byte) db.Iterator {
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

	return newIter(t, prefix, stop)
}

// MerkleRoot returns the hash of the root node. If there an no entries, the
// hash of Null{} is returned.
func (t *Trie) MerkleRoot() (multihash.Multihash, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	return t.cache.Hash(nil)
}

// Proof returns a proof of the value for the given key.
func (t *Trie) Proof(key []byte) (Proof, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	defer t.atomicCache.Reset()

	key = newNibsWithoutCopy(key, false).Expand()

	// Compute hashes if needed.
	if _, err := t.cache.Hash(nil); err != nil {
		return nil, err
	}

	root, err := t.rootNode()
	if err != nil {
		return nil, err
	}

	proof, found, err := t.recProof(root, nil, key)
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
	if e, ok := n.(*edge); ok {
		// Follow edge but don't include it in the proof.
		if !bytes.HasPrefix(key, e.Path) {
			return nil, false, nil
		}

		key = key[len(e.Path):]
		prefix = append(prefix, e.Path...)

		var err error

		n, err = t.getNode(prefix)
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
func (t *Trie) Put(key, value []byte) error {
	if len(value) < 1 {
		return t.Delete(key)
	}

	t.mu.Lock()
	defer t.mu.Unlock()
	defer t.atomicCache.Reset()

	root, err := t.rootNode()
	if err != nil {
		return err
	}

	key = newNibsWithoutCopy(key, false).Expand()

	_, _, err = t.recPut(root, nil, key, value)

	if err != nil {
		return err
	}

	// Commit the atomic cache (an error should never occur).
	return t.atomicCache.Commit()
}

// recPut recursively puts a value in the trie, creating new nodes if
// necessary.
//
//	- prefix is the part of the key visited so far
//	- key is the part of the key left to visit
//	- node is the node corresponding to the prefix
//
// It returns the updated node and its key.
func (t *Trie) recPut(n node, prefix, key []uint8, value []byte) (node, []uint8, error) {
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
			return t.fork(e, prefix, key, value, i)
		}

		if i < len(e.Path) {
			// Split.
			//
			// \       \
			//  \       B
			//   \       \
			//    A       A
			return t.split(e, prefix, key, value, i)
		}

		// Other cases can be handles normally.
		prefix = append(prefix, e.Path...)
		key = key[len(e.Path):]

		var err error

		n, err = t.getNode(prefix)
		if err != nil {
			return null{}, nil, err
		}
	}

	if len(key) < 1 {
		// The end of the key was reached, so the current node is
		// the one that needs to be updated.
		return t.putNodeValue(n, prefix, key, value)
	}

	// Otherwise this node has to be turned into a branch if not already
	// one or updated.
	return t.putChildValue(n, prefix, key, value)
}

// putNodeValue sets the value of a node.
func (t *Trie) putNodeValue(n node, prefix, key []uint8, value []byte) (node, []uint8, error) {
	switch v := n.(type) {
	case null:
		n = &leaf{Value: value}
	case *branch:
		v.Value = value
		n = v
	case *leaf:
		v.Value = value
		n = v
	default:
		return null{}, nil, errors.WithStack(ErrInvalidNodeType)
	}

	if err := t.putNode(prefix, n); err != nil {
		return null{}, nil, err
	}

	return n, prefix, nil
}

// putChildValue upgrades a node to a branch if needed and sets the value a
// child.
func (t *Trie) putChildValue(n node, prefix, key []uint8, value []byte) (node, []uint8, error) {
	var b *branch

	switch n := n.(type) {
	case null:
		b = newEmptyBranch()
	case *branch:
		b = n
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

	_, path, err := t.recPut(e, prefix, key, value)
	if err != nil {
		return null{}, nil, err
	}

	// Hash will be lazily computed.
	b.EmbeddedNodes[key[0]] = &edge{Path: path[len(prefix):]}

	if err := t.putNode(prefix, b); err != nil {
		return null{}, nil, err
	}

	return b, prefix, nil
}

// fork forks an edge, creating two new nodes.
func (t *Trie) fork(e *edge, prefix, key []uint8, value []byte, i int) (node, []uint8, error) {
	b := newEmptyBranch()

	// Existing node.
	b.EmbeddedNodes[e.Path[i]] = &edge{
		Path: e.Path[i:],
		Hash: e.Hash,
	}

	// New node.
	_, _, err := t.putNodeValue(
		null{},
		append(prefix, key...),
		nil,
		value,
	)
	if err != nil {
		return null{}, nil, err
	}

	// Hash will be lazily computed.
	b.EmbeddedNodes[key[i]] = &edge{Path: key[i:]}

	// Save forking branch.
	path := append(prefix, key[:i]...)
	if err := t.putNode(path, b); err != nil {
		return null{}, nil, err
	}

	return b, path, nil
}

// split splits an edge in two.
func (t *Trie) split(e *edge, prefix, key []uint8, value []byte, i int) (node, []uint8, error) {
	b := newEmptyBranch()
	b.Value = value

	b.EmbeddedNodes[e.Path[i]] = &edge{
		Path: e.Path[i:],
		Hash: e.Hash,
	}

	path := append(prefix, key[:i]...)
	if err := t.putNode(path, b); err != nil {
		return null{}, nil, err
	}

	return b, path, nil
}

// Delete removes the value for the given key. Deleting a non-existing key is
// a NOP.
func (t *Trie) Delete(key []byte) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	defer t.atomicCache.Reset()

	root, err := t.rootNode()
	if err != nil {
		return err
	}

	key = newNibsWithoutCopy(key, false).Expand()

	_, _, _, err = t.recDelete(root, nil, key)

	if err != nil {
		return err
	}

	// Commit the atomic cache (an error should never occur).
	return t.atomicCache.Commit()
}

// recDelete recursively removes a value in the trie, deleting nodes if
// needed.
//
//	- prefix is the part of the key visited so far
//	- key is the part of the key left to visit
//	- node is the node corresponding to the prefix
//
// It returns the updated node, its path, and whether the key was found.
func (t *Trie) recDelete(n node, prefix, key []uint8) (node, []uint8, bool, error) {
	var err error

	if e, ok := n.(*edge); ok {
		key = key[len(e.Path):]
		prefix = append(prefix, e.Path...)

		n, err = t.getNode(prefix)
		if err != nil {
			return null{}, nil, false, err
		}
	}

	if len(key) < 1 {
		return t.deleteNodeVal(n, prefix, key)
	}

	// We're not a the end of the key yet, so we have to update the branch.
	return t.deleteChildVal(n, prefix, key)
}

// deleteNodeVal deletes a value from a node, which may delete nodes along the
// ways.
func (t *Trie) deleteNodeVal(n node, prefix, key []uint8) (node, []uint8, bool, error) {
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
				if err := t.deleteNode(prefix); err != nil {
					return null{}, nil, false, err
				}

				childPrefix := append(prefix, lastEdge.Path...)

				child, err := t.getNode(childPrefix)
				if err != nil {
					return null{}, nil, false, err
				}

				return child, childPrefix, true, nil
			}
		}

		n.Value = nil
		if err := t.putNode(prefix, n); err != nil {
			return null{}, nil, false, err
		}

		return n, prefix, true, nil

	case *leaf:
		if err := t.deleteNode(prefix); err != nil {
			return null{}, nil, false, err
		}

		return null{}, nil, true, nil
	}

	return null{}, nil, false, errors.WithStack(ErrInvalidNodeType)
}

// deleteChildVal removes a value from a child node.
func (t *Trie) deleteChildVal(n node, prefix, key []uint8) (node, []uint8, bool, error) {
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

			child, childPrefix, found, err := t.recDelete(e, prefix, key)
			if err != nil {
				return null{}, nil, false, err
			}

			if !found {
				// NOP.
				return null{}, nil, false, nil
			}

			// Recompute edges and hash children.
			switch child.(type) {
			case null:
				n.EmbeddedNodes[key[0]] = null{}

			case *branch:
				// Hash will be lazily computed.
				n.EmbeddedNodes[key[0]] = &edge{
					Path: childPrefix[len(prefix):],
				}

			case *leaf:
				// Hash will be lazily computed.
				n.EmbeddedNodes[key[0]] = &edge{
					Path: childPrefix[len(prefix):],
				}

			default:
				return null{}, nil, false, errors.WithStack(ErrInvalidNodeType)
			}

			return t.restructureBranch(n, prefix, key)
		}
	}

	return null{}, nil, false, errors.WithStack(ErrInvalidNodeType)
}

// restructureBranch restructures a branch after one of its child node was
// deleted.
func (t *Trie) restructureBranch(b *branch, prefix, key []uint8) (node, []uint8, bool, error) {
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

			if err := t.putNode(prefix, l); err != nil {
				return null{}, nil, false, err
			}

			return l, prefix, true, nil
		}

		// No value stored so we can delete the node.
		if err := t.deleteNode(prefix); err != nil {
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
		if err := t.deleteNode(prefix); err != nil {
			return null{}, nil, false, err
		}

		childPrefix := append(prefix, lastEdge.Path...)

		child, err := t.getNode(childPrefix)
		if err != nil {
			return null{}, nil, false, err
		}

		return child, childPrefix, true, nil
	}

	if err := t.putNode(prefix, b); err != nil {
		return null{}, nil, false, err
	}

	return b, prefix, true, nil
}

// rootNode returns the root node.
func (t *Trie) rootNode() (node, error) {
	return t.getNode(nil)
}

// dbKey returns the key of a node given its key in the trie.
func (t *Trie) dbKey(key []uint8) ([]byte, error) {
	k, err := path(newNibsFromNibs(key...)).MarshalBinary()
	if err != nil {
		return nil, err
	}

	return append(t.prefix, k...), nil
}

// getNode gets a node from its key from the cache.
func (t *Trie) getNode(key []uint8) (node, error) {
	return t.atomicCache.Get(key)
}

// putNode inserts the node with the given key in the cache.
func (t *Trie) putNode(key []uint8, n node) error {
	return t.atomicCache.Put(key, n)
}

// deleteNode removes the node with the given key from the cache.
func (t *Trie) deleteNode(key []uint8) error {
	return t.atomicCache.Delete(key)
}

// doGetNode gets a node from its key from the database.
func (t *Trie) doGetNode(key []uint8) (node, error) {
	k, err := t.dbKey(key)
	if err != nil {
		return null{}, err
	}

	buf, err := t.dbrw.Get(k)
	if err != nil {
		if errors.Cause(err) == db.ErrNotFound {
			return null{}, nil
		}

		return nil, err
	}

	n, _, err := unmarshalNode(buf)

	return n, err
}

// doPutNode inserts the node with the given key in the database.
func (t *Trie) doPutNode(key []uint8, n node) error {
	k, err := t.dbKey(key)
	if err != nil {
		return err
	}

	v, err := n.MarshalBinary()
	if err != nil {
		return err
	}

	return t.dbrw.Put(k, v)
}

// doDeleteNode removes the node with the given key from the database.
func (t *Trie) doDeleteNode(key []uint8) error {
	k, err := t.dbKey(key)
	if err != nil {
		return err
	}

	return t.dbrw.Delete(k)
}
