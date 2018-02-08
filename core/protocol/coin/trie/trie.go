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
// It works similarly to a Radix tree:
//
//	  Ali
//	 /   \
//	ce   en
//	    /  \
//	        s
//
// The values leading to a leaf are concatenated to form the final value, in
// this example, from left to right, "Alice", "Alien", and "Aliens".
//
// A Patricia Merkle Trie is a modified Radix tree that makes it
// cryptographically secure. A parent node contains the hash of each of its
// child node (so one hash per child node, as opposed to one hash for all the
// child nodes like a Merkle Tree). By having all the nodes leading to the
// desired value, you have cryptographic proof that it belongs in the trie.
//
// This particular implementation uses a radix of sixteen, meaning a node can
// have up to sixteen children. To compute the index of a child node, you look
// at the first four bits (0-15) of the value contained in the child node.
//
// Because the index takes four bits and not an entire byte, nodes have parity.
// If a node is odd, the first four less significant bits of the last byte of
// its value are ignored.
package trie

import (
	"github.com/stratumn/alice/core/protocol/coin/db"

	"gx/ipfs/QmZyZDi491cCNTLfAhwcaDii2Kg4pwKRkhqQzURGDvY6ua/go-multihash"
)

const (
	nodePrefix byte = iota
)

// Opt is an option for a Patricia Merkle Trie.
type Opt func(*Trie)

// OptPrefix sets a prefix for all the database keys.
var OptPrefix = func(prefix []byte) Opt {
	return func(t *Trie) {
		t.prefix = prefix
	}
}

// Trie represents a Patricia Merkle Trie.
type Trie struct {
	dbrw db.ReadWriter

	prefix []byte
}

// NewTrie create a new Patricia Merkle Trie.
func NewTrie(dbrw db.ReadWriter, opts ...Opt) *Trie {
	t := &Trie{dbrw: dbrw}

	for _, o := range opts {
		o(t)
	}

	return t
}

// Add adds a value.
func (t *Trie) Add(value []byte) error {
	// So linter doesn't complain.
	t.rootKey()

	return nil
}

// Delete removes a value. Deleting a non-existing value is a NOP.
func (t *Trie) Delete(value []byte) error {
	return nil
}

// MerkleRoot returns the hash of the root node.
func (t *Trie) MerkleRoot() (multihash.Multihash, error) {
	return nil, nil
}

// Proof returns a proof that the Patricia Merkle Trie contains the value.
func (t *Trie) Proof(value []byte) (Proof, error) {
	return nil, nil
}

// rootKey returns the key of the root node.
func (t *Trie) rootKey() []byte {
	return t.nodeKey(nil)
}

// nodeKey returns the key of a node given its path in the trie.
func (t *Trie) nodeKey(path []byte) []byte {
	return append(append(t.prefix, nodePrefix), path...)
}
