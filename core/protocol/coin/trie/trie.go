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
	"github.com/multiformats/go-multihash"
	"github.com/pkg/errors"
	"github.com/stratumn/alice/core/protocol/coin/db"
)

// Opt is an option for a Patricia Merkle Trie.
type Opt func(*Trie)

// OptPrefix sets a prefix for all the database keys.
var OptPrefix = func(prefix []byte) Opt {
	return func(t *Trie) {
		t.prefix = prefix
	}
}

// OptHashCode sets the hash algorithm (see Multihash for codes).
//
// The default value is SHA2_256.
var OptHashCode = func(code uint64) Opt {
	return func(t *Trie) {
		t.hashCode = code
	}
}

// Trie represents a Patricia Merkle Trie.
type Trie struct {
	dbrw db.ReadWriter

	prefix   []byte
	hashCode uint64
}

// New create a new Patricia Merkle Trie.
func New(dbrw db.ReadWriter, opts ...Opt) *Trie {
	t := &Trie{
		dbrw:     dbrw,
		hashCode: multihash.SHA2_256,
	}

	for _, o := range opts {
		o(t)
	}

	return t
}

// Get gets the value of the given key.
func (t *Trie) Get(key []byte) ([]byte, error) {
	node, err := t.dbGet(NewNibs(key, false).Expand())
	if err != nil {
		return nil, err
	}

	switch v := node.(type) {
	case Null:
		return nil, errors.WithStack(db.ErrNotFound)
	case *Branch:
		if len(v.Value) < 1 {
			return nil, errors.WithStack(db.ErrNotFound)
		}

		return v.Value, nil
	case *Leaf:
		if len(v.Value) < 1 {
			return nil, errors.WithStack(db.ErrNotFound)
		}

		return v.Value, nil
	}

	return nil, errors.WithStack(ErrInvalidNodeType)
}

// MerkleRoot returns the hash of the root node. If there an no entries, the
// hash of Null{} is returned.
func (t *Trie) MerkleRoot() (multihash.Multihash, error) {
	root, err := t.dbRoot()
	if err != nil {
		return nil, err
	}

	return t.HashNode(root)
}

// HashNode hashes a node.
func (t *Trie) HashNode(node Node) (multihash.Multihash, error) {
	buf, err := node.MarshalBinary()
	if err != nil {
		return nil, err
	}

	hash, err := multihash.Sum(buf, t.hashCode, -1)

	return hash, errors.WithStack(err)
}

// Proof returns a proof of the value for the given key.
func (t *Trie) Proof(key []byte) (Proof, error) {
	root, err := t.dbRoot()
	if err != nil {
		return nil, err
	}

	nodes, found, err := t.recProof(root, nil, NewNibsWithoutCopy(key, false).Expand())
	if err != nil {
		return nil, err
	}

	if !found {
		return nil, errors.WithStack(db.ErrNotFound)
	}

	return Proof(nodes), nil
}

// recProof goes down the trie to the given key and returns all the nodes
// visited bottom up.
//
//	- prefix is the part of the key visited so far
//	- key is the part of the key left to visit
//	- node is the node corresponding to the prefix
func (t *Trie) recProof(node Node, prefix, key []uint8) ([]Node, bool, error) {
	if _, ok := node.(*Hash); ok {
		// If the node is a hash, load the actual node from the prefix,
		// but also collect the hash node.
		n, err := t.dbGet(prefix)
		if err != nil {
			return nil, false, err
		}

		nodes, found, err := t.recProof(n, prefix, key)
		if err != nil {
			return nil, false, err
		}

		return append(nodes, node), found, nil
	}

	if len(key) < 1 {
		// The end of the key was reached, so the current node is
		// the last one.
		return []Node{node}, true, nil
	}

	switch v := node.(type) {
	case Null:
		return nil, false, nil
	case *Branch:
		// The embedded node will be a hash node if a child already
		// exists, null otherwise.
		child := v.EmbeddedNodes[key[0]]

		switch child.(type) {
		case Null:
			return []Node{node}, false, nil
		case *Hash:
			nodes, found, err := t.recProof(child,
				append(prefix, key[0]),
				key[1:],
			)
			if err != nil {
				return nil, false, err
			}

			return append(nodes, node), found, nil
		}
	case *Leaf:
		return []Node{node}, false, nil
	}

	return nil, false, errors.WithStack(ErrInvalidNodeType)
}

// Put sets the value of the given key. Putting a nil or empty value is the
// same as deleting it.
func (t *Trie) Put(key, value []byte) error {
	if len(value) < 1 {
		return t.Delete(key)
	}

	root, err := t.dbRoot()
	if err != nil {
		return err
	}

	_, err = t.recPut(root, nil, NewNibsWithoutCopy(key, false).Expand(), value)

	return err
}

// recPut recursively puts a value in the trie, creating new nodes if
// necessary.
//
//	- prefix is the part of the key visited so far
//	- key is the part of the key left to visit
//	- node is the node corresponding to the prefix
func (t *Trie) recPut(node Node, prefix, key []uint8, value []byte) (Node, error) {
	if _, ok := node.(*Hash); ok {
		// If the node is a hash, load the actual node from the prefix.
		var err error
		node, err = t.dbGet(prefix)
		if err != nil {
			return Null{}, err
		}
	}

	if len(key) < 1 {
		// The end of the key was reached, so the current node is
		// the one that needs to be updated.
		return t.putNodeValue(node, prefix, key, value)
	}

	// Otherwise this node has to be turned into a branch if not already
	// one or updated.
	return t.putChildValue(node, prefix, key, value)
}

// putNodeValue sets the value of a node.
func (t *Trie) putNodeValue(node Node, prefix, key []uint8, value []byte) (Node, error) {
	switch v := node.(type) {
	case Null:
		node = &Leaf{Value: value}
	case *Branch:
		v.Value = value
		node = v
	case *Leaf:
		v.Value = value
		node = v
	default:
		return Null{}, errors.WithStack(ErrInvalidNodeType)
	}

	// Save the node to the database.
	if err := t.dbPut(prefix, node); err != nil {
		return Null{}, err
	}

	return node, nil
}

// putChildValue upgrade a node to a branch if needed and sets the value a
// child.
func (t *Trie) putChildValue(node Node, prefix, key []uint8, value []byte) (Node, error) {
	var branch *Branch

	switch v := node.(type) {
	case Null:
		branch = NewEmptyBranch()
	case *Branch:
		branch = v
	case *Leaf:
		// Upgrade leaf to a branch.
		branch = NewEmptyBranch()
		branch.Value = v.Value
	default:
		return Null{}, errors.WithStack(ErrInvalidNodeType)
	}

	// The embedded node will be a hash node if a child already
	// exists, null otherwise.
	child := branch.EmbeddedNodes[key[0]]

	child, err := t.recPut(
		child,
		append(prefix, key[0]),
		key[1:],
		value,
	)
	if err != nil {
		return Null{}, err
	}

	// Store the new hash of the child.
	hash, err := t.HashNode(child)
	if err != nil {
		return nil, err
	}

	branch.EmbeddedNodes[key[0]] = &Hash{hash}

	// Save the node to the database.
	if err := t.dbPut(prefix, branch); err != nil {
		return Null{}, err
	}

	return branch, nil
}

// Delete removes the value for the given key. Deleting a non-existing key is
// a NOP.
func (t *Trie) Delete(key []byte) error {
	root, err := t.dbRoot()
	if err != nil {
		return err
	}

	_, err = t.recDelete(root, nil, NewNibsWithoutCopy(key, false).Expand())

	return err
}

// recDelete recursively removes a value in the trie, deleting nodes if
// needed.
//
//	- prefix is the part of the key visited so far
//	- key is the part of the key left to visit
//	- node is the node corresponding to the prefix
func (t *Trie) recDelete(node Node, prefix, key []uint8) (Node, error) {
	switch node.(type) {
	case Null:
		// NOP.
		return Null{}, nil
	case *Hash:
		// If the node is a hash, load the actual node from the prefix.
		var err error
		node, err = t.dbGet(prefix)
		if err != nil {
			return Null{}, err
		}
	}

	if len(key) < 1 {
		return t.deleteNodeVal(node, prefix, key)
	}

	// We're not a the end of the key yet, so we have to update the branch.
	return t.deleteChildVal(node, prefix, key)
}

// deleteNodeVal deletes a value from a node, which may delete the node.
func (t *Trie) deleteNodeVal(node Node, prefix, key []uint8) (Node, error) {
	switch v := node.(type) {
	case *Branch:
		v.Value = nil
		if err := t.dbPut(prefix, v); err != nil {
			return Null{}, err
		}

		return Null{}, nil
	case *Leaf:
		if err := t.dbDelete(prefix); err != nil {
			return Null{}, err
		}

		return Null{}, nil
	}

	return Null{}, errors.WithStack(ErrInvalidNodeType)
}

// deleteChildVal removes a value from a child node, delete node if needed.
func (t *Trie) deleteChildVal(node Node, prefix, key []uint8) (Node, error) {
	switch v := node.(type) {
	case *Branch:
		// The embedded node will be a hash node if a child already
		// exists, null otherwise.
		child := v.EmbeddedNodes[key[0]]

		child, err := t.recDelete(
			child,
			append(prefix, key[0]),
			key[1:],
		)
		if err != nil {
			return Null{}, err
		}

		switch child.(type) {
		case Null:
			v.EmbeddedNodes[key[0]] = Null{}
		default:
			hash, err := t.HashNode(child)
			if err != nil {
				return nil, err
			}

			v.EmbeddedNodes[key[0]] = &Hash{hash}
		}

		// Convert to leaf if needed.
		stillBranch := false
		for _, child := range v.EmbeddedNodes {
			if _, null := child.(Null); !null {
				stillBranch = true
			}
		}

		if stillBranch {
			node = v
		} else {
			if len(v.Value) < 1 {
				if err := t.dbDelete(prefix); err != nil {
					return nil, err
				}

				return Null{}, nil
			}

			node = &Leaf{Value: v.Value}
		}

		if err := t.dbPut(prefix, node); err != nil {
			return nil, err
		}

		return node, nil
	case *Leaf:
		return Null{}, nil
	}

	return Null{}, errors.WithStack(ErrInvalidNodeType)
}

// dbRoot returns the dbRoot node from the database.
func (t *Trie) dbRoot() (Node, error) {
	return t.dbGet(nil)
}

// dbGet gets a node in the database from its key.
func (t *Trie) dbGet(key []uint8) (Node, error) {
	k, err := t.dbKey(key)
	if err != nil {
		return Null{}, err
	}

	buf, err := t.dbrw.Get(k)
	if err != nil {
		if errors.Cause(err) == db.ErrNotFound {
			return Null{}, nil
		}

		return nil, err
	}

	n, _, err := UnmarshalNode(buf)

	return n, err
}

// dbPut inserts a node in the database given its key.
func (t *Trie) dbPut(key []uint8, node Node) error {
	k, err := Path(NewNibsFromNibs(key...)).MarshalBinary()
	if err != nil {
		return err
	}

	buf, err := node.MarshalBinary()
	if err != nil {
		return err
	}

	return t.dbrw.Put(k, buf)
}

// dbDelete removes a node from the database given its key.
func (t *Trie) dbDelete(key []uint8) error {
	k, err := Path(NewNibsFromNibs(key...)).MarshalBinary()
	if err != nil {
		return err
	}

	return t.dbrw.Delete(k)
}

// dbKey returns the key of a node given its key in the trie.
func (t *Trie) dbKey(key []uint8) ([]byte, error) {
	k, err := Path(NewNibsFromNibs(key...)).MarshalBinary()
	if err != nil {
		return nil, err
	}

	return append(t.prefix, k...), nil
}
