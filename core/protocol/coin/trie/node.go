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

	"github.com/multiformats/go-multihash"
	"github.com/pkg/errors"
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
	NodeTypeBranch
	NodeTypeLeaf
	NodeTypeEdge
)

// String returns a string representation of a node type.
func (n NodeType) String() string {
	switch n {
	case NodeTypeNull:
		return "<null>"
	case NodeTypeBranch:
		return "<branch>"
	case NodeTypeLeaf:
		return "<leaf>"
	case NodeTypeEdge:
		return "<edge>"
	}

	return "<invalid>"
}

// Node represents a node in a Patricia Merkle Trie.
//
// A node has a compact binary encoding.
//
// Node Encoding
//
//	Null node
//	---------
//	00000000
//
//	Branch node
//	-----------
//	00000001  10100100               1001010010 ...  0100110101...
//	type      value len (uvarint64)  value           16 embedded nodes of
//	                                                 type Edge
//
//	Leaf node
//	---------
//	00000010  10100100               1001010010 ...
//	type      value len (uvarint64)  value
//
//	Edge node
//	---------
//	00000011  11001101 ...  11001101 ...
//	type      multihash     path (see Path for encoding)
//
// Multihash
//
//	10010111          11100011                10101010 ...
//	algo (uvarint64)  digest len (uvarint64)  digest
type Node interface {
	MarshalBinary() ([]byte, error)
	String() string
}

// Null is an empty node.
type Null struct{}

// MarshalBinary marshals the node.
func (n Null) MarshalBinary() ([]byte, error) {
	return []byte{byte(NodeTypeNull) << 4}, nil
}

// String returns a string representation of the node.
func (n Null) String() string {
	return NodeTypeNull.String()
}

// Branch is a node that has children.
type Branch struct {
	Value         []byte
	EmbeddedNodes [16]Node
}

// NewEmptyBranch branch creates a new branch with all embedded nodes set to
// Null.
func NewEmptyBranch() *Branch {
	return &Branch{
		EmbeddedNodes: [...]Node{
			Null{}, Null{}, Null{}, Null{},
			Null{}, Null{}, Null{}, Null{},
			Null{}, Null{}, Null{}, Null{},
			Null{}, Null{}, Null{}, Null{},
		},
	}
}

// MarshalBinary marshals the node.
func (n *Branch) MarshalBinary() ([]byte, error) {
	headroom := 0
	for _, e := range n.EmbeddedNodes {
		switch v := e.(type) {
		case Null:
			headroom++
		case *Edge:
			headroom += 1 + len(v.Hash) + binary.MaxVarintLen32 + (len(v.Path)+1)/2
		}
	}

	buf := marshalValueNode(NodeTypeBranch, n.Value, headroom)

	for _, child := range n.EmbeddedNodes {
		b, err := child.MarshalBinary()
		if err != nil {
			return nil, err
		}

		buf = append(buf, b...)
	}

	return buf, nil
}

// String returns a string representation of the node.
func (n *Branch) String() string {
	return fmt.Sprintf("%v %x %v", NodeTypeBranch, n.Value, n.EmbeddedNodes)
}

// Leaf has no children.
type Leaf struct {
	Value []byte
}

// MarshalBinary marshals the node.
func (n *Leaf) MarshalBinary() ([]byte, error) {
	return marshalValueNode(NodeTypeLeaf, n.Value, 0), nil
}

// String returns a string representation of the node.
func (n *Leaf) String() string {
	return fmt.Sprintf("%v %x", NodeTypeLeaf, n.Value)
}

// Edge contains a partial path to another node and the hash of the target
// node.
type Edge struct {
	Path []uint8
	Hash multihash.Multihash
}

// MarshalBinary marshals the node.
func (n *Edge) MarshalBinary() ([]byte, error) {
	path := NewNibsFromNibs(n.Path...)

	buf := make([]byte, 1+binary.MaxVarintLen32+path.ByteLen()+len(n.Hash))
	buf[0] = byte(NodeTypeEdge)

	pathLen, err := Path(path).MarshalInto(buf[1:])
	if err != nil {
		return nil, err
	}

	copy(buf[1+pathLen:], n.Hash)

	return buf[:1+pathLen+len(n.Hash)], nil
}

// String returns a string representation of the node.
func (n *Edge) String() string {
	return fmt.Sprintf(
		"%v %v %s",
		NodeTypeEdge,
		NewNibsFromNibs(n.Path...),
		n.Hash.B58String(),
	)
}

// UnmarshalNode unmarshals a node. It returns a node and the number of bytes
// read if no error occured.
func UnmarshalNode(buf []byte) (Node, int, error) {
	if len(buf) < 1 {
		return nil, 0, errors.WithStack(ErrInvalidNodeLen)
	}

	// Get node type (the first byte).
	typ := NodeType(buf[0])
	if typ == NodeTypeNull {
		return Null{}, 1, nil
	}

	buf = buf[1:]
	read := 1

	if typ == NodeTypeEdge {
		// Edge type is followed by a path and a multihash.
		path, pathLen, err := UnmarshalPath(buf)
		if err != nil {
			return nil, 0, err
		}

		buf = buf[pathLen:]
		read += pathLen

		hashLen := MultihashLen(buf)
		if hashLen <= 0 || len(buf) < hashLen {
			return nil, 0, errors.WithStack(ErrInvalidNodeLen)
		}

		// Copy the buffer.
		hash := make(multihash.Multihash, hashLen)
		copy(hash, buf[:hashLen])

		return &Edge{
			Path: Nibs(path).Expand(),
			Hash: hash,
		}, read + hashLen, nil
	}

	// Get value part.
	valLen, valRead := binary.Uvarint(buf)
	if read <= 0 || len(buf) < int(valLen)+valRead {
		return nil, 0, errors.WithStack(ErrInvalidNodeLen)
	}

	buf = buf[valRead:]
	val := buf[:valLen]
	buf = buf[valLen:]
	read += valRead + int(valLen)

	switch typ {
	case NodeTypeBranch:
		n := &Branch{Value: val}

		// Get child hash nodes.
		for i := range n.EmbeddedNodes {
			child, childRead, err := UnmarshalNode(buf)
			if err != nil {
				return nil, 0, err
			}

			// They can only be null or edge.
			switch child.(type) {
			case Null, *Edge:
			default:
				return nil, 0, errors.WithStack(ErrInvalidNodeType)
			}

			n.EmbeddedNodes[i] = child
			buf = buf[childRead:]
			read += childRead
		}

		return n, read, nil

	case NodeTypeLeaf:
		return &Leaf{Value: val}, read, nil
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

// marshalValueNode marshals a node type, flags, and a value.
func marshalValueNode(typ NodeType, val []byte, headroom int) []byte {
	buf := make([]byte, 1+binary.MaxVarintLen64+len(val)+headroom)
	buf[0] = byte(typ)

	l := binary.PutUvarint(buf[1:], uint64(len(val)))
	copy(buf[1+l:], val)

	return buf[:1+l+len(val)]
}
