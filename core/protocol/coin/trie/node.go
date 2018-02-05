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
	"encoding/binary"
	"fmt"

	"github.com/pkg/errors"

	"gx/ipfs/QmZyZDi491cCNTLfAhwcaDii2Kg4pwKRkhqQzURGDvY6ua/go-multihash"
)

var (
	// ErrInvalidNodeType is returned when a node has an invalid type.
	ErrInvalidNodeType = errors.New("the node has an invalid type")

	// ErrInvalidNodeLen is returned when an encoded node has an invalid
	// number of bytes.
	ErrInvalidNodeLen = errors.New("the encoded node has an invalid number of bytes")
)

// NodeType represents the type of a node.
type NodeType byte

// Node types.
const (
	NodeTypeNull NodeType = iota
	NodeTypeParent
	NodeTypeLeaf
	NodeTypeHash
)

// String returns a string representation of a node type.
func (n NodeType) String() string {
	switch n {
	case NodeTypeNull:
		return "<null>"
	case NodeTypeParent:
		return "<parent>"
	case NodeTypeLeaf:
		return "<leaf>"
	case NodeTypeHash:
		return "<hash>"
	}

	return "<invalid>"
}

// NodeFlag represents a node flag.
type NodeFlag byte

// Node flags.
const (
	NodeFlagOdd NodeFlag = 1 << iota
)

// Node represents a node in a Patricia Merkle Trie.
//
// A node has a compact binary encoding.
//
// Node Encoding
//
//	Null node
//	---------
//	0000  0000
//	type  flags
//
//	Parent node
//	-----------
//	0001  0001   10100                1001010010...  0100110101...
//	type  flags  value len (uvarint)  value          16 child nodes
//
//	Leaf node
//	---------
//	0010  0001   10100                1001010010...
//	type  flags  value len (uvarint)  value
//
//	Hash node
//	---------
//	0011  0000   0100110101...
//	type  flags  multihash
//
// Flags
//
//	0001 odd
//	--------
//	An odd node means that the first four less significant bits of the last
//	byte of the value should be ignored.
//
// Multihash
//
//	0001111         0110101               1010101011...
//	algo (uvarint)  digest len (uvarint)  digest
type Node interface {
	Marshal() []byte
	String() string
}

// NullNode is an empty node.
type NullNode struct{}

// Marshal marshals the node.
func (n NullNode) Marshal() []byte {
	return []byte{byte(NodeTypeNull)}
}

// String returns a string representation of the node.
func (n NullNode) String() string {
	return NodeTypeNull.String()
}

// ParentNode is a node that has children.
type ParentNode struct {
	IsOdd    bool
	Children []Node
	Value    []byte
}

// Marshal marshals the node.
func (n ParentNode) Marshal() []byte {
	buf := marshalValueNode(NodeTypeParent, n.IsOdd, n.Value)

	for _, child := range n.Children {
		buf = append(buf, child.Marshal()...)
	}

	return buf
}

// String returns a string representation of the node.
func (n ParentNode) String() string {
	return fmt.Sprintf("%v %v %x %v", NodeTypeParent, n.IsOdd, n.Value, n.Children)
}

// LeafNode has no children.
type LeafNode struct {
	IsOdd bool
	Value []byte
}

// Marshal marshals the node.
func (n LeafNode) Marshal() []byte {
	return marshalValueNode(NodeTypeLeaf, n.IsOdd, n.Value)
}

// String returns a string representation of the node.
func (n LeafNode) String() string {
	return fmt.Sprintf("%v %v %x", NodeTypeLeaf, n.IsOdd, n.Value)
}

// HashNode contains the hash of another node.
type HashNode struct {
	Hash multihash.Multihash
}

// Marshal marshals the node.
func (n HashNode) Marshal() []byte {
	return append([]byte{byte(NodeTypeHash << 4)}, n.Hash...)
}

// String returns a string representation of the node.
func (n HashNode) String() string {
	return fmt.Sprintf("%v %sx", NodeTypeHash, n.Hash.B58String())
}

// UnmarshalNode unmarshals a node. It returns a node and the number of bytes
// read if no error occured.
func UnmarshalNode(buf []byte) (Node, int, error) {
	if len(buf) < 1 {
		return nil, 0, errors.WithStack(ErrInvalidNodeLen)
	}

	// Get node type (four most significant bits).
	typ := NodeType(buf[0] >> 4)
	if typ == NodeTypeNull {
		return NullNode{}, 1, nil
	}

	// Get flags (next four most significant bits).
	flags := buf[0] & 0x0F
	isOdd := flags&byte(NodeFlagOdd) > 0

	buf = buf[1:]
	read := 1

	// Hash type is a little different.
	if typ == NodeTypeHash {
		// It is simply followed by the multihash.
		hashLen := MultihashLen(buf)
		if hashLen <= 0 || len(buf) < hashLen {
			return nil, 0, errors.WithStack(ErrInvalidNodeLen)
		}

		return HashNode{
			Hash: multihash.Multihash(buf[:hashLen]),
		}, read + hashLen, nil
	}

	// Get value/hash part.
	valLen, valRead := binary.Uvarint(buf)
	if read <= 0 || len(buf) < int(valLen)+valRead {
		return nil, 0, errors.WithStack(ErrInvalidNodeLen)
	}

	buf = buf[valRead:]
	val := buf[:valLen]
	buf = buf[valLen:]
	read += valRead + int(valLen)

	switch typ {
	case NodeTypeParent:
		n := ParentNode{
			IsOdd: isOdd,
			Value: val,
		}

		// Get child nodes.
		for i := 0; i < 16; i++ {
			child, childRead, err := UnmarshalNode(buf)
			if err != nil {
				return nil, 0, err
			}

			// Child nodes can only be null or hash.
			switch child.(type) {
			case NullNode, HashNode:
			default:
				return nil, 0, ErrInvalidNodeType
			}

			n.Children = append(n.Children, child)
			buf = buf[childRead:]
			read += childRead
		}

		return n, read, nil

	case NodeTypeLeaf:
		return LeafNode{
			IsOdd: isOdd,
			Value: val,
		}, read, nil
	}

	return nil, 0, errors.WithStack(ErrInvalidNodeType)
}

// MultihashLen returns the length of a multihash.
func MultihashLen(buf []byte) int {
	_, n1 := binary.Uvarint(buf)
	if n1 <= 0 {
		return 0
	}

	digestLen, n2 := binary.Uvarint(buf[n1:])
	if n2 <= 0 {
		return 0
	}

	return n1 + n2 + int(digestLen)
}

// marshalValueNode marshals a node type, parity flag, and a value.
func marshalValueNode(typ NodeType, isOdd bool, val []byte) []byte {
	buf := make([]byte, 1+binary.MaxVarintLen64+len(val))
	buf[0] = byte(typ) << 4

	if isOdd {
		buf[0] |= byte(NodeFlagOdd)
	}

	l := binary.PutUvarint(buf[1:], uint64(len(val)))
	copy(buf[1+l:], val)

	return buf[:1+l+len(val)]
}
