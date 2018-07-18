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
// all the information about the state in a Patricia Merkle Trie, the Merkle
// Root of the trie uniquely describes the current state of the application.
// Extracting a proof (the nodes leading to a value) gives you cryptographic
// evidence that the value is part of the current state. By adding a Merkle
// Root of the state to block headers, you only need the last block header and
// a Merkle Proof to prove the value of an account. It makes it very easy for a
// light node to easily sync up the  accounts it is tracking. Without it it
// would need much more data to know without a doubt the value of an account.
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
// As opposed to a tree, in a trie a value gives you its position the tree, and
// vice-versa. So a value will always have the same position in the trie. You
// can think of a value as a path in the trie. For instance, the path of the
// left leaf in the example is [A l i c e], giving you the value "Alice". This
// property of tries make them canonical -- a trie containing the same values
// will always be the same.
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
// avoid confusion, we call the path of a node the Key (not a value like
// previously), and the data it stores the Value, so it becomes more similar to
// a key-value database. In the previous example you could set an arbitrary
// value for the key 0x3e20f. More concretely, you could use the ID of an
// account as its key and the serialized account as its value, so in the
// previous example you could store the account whose ID is 0x3e02f in the left
// leaf.
//
// In practice with arbitrary data you would end up with long series of
// branches having a single child, which is inefficient. So, like Ethereum,
// we introduce another type of node which contains a partial path that can be
// used to skip such nodes. For instance:
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
// Can be represented by a single node of type edge containing the path between
// the two nodes:
//
//	  3 [Edge: e0bc]
//
// A Patricia Merkle Trie is a modified Radix trie that makes it
// cryptographically secure. A branch contains the hash of each of its child
// nodes (so one hash per child node, as opposed to one hash for all the
// child nodes like a Merkle Tree). The Merkle root is the hash of the root
// node. By having all the nodes leading to the desired value, you have
// cryptographic proof that it belongs in the trie.
//
// Put, Delete, and Proof are O(k), where k is the length of the key. In
// practice they are faster because there will be fewer edges.
//
// In this implementation, Get is O(1) because a key directly maps to its
// value in the database.
package trie

// Dev Notes
//
// By convention, keys are defined in []byte whereas []uint8 represents a
// slice of nibble (0-15).
//
// Nodes are pointers (except null). For performance reasons, the cache doesn't
// clone nodes, so be careful when mutating them. In general, always clone an
// externally created node before mutating it.
