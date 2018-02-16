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
	pb "github.com/stratumn/alice/pb/trie"
)

var (
	// ErrProofEmpty is returned when the length of the proof is zero.
	ErrProofEmpty = errors.New("the proof is empty")

	// ErrChildNotFound is returned when the hash of a node was not found
	// in its parent.
	ErrChildNotFound = errors.New("the child node was not found")

	// ErrInvalidMerkleRoot is returned when the Merkle root of the proof
	// is invalid.
	ErrInvalidMerkleRoot = errors.New("the Merkle root is invalid")

	// ErrInvalidKey is returned when the value of the proof is invalid.
	ErrInvalidKey = errors.New("the key is invalid")

	// ErrInvalidValue is returned when the value of the proof is invalid.
	ErrInvalidValue = errors.New("the value is invalid")
)

// ProofNode describes a node in a proof.
type ProofNode struct {
	// Key and Value are only set for value branches and leaves.
	Key   []byte
	Value []byte

	// ChildHashes contains the hash of the child nodes if the node is a
	// branch.
	ChildHashes []multihash.Multihash
}

// NewProofNodeFromProto converts protobuf message to a proof node.
func NewProofNodeFromProto(msg *pb.ProofNode) ProofNode {
	var childHashes []multihash.Multihash

	if len(msg.ChildHashes) > 0 {
		// Otherwise test fails because nil != []multiaddr.Multiaddr{}.
		childHashes = make([]multihash.Multihash, len(msg.ChildHashes))
	}

	for i, hash := range msg.ChildHashes {
		childHashes[i] = make(multihash.Multihash, len(hash))
		copy(childHashes[i], hash)
	}

	return ProofNode{
		Key:         msg.Key,
		Value:       msg.Value,
		ChildHashes: childHashes,
	}
}

// Hash hashes the node.
//
// Hashing is really simple. The data hashed is the concatenation of the key,
// the value, and all the child hashes.
//
// See Multihash for hash codes.
func (n ProofNode) Hash(hashCode uint64) (multihash.Multihash, error) {
	bufLen := len(n.Key) + len(n.Value)

	for _, childHash := range n.ChildHashes {
		bufLen += len(childHash)
	}

	buf := make([]byte, bufLen)
	b := buf

	copy(b, n.Key)
	b = b[len(n.Key):]

	copy(b, n.Value)
	b = b[len(n.Value):]

	for _, childHash := range n.ChildHashes {
		copy(b, childHash)
		b = b[len(childHash):]
	}

	hash, err := multihash.Sum(buf, hashCode, -1)

	return hash, errors.WithStack(err)
}

// ToProto converts the node to a protobuf message.
func (n ProofNode) ToProto() *pb.ProofNode {
	childHashes := make([][]byte, len(n.ChildHashes))

	for i, hash := range n.ChildHashes {
		childHashes[i] = make([]byte, len(hash))
		copy(childHashes[i], hash)
	}

	return &pb.ProofNode{
		Key:         n.Key,
		Value:       n.Value,
		ChildHashes: childHashes,
	}
}

// Proof contains evidence that a value is in a Patricia Merkle Trie.
type Proof []ProofNode

// NewProofFromProto converts a slice of protobuf messages to a proof.
func NewProofFromProto(msgs []*pb.ProofNode) Proof {
	p := make(Proof, len(msgs))

	for i, m := range msgs {
		p[i] = NewProofNodeFromProto(m)
	}

	return p
}

// Key returns the key contained in the proof.
func (p Proof) Key() ([]byte, error) {
	if len(p) < 0 {
		return nil, errors.WithStack(ErrProofEmpty)
	}

	return p[0].Key, nil
}

// Value returns the value contained in the proof.
func (p Proof) Value() ([]byte, error) {
	if len(p) < 0 {
		return nil, errors.WithStack(ErrProofEmpty)
	}

	return p[0].Value, nil
}

// MerkleRoot returns the Merkle root contained in the proof.
func (p Proof) MerkleRoot(hashCode uint64) ([]byte, error) {
	if len(p) < 0 {
		return nil, errors.WithStack(ErrProofEmpty)
	}

	return p[len(p)-1].Hash(hashCode)
}

// Verify verifies the proof the given Merkle root, key, and value.
func (p Proof) Verify(merkleRoot multihash.Multihash, key, val []byte) error {
	if len(p) < 0 {
		return errors.WithStack(ErrProofEmpty)
	}

	decodedRoot, err := multihash.Decode(merkleRoot)
	if err != nil {
		return errors.WithStack(err)
	}

	mr, err := p.MerkleRoot(decodedRoot.Code)
	if err != nil {
		return err
	}

	if bytes.Compare(mr, merkleRoot) != 0 {
		return errors.WithStack(ErrInvalidMerkleRoot)
	}

	k, err := p.Key()
	if err != nil {
		return err
	}

	if bytes.Compare(k, key) != 0 {
		return errors.WithStack(ErrInvalidKey)
	}

	v, err := p.Value()
	if err != nil {
		return err
	}

	if bytes.Compare(v, val) != 0 {
		return errors.WithStack(ErrInvalidValue)
	}

	// Verify parent hashes.
	child := p[0]

nodeLoop:
	for _, parent := range p[1:] {
		for _, childHash := range parent.ChildHashes {
			decodedRoot, err := multihash.Decode(childHash)
			if err != nil {
				return errors.WithStack(err)
			}

			hash, err := child.Hash(decodedRoot.Code)
			if err != nil {
				return err
			}

			if bytes.Compare(childHash, hash) != 0 {
				continue
			}

			child = parent

			continue nodeLoop
		}

		return errors.WithStack(ErrChildNotFound)
	}

	return nil
}

// ToProto converts the proof to a slice of protobuf messages.
func (p Proof) ToProto() []*pb.ProofNode {
	msgs := make([]*pb.ProofNode, len(p))

	for i, n := range p {
		msgs[i] = n.ToProto()
	}

	return msgs
}

// nodeToProof converts a node to a proof node. The key must be in nibble
// format.
func nodeToProof(key []uint8, n node) ProofNode {
	key = newNibsFromNibs(key...).buf
	proof := ProofNode{}

	switch n := n.(type) {
	case *branch:
		if len(proof.Value) > 0 {
			proof.Key = key
			proof.Value = n.Value
		}

		for _, e := range n.EmbeddedNodes {
			e, ok := e.(*edge)
			if !ok {
				continue
			}

			proof.ChildHashes = append(proof.ChildHashes, e.Hash)
		}

	case *leaf:
		proof.Key = key
		proof.Value = n.Value
	}

	return proof
}
