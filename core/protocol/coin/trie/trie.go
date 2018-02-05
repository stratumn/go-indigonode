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

// Package trie implements a Patricia Merkle Trie.
//
// A Patricia Merkle Trie can be used to cryptographically prove that a value
// is part of a set by only showing the nodes leading to that value.
//
// It is similar to a Merkle Tree but it can easily and efficiently remove
// values from the set. It is also canonical, meaning that all tries containing
// the same values will be the same no matter which order the values were added
// or removed.
//
// In the context of a blockchain, it is useful for light clients. By keeping
// all the information about the state (for example a hash of every account) in
// a Patricia Merkle Trie, the Merkle Root of the trie uniquely describes the
// current state of the application. Extracting a proof (the nodes leading to
// a value) gives you cryptographic evidence that the value is part of the
// current state. By adding a Merkle Root of the state to block headers, you
// only need the last block header and a Merkle Proof to prove the balance of
// an account. It makes it very easy for a light node to easily sync up the
// balance of the accounts it is tracking. Without it it would need much more
// data to know without a doubt the balance of an account.
//
// It works similarly to a Radix trie:
//
//	  A
//	   \
//	    l
//	     \
//	      i
//	     / \
//	    c   e
//	   /     \
//	  e       n
//
// As opposed to a tree, in a trie a value gives you its position the tree. So
// a value will always have the same position in the trie. You can think of a
// value as a path in the trie. For instance, the path of the left leaf in the
// example is [A l i c e], giving you the value "Alice". This property of tries
// make them deterministic -- a trie containing the same values will always
// be the same.
//
// In this example we are using strings, so each branch can have as many
// children as the number of allowed characters. This number is said to be
// the radix of the tree.
//
// Our implementation uses a radix of 16, corresponding to four bits of the
// value (a nibble). You can think of a nibble as a single hexadecimal
// character. For instance:
//
//	  3
//	   \
//	    e
//	     \
//	      0
//	     / \
//	    2   b
//	   /     \
//	  f       c
//
// In this example the left leaf represents the value 0x3e02f. To figure out
// the position of a node you can just look at the hexadecimal digits of the
// value.
//
// This implementation allows you to store arbitrary data in the nodes. To
// avoid confusions, we call the path of a node the Key (not a value like
// previously), and the data it stores the Value, so it becomes more similar to
// a key-value database. In the previous example you could set an arbitrary
// value for the key 0x3e20f. More concretely, you could use the ID of an
// account as its key and the serialized account as its value, so in in the
// previous example you could store the account whose ID is 0x3e02f in the left
// leaf.
//
// In practice with arbitrary data you would end up with long series of
// branches having a single child, which is inefficient. So, like Ethereum,
// we introduce another type of node which contains a partial path that can be
// used to skip such nodes [TODO]. For instance:
//
//	  3
//	   \
//	    e
//	     \
//	      0
//	       \
//	        b
//	         \
//	          c
//
// Can be represented by a single node:
//
//	  3 [Path: 0xe0cb]
//
// A Patricia Merkle Trie is a modified Radix trie that makes it
// cryptographically secure. A branch contains the hash of each of its child
// nodes (so one hash per child node, as opposed to one hash for all the
// child nodes like a Merkle Tree). By having all the nodes leading to the
// desired value, you have cryptographic proof that it belongs in the trie.
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
	node, err := t.find(NewNibs(key, false).Expand())
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

// Put sets the value of the given key. Putting a nil or empty value is the
// same as deleting it.
func (t *Trie) Put(key, value []byte) error {
	if len(value) < 1 {
		return t.Delete(key)
	}

	root, err := t.root()
	if err != nil {
		return err
	}

	_, err = t.update(root, nil, NewNibs(key, false).Expand(), value)

	return err
}

// Delete removes the value for the given key. Deleting a non-existing key is
// a NOP.
func (t *Trie) Delete(key []byte) error {
	root, err := t.root()
	if err != nil {
		return err
	}

	_, err = t.remove(root, nil, NewNibs(key, false).Expand())

	return err
}

// MerkleRoot returns the hash of the root node. If there an no entries, the
// hash of Null{} is returned.
func (t *Trie) MerkleRoot() (multihash.Multihash, error) {
	root, err := t.root()
	if err != nil {
		return nil, err
	}

	return t.HashNode(root)
}

// Proof returns a proof of the value for the given key.
func (t *Trie) Proof(key []byte) (Proof, error) {
	root, err := t.root()
	if err != nil {
		return nil, err
	}

	nodes, found, err := t.collect(root, nil, NewNibs(key, false).Expand())
	if err != nil {
		return nil, err
	}

	if !found {
		return nil, errors.WithStack(db.ErrNotFound)
	}

	return Proof(nodes), nil
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

// root returns the root node.
func (t *Trie) root() (Node, error) {
	return t.find(nil)
}

// find finds a node in the database from its key.
func (t *Trie) find(key []uint8) (Node, error) {
	k, err := t.nodeKey(key)
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

// collect goes down the trie to the given key and returns all the nodes
// visited bottom up.
//
//	- prefix is the part of the key visited so far
//	- key is the part of the key left to visit
//	- node is the node corresponding to the prefix
func (t *Trie) collect(node Node, prefix, key []uint8) ([]Node, bool, error) {
	if _, ok := node.(*Hash); ok {
		// If the node is a hash, load the actual node from the prefix,
		// but also collect the hash node.
		n, err := t.find(prefix)
		if err != nil {
			return nil, false, err
		}

		nodes, found, err := t.collect(n, prefix, key)
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
			nodes, found, err := t.collect(child,
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

// update recursively updates a value in the trie, creating a new node if
// necessary.
//
//	- prefix is the part of the key visited so far
//	- key is the part of the key left to visit
//	- node is the node corresponding to the prefix
func (t *Trie) update(node Node, prefix, key []uint8, value []byte) (Node, error) {
	if _, ok := node.(*Hash); ok {
		// If the node is a hash, load the actual node from the prefix.
		var err error
		node, err = t.find(prefix)
		if err != nil {
			return Null{}, err
		}
	}

	if len(key) < 1 {
		// The end of the key was reached, so the current node is
		// the one that needs to be updated.
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
	} else {
		// We're not a the end of the key yet, so we have to make the
		// current node a branch if it isn't already.
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

		child, err := t.update(
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

		node = branch
	}

	// Save the node to the database.
	if err := t.insert(prefix, node); err != nil {
		return Null{}, err
	}

	return node, nil
}

// insert inserts a node in the database given its key.
func (t *Trie) insert(key []uint8, node Node) error {
	k, err := Path(NewNibsFromNibs(key)).MarshalBinary()
	if err != nil {
		return err
	}

	buf, err := node.MarshalBinary()
	if err != nil {
		return err
	}

	return t.dbrw.Put(k, buf)
}

// remove recursively removes a value in the trie, deleting nodes if
// necessary.
//
//	- prefix is the part of the key visited so far
//	- key is the part of the key left to visit
//	- node is the node corresponding to the prefix
func (t *Trie) remove(node Node, prefix, key []uint8) (Node, error) {
	switch node.(type) {
	case Null:
		return Null{}, nil
	case *Hash:
		// If the node is a hash, load the actual node from the prefix.
		var err error
		node, err = t.find(prefix)
		if err != nil {
			return Null{}, err
		}
	}

	if len(key) < 1 {
		// The end of the key was reached, so the current node is
		// the one that needs to be updated.
		switch v := node.(type) {
		case *Branch:
			v.Value = nil
			if err := t.insert(prefix, v); err != nil {
				return Null{}, err
			}

			return Null{}, nil
		case *Leaf:
			if err := t.rem(prefix); err != nil {
				return Null{}, err
			}

			return Null{}, nil
		default:
			return Null{}, errors.WithStack(ErrInvalidNodeType)
		}
	}

	// We're not a the end of the key yet, so we have to update the branch.
	switch v := node.(type) {
	case *Branch:
		// The embedded node will be a hash node if a child already
		// exists, null otherwise.
		child := v.EmbeddedNodes[key[0]]

		child, err := t.remove(
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
				if err := t.rem(prefix); err != nil {
					return nil, err
				}

				return Null{}, nil
			}

			node = &Leaf{Value: v.Value}
		}

		if err := t.insert(prefix, node); err != nil {
			return nil, err
		}

		return node, nil
	case *Leaf:
		return Null{}, nil
	}

	return Null{}, errors.WithStack(ErrInvalidNodeType)
}

// remove removes a node from the database given its key.
func (t *Trie) rem(key []uint8) error {
	k, err := Path(NewNibsFromNibs(key)).MarshalBinary()
	if err != nil {
		return err
	}

	return t.dbrw.Delete(k)
}

// nodeKey returns the key of a node given its key in the trie.
func (t *Trie) nodeKey(key []uint8) ([]byte, error) {
	k, err := Path(NewNibsFromNibs(key)).MarshalBinary()
	if err != nil {
		return nil, err
	}

	return append(t.prefix, k...), nil
}
