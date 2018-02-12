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

	"github.com/multiformats/go-multihash"
	"github.com/pkg/errors"
	"github.com/stratumn/alice/core/protocol/coin/db"
)

// Opt is an option for a Patricia Merkle Trie.
type Opt func(*Trie)

// OptDB sets the database. By default it uses an in-memory map.
var OptDB = func(db db.ReadWriter) Opt {
	return func(t *Trie) {
		t.dbrw = db
	}
}

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
func New(opts ...Opt) *Trie {
	t := &Trie{
		dbrw:     newMapDB(),
		hashCode: multihash.SHA2_256,
	}

	for _, o := range opts {
		o(t)
	}

	return t
}

// Get gets the value of the given key.
func (t *Trie) Get(key []byte) ([]byte, error) {
	node, err := t.getNode(NewNibs(key, false).Expand())
	if err != nil {
		return nil, err
	}

	switch node := node.(type) {
	case Null:
		return nil, errors.WithStack(db.ErrNotFound)

	case *Branch:
		if len(node.Value) < 1 {
			return nil, errors.WithStack(db.ErrNotFound)
		}

		return node.Value, nil

	case *Leaf:
		if len(node.Value) < 1 {
			return nil, errors.WithStack(db.ErrNotFound)
		}

		return node.Value, nil
	}

	return nil, errors.WithStack(ErrInvalidNodeType)
}

// MerkleRoot returns the hash of the root node. If there an no entries, the
// hash of Null{} is returned.
func (t *Trie) MerkleRoot() (multihash.Multihash, error) {
	root, err := t.rootNode()
	if err != nil {
		return nil, err
	}

	return t.hashNode(root)
}

// Proof returns a proof of the value for the given key.
func (t *Trie) Proof(key []byte) (Proof, error) {
	root, err := t.rootNode()
	if err != nil {
		return nil, err
	}

	key = NewNibsWithoutCopy(key, false).Expand()

	nodes, found, err := t.recProof(root, nil, key)
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
	if edge, ok := node.(*Edge); ok {
		// If the node is an edge, load the actual node from the prefix
		// and the path but also collect the edge node.
		if !bytes.HasPrefix(key, edge.Path) {
			return nil, false, nil
		}

		key = key[len(edge.Path):]
		prefix = append(prefix, edge.Path...)

		n, err := t.getNode(prefix)
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

	switch node := node.(type) {
	case Null:
		return nil, false, nil
	case *Branch:
		// The embedded node will be an edge node if a child already
		// exists, null otherwise.
		edge := node.EmbeddedNodes[key[0]]

		switch edge.(type) {
		case Null:
			return []Node{node}, false, nil
		case *Edge:
			nodes, found, err := t.recProof(edge, prefix, key)
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

	root, err := t.rootNode()
	if err != nil {
		return err
	}

	key = NewNibsWithoutCopy(key, false).Expand()

	_, _, err = t.recPut(root, nil, key, value)

	return err
}

// recPut recursively puts a value in the trie, creating new nodes if
// necessary.
//
//	- prefix is the part of the key visited so far
//	- key is the part of the key left to visit
//	- node is the node corresponding to the prefix
//
// It returns the updated node and its key.
func (t *Trie) recPut(node Node, prefix, key []uint8, value []byte) (Node, []uint8, error) {
	if edge, ok := node.(*Edge); ok {
		// Handle edge cases.
		i := 0
		for ; i < len(key) && i < len(edge.Path); i++ {
			if key[i] != edge.Path[i] {
				break
			}
		}

		if i < len(key) && i < len(edge.Path) {
			// Fork.
			//
			// \        \
			//  \        C
			//   \      / \
			//    A    A   B
			return t.fork(edge, prefix, key, value, i)
		}

		if i < len(edge.Path) {
			// Split.
			//
			// \       \
			//  \       B
			//   \       \
			//    A       A
			return t.split(edge, prefix, key, value, i)
		}

		// Other cases can be handles normally.
		prefix = append(prefix, edge.Path...)
		key = key[len(edge.Path):]

		var err error

		node, err = t.getNode(prefix)
		if err != nil {
			return Null{}, nil, err
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
func (t *Trie) putNodeValue(node Node, prefix, key []uint8, value []byte) (Node, []uint8, error) {
	switch n := node.(type) {
	case Null:
		node = &Leaf{Value: value}
	case *Branch:
		n.Value = value
		node = n
	case *Leaf:
		n.Value = value
		node = n
	default:
		return Null{}, nil, errors.WithStack(ErrInvalidNodeType)
	}

	if err := t.putNode(prefix, node); err != nil {
		return Null{}, nil, err
	}

	return node, prefix, nil
}

// putChildValue upgrades a node to a branch if needed and sets the value a
// child.
func (t *Trie) putChildValue(node Node, prefix, key []uint8, value []byte) (Node, []uint8, error) {
	var branch *Branch

	switch node := node.(type) {
	case Null:
		branch = NewEmptyBranch()
	case *Branch:
		branch = node
	case *Leaf:
		// Upgrade leaf to a branch.
		branch = NewEmptyBranch()
		branch.Value = node.Value
	default:
		return Null{}, nil, errors.WithStack(ErrInvalidNodeType)
	}

	// The embedded node will be an edge node if a child already exists,
	// null otherwise.
	edge := branch.EmbeddedNodes[key[0]]

	if _, null := edge.(Null); null {
		edge = &Edge{Path: key}
	}

	node, path, err := t.recPut(edge, prefix, key, value)
	if err != nil {
		return Null{}, nil, err
	}

	// Store the new hash of the child.
	hash, err := t.hashNode(node)
	if err != nil {
		return nil, nil, err
	}

	branch.EmbeddedNodes[key[0]] = &Edge{Path: path[len(prefix):], Hash: hash}

	if err := t.putNode(prefix, branch); err != nil {
		return Null{}, nil, err
	}

	return branch, prefix, nil
}

// fork splits forks an edge.
func (t *Trie) fork(edge *Edge, prefix, key []uint8, value []byte, i int) (Node, []uint8, error) {
	branch := NewEmptyBranch()

	// Existing node.
	branch.EmbeddedNodes[edge.Path[i]] = &Edge{
		Path: edge.Path[i:],
		Hash: edge.Hash,
	}

	// New node.
	child, _, err := t.putNodeValue(
		Null{},
		append(prefix, key...),
		nil,
		value,
	)
	if err != nil {
		return Null{}, nil, err
	}

	hash, err := t.hashNode(child)
	if err != nil {
		return Null{}, nil, err
	}

	branch.EmbeddedNodes[key[i]] = &Edge{
		Path: key[i:],
		Hash: hash,
	}

	// Save forking branch.
	path := append(prefix, key[:i]...)
	if err := t.putNode(path, branch); err != nil {
		return Null{}, nil, err
	}

	return branch, path, nil
}

// split splits an edge in two.
func (t *Trie) split(edge *Edge, prefix, key []uint8, value []byte, i int) (Node, []uint8, error) {
	branch := NewEmptyBranch()
	branch.Value = value

	branch.EmbeddedNodes[edge.Path[i]] = &Edge{
		Path: edge.Path[i:],
		Hash: edge.Hash,
	}

	path := append(prefix, key[:i]...)
	if err := t.putNode(path, branch); err != nil {
		return Null{}, nil, err
	}

	return branch, path, nil
}

// Delete removes the value for the given key. Deleting a non-existing key is
// a NOP.
func (t *Trie) Delete(key []byte) error {
	root, err := t.rootNode()
	if err != nil {
		return err
	}

	key = NewNibsWithoutCopy(key, false).Expand()

	_, _, _, err = t.recDelete(root, nil, key)

	return err
}

// recDelete recursively removes a value in the trie, deleting nodes if
// needed.
//
//	- prefix is the part of the key visited so far
//	- key is the part of the key left to visit
//	- node is the node corresponding to the prefix
//
// It returns the updated node, its path, and whether the key was found.
func (t *Trie) recDelete(node Node, prefix, key []uint8) (Node, []uint8, bool, error) {
	var err error

	if edge, ok := node.(*Edge); ok {
		key = key[len(edge.Path):]
		prefix = append(prefix, edge.Path...)

		node, err = t.getNode(prefix)
		if err != nil {
			return Null{}, nil, false, err
		}
	}

	if len(key) < 1 {
		return t.deleteNodeVal(node, prefix, key)
	}

	// We're not a the end of the key yet, so we have to update the branch.
	return t.deleteChildVal(node, prefix, key)
}

// deleteNodeVal deletes a value from a node, which may delete nodes along the
// ways.
func (t *Trie) deleteNodeVal(node Node, prefix, key []uint8) (Node, []uint8, bool, error) {
	switch node := node.(type) {
	case Null:
		// NOP.
		return Null{}, nil, false, nil

	case *Branch:
		if len(prefix) > 0 {
			var lastEdge *Edge
			numEdges := 0

			for _, n := range node.EmbeddedNodes {
				if edge, ok := n.(*Edge); ok {
					lastEdge = edge
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
					return Null{}, nil, false, err
				}

				childPrefix := append(prefix, lastEdge.Path...)

				child, err := t.getNode(childPrefix)
				if err != nil {
					return Null{}, nil, false, err
				}

				return child, childPrefix, true, nil
			}
		}

		node.Value = nil
		if err := t.putNode(prefix, node); err != nil {
			return Null{}, nil, false, err
		}

		return node, prefix, true, nil

	case *Leaf:
		if err := t.deleteNode(prefix); err != nil {
			return Null{}, nil, false, err
		}

		return Null{}, nil, true, nil
	}

	return Null{}, nil, false, errors.WithStack(ErrInvalidNodeType)
}

// deleteChildVal removes a value from a child node.
func (t *Trie) deleteChildVal(node Node, prefix, key []uint8) (Node, []uint8, bool, error) {
	switch node := node.(type) {
	case Null, *Leaf:
		// NOP.
		return Null{}, nil, false, nil

	case *Branch:
		edge := node.EmbeddedNodes[key[0]]

		switch edge := edge.(type) {
		case Null:
			// NOP.
			return Null{}, nil, false, nil

		case *Edge:
			if !bytes.HasPrefix(key, edge.Path) {
				// NOP.
				return Null{}, nil, false, nil
			}

			child, childPrefix, found, err := t.recDelete(edge, prefix, key)
			if err != nil {
				return Null{}, nil, false, err
			}

			if !found {
				// NOP.
				return Null{}, nil, false, nil
			}

			// Recompute edges and hash children.
			switch child := child.(type) {
			case Null:
				node.EmbeddedNodes[key[0]] = Null{}

			case *Branch:
				hash, err := t.hashNode(child)
				if err != nil {
					return nil, nil, false, err
				}

				node.EmbeddedNodes[key[0]] = &Edge{
					Path: childPrefix[len(prefix):],
					Hash: hash,
				}

			case *Leaf:
				hash, err := t.hashNode(child)
				if err != nil {
					return nil, nil, false, err
				}

				node.EmbeddedNodes[key[0]] = &Edge{
					Path: childPrefix[len(prefix):],
					Hash: hash,
				}

			default:
				return Null{}, nil, false, errors.WithStack(ErrInvalidNodeType)
			}

			return t.restructureBranch(node, prefix, key)
		}
	}

	return Null{}, nil, false, errors.WithStack(ErrInvalidNodeType)
}

// restrucutreBranch restructures a branch after on of its child node was
// deleted.
func (t *Trie) restructureBranch(branch *Branch, prefix, key []uint8) (Node, []uint8, bool, error) {
	var lastEdge *Edge
	numEdges := 0

	for _, n := range branch.EmbeddedNodes {
		if edge, ok := n.(*Edge); ok {
			lastEdge = edge
			numEdges++
		}
	}

	if numEdges == 0 {
		// No more children.

		if len(branch.Value) > 0 {
			// The node has a value, so we can't delete it but we
			// can downgrade it to a leaf.
			leaf := &Leaf{Value: branch.Value}

			if err := t.putNode(prefix, leaf); err != nil {
				return Null{}, nil, false, err
			}

			return leaf, prefix, true, nil
		}

		// No value stored so we can delete the node.
		if err := t.deleteNode(prefix); err != nil {
			return Null{}, nil, false, err
		}

		return Null{}, nil, true, nil
	}

	if numEdges == 1 && len(branch.Value) <= 0 && len(prefix) > 0 {
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
			return Null{}, nil, false, err
		}

		childPrefix := append(prefix, lastEdge.Path...)

		child, err := t.getNode(childPrefix)
		if err != nil {
			return Null{}, nil, false, err
		}

		return child, childPrefix, true, nil
	}

	if err := t.putNode(prefix, branch); err != nil {
		return Null{}, nil, false, err
	}

	return branch, prefix, true, nil
}

// rootNode returns the root node.
func (t *Trie) rootNode() (Node, error) {
	return t.getNode(nil)
}

// getNode gets a node from its key.
func (t *Trie) getNode(key []uint8) (Node, error) {
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

// putNode inserts the node with the given key.
func (t *Trie) putNode(key []uint8, node Node) error {
	k, err := t.dbKey(key)
	if err != nil {
		return err
	}

	buf, err := node.MarshalBinary()
	if err != nil {
		return err
	}

	return t.dbrw.Put(k, buf)
}

// deleteNode removes the node with the given key.
func (t *Trie) deleteNode(key []uint8) error {
	k, err := t.dbKey(key)
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

// hashNode hashes a node.
func (t *Trie) hashNode(node Node) (multihash.Multihash, error) {
	buf, err := node.MarshalBinary()
	if err != nil {
		return nil, err
	}

	hash, err := multihash.Sum(buf, t.hashCode, -1)

	return hash, errors.WithStack(err)
}
