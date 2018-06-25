// Copyright © 2017-2018 Stratumn SAS
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

// nodeType represents the type of a node.
type nodeType byte

// Node types.
const (
	nodeTypeNull nodeType = iota
	nodeTypeBranch
	nodeTypeLeaf
	nodeTypeEdge
)

// String returns a string representation of a node type.
func (n nodeType) String() string {
	switch n {
	case nodeTypeNull:
		return "<null>"
	case nodeTypeBranch:
		return "<branch>"
	case nodeTypeLeaf:
		return "<leaf>"
	case nodeTypeEdge:
		return "<edge>"
	}

	return "<invalid>"
}

// node represents a node in a Patricia Merkle Trie.
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
type node interface {
	MarshalBinary() ([]byte, error)
	Clone() node
	String() string
}

// null is an empty node.
type null struct{}

// MarshalBinary marshals the node.
func (n null) MarshalBinary() ([]byte, error) {
	return []byte{byte(nodeTypeNull) << 4}, nil
}

// Clone deep-copies the node.
func (n null) Clone() node {
	return null{}
}

// String returns a string representation of the node.
func (n null) String() string {
	return nodeTypeNull.String()
}

// branch is a node that has children.
type branch struct {
	Value         []byte
	EmbeddedNodes [16]node
}

// newEmptyBranch branch creates a new branch with all embedded nodes set to
// Null.
func newEmptyBranch() *branch {
	return &branch{
		EmbeddedNodes: [...]node{
			null{}, null{}, null{}, null{},
			null{}, null{}, null{}, null{},
			null{}, null{}, null{}, null{},
			null{}, null{}, null{}, null{},
		},
	}
}

// MarshalBinary marshals the node.
func (n *branch) MarshalBinary() ([]byte, error) {
	headroom := 0
	for _, e := range n.EmbeddedNodes {
		switch v := e.(type) {
		case null:
			headroom++
		case *edge:
			headroom += 1 + len(v.Hash) + binary.MaxVarintLen32 + (len(v.Path)+1)/2
		}
	}

	buf := marshalValueNode(nodeTypeBranch, n.Value, headroom)

	for _, child := range n.EmbeddedNodes {
		b, err := child.MarshalBinary()
		if err != nil {
			return nil, err
		}

		buf = append(buf, b...)
	}

	return buf, nil
}

// Clone deep-copies the node.
func (n *branch) Clone() node {
	clone := &branch{
		Value: make([]byte, len(n.Value)),
	}

	copy(clone.Value, n.Value)

	for i, n := range n.EmbeddedNodes {
		clone.EmbeddedNodes[i] = n.Clone()
	}

	return clone
}

// String returns a string representation of the node.
func (n *branch) String() string {
	return fmt.Sprintf("%v %x %v", nodeTypeBranch, n.Value, n.EmbeddedNodes)
}

// leaf has no children.
type leaf struct {
	Value []byte
}

// MarshalBinary marshals the node.
func (n *leaf) MarshalBinary() ([]byte, error) {
	return marshalValueNode(nodeTypeLeaf, n.Value, 0), nil
}

// Clone deep-copies the node.
func (n *leaf) Clone() node {
	clone := &leaf{
		Value: make([]byte, len(n.Value)),
	}

	copy(clone.Value, n.Value)

	return clone
}

// String returns a string representation of the node.
func (n *leaf) String() string {
	return fmt.Sprintf("%v %x", nodeTypeLeaf, n.Value)
}

// edge contains a partial path to another node and the hash of the target
// node.
type edge struct {
	Path []uint8
	Hash multihash.Multihash
}

// MarshalBinary marshals the node.
func (n *edge) MarshalBinary() ([]byte, error) {
	p := newNibsFromNibs(n.Path...)

	buf := make([]byte, 1+binary.MaxVarintLen32+p.ByteLen()+len(n.Hash))
	buf[0] = byte(nodeTypeEdge)

	pathLen, err := path(p).MarshalInto(buf[1:])
	if err != nil {
		return nil, err
	}

	copy(buf[1+pathLen:], n.Hash)

	return buf[:1+pathLen+len(n.Hash)], nil
}

// Clone deep-copies the node.
func (n *edge) Clone() node {
	clone := &edge{
		Path: make([]byte, len(n.Path)),
		Hash: make([]byte, len(n.Hash)),
	}

	copy(clone.Path, n.Path)
	copy(clone.Hash, n.Hash)

	return clone
}

// String returns a string representation of the node.
func (n *edge) String() string {
	return fmt.Sprintf(
		"%v %v %s",
		nodeTypeEdge,
		newNibsFromNibs(n.Path...),
		n.Hash.B58String(),
	)
}

// unmarshalNode unmarshals a node. It returns a node and the number of bytes
// read if no error occured.
func unmarshalNode(buf []byte) (node, int, error) {
	if len(buf) < 1 {
		return nil, 0, errors.WithStack(ErrInvalidNodeLen)
	}

	// Get node type (the first byte).
	typ := nodeType(buf[0])
	if typ == nodeTypeNull {
		return null{}, 1, nil
	}

	buf = buf[1:]
	read := 1

	if typ == nodeTypeEdge {
		// Edge type is followed by a path and a multihash.
		p, pathLen, err := unmarshalPath(buf)
		if err != nil {
			return nil, 0, err
		}

		buf = buf[pathLen:]
		read += pathLen

		hashLen := multihashLen(buf)
		if hashLen <= 0 || len(buf) < hashLen {
			return nil, 0, errors.WithStack(ErrInvalidNodeLen)
		}

		// Copy the buffer.
		hash := make(multihash.Multihash, hashLen)
		copy(hash, buf[:hashLen])

		return &edge{
			Path: nibs(p).Expand(),
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
	case nodeTypeBranch:
		n := &branch{Value: val}

		// Get child hash nodes.
		for i := range n.EmbeddedNodes {
			child, childRead, err := unmarshalNode(buf)
			if err != nil {
				return nil, 0, err
			}

			// They can only be null or edge.
			switch child.(type) {
			case null, *edge:
			default:
				return nil, 0, errors.WithStack(ErrInvalidNodeType)
			}

			n.EmbeddedNodes[i] = child
			buf = buf[childRead:]
			read += childRead
		}

		return n, read, nil

	case nodeTypeLeaf:
		return &leaf{Value: val}, read, nil
	}

	return nil, 0, errors.WithStack(ErrInvalidNodeType)
}

// multihashLen returns the length of a multihash.
func multihashLen(buf []byte) int {
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
func marshalValueNode(typ nodeType, val []byte, headroom int) []byte {
	buf := make([]byte, 1+binary.MaxVarintLen64+len(val)+headroom)
	buf[0] = byte(typ)

	l := binary.PutUvarint(buf[1:], uint64(len(val)))
	copy(buf[1+l:], val)

	return buf[:1+l+len(val)]
}
