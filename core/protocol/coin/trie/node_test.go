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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	multihash "gx/ipfs/QmZyZDi491cCNTLfAhwcaDii2Kg4pwKRkhqQzURGDvY6ua/go-multihash"
)

var (
	testHash1 multihash.Multihash
	testHash2 multihash.Multihash
)

func init() {
	var err error

	testHash1, err = multihash.Sum([]byte("bob"), multihash.SHA2_256, -1)
	if err != nil {
		panic(err)
	}

	testHash2, err = multihash.Sum([]byte("charlie"), multihash.SHA2_256, -1)
	if err != nil {
		panic(err)
	}
}

func TestNode_Encoding(t *testing.T) {
	tests := []struct {
		name string
		node Node
		err  error
	}{{
		"null-node",
		NullNode{},
		nil,
	}, {
		"leaf-node",
		LeafNode{Value: []byte("Alice")},
		nil,
	}, {
		"leaf-node-odd",
		LeafNode{Value: []byte("Alice"), IsOdd: true},
		nil,
	}, {
		"hash-node",
		HashNode{Hash: testHash1},
		nil,
	}, {
		"parent-node",
		ParentNode{
			Value: []byte("Alice"),
			Children: []Node{
				NullNode{},
				NullNode{},
				NullNode{},
				HashNode{Hash: testHash1},
				NullNode{},
				NullNode{},
				NullNode{},
				NullNode{},
				NullNode{},
				NullNode{},
				NullNode{},
				HashNode{Hash: testHash2},
				NullNode{},
				NullNode{},
				NullNode{},
				NullNode{},
			},
		},
		nil,
	}, {
		"parent-node-invalid-length",
		ParentNode{
			Value: []byte("Alice"),
			Children: []Node{
				NullNode{},
				HashNode{Hash: testHash1},
			},
		},
		ErrInvalidNodeLen,
	}, {
		"parent-node-invalid-type",
		ParentNode{
			Value: []byte("Alice"),
			Children: []Node{
				NullNode{},
				NullNode{},
				NullNode{},
				HashNode{Hash: testHash1},
				NullNode{},
				NullNode{},
				NullNode{},
				NullNode{},
				NullNode{},
				NullNode{},
				NullNode{},
				LeafNode{Value: testHash2},
				NullNode{},
				NullNode{},
				NullNode{},
				NullNode{},
			},
		},
		ErrInvalidNodeType,
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := tt.node.Marshal()
			node, read, err := UnmarshalNode(buf)

			if tt.err != nil {
				assert.EqualError(t, err, tt.err.Error(), "UnmarshalNode()")
				return
			}

			require.NoError(t, err, "UnmarshalNode()")
			assert.Equal(t, len(buf), read, "UnmarshalNode().read")
			assert.Equal(t, tt.node, node, "UnmarshalNode().node")
		})
	}
}

func TestNode_String(t *testing.T) {
	n := ParentNode{
		Value: []byte("Alice"),
		Children: []Node{
			NullNode{},
			NullNode{},
			NullNode{},
			HashNode{Hash: testHash1},
			NullNode{},
			NullNode{},
			NullNode{},
			NullNode{},
			NullNode{},
			NullNode{},
			NullNode{},
			HashNode{Hash: testHash2},
			NullNode{},
			NullNode{},
			NullNode{},
			NullNode{},
		},
	}

	expect := "<parent> false 416c696365 [<null> <null> <null> <hash> " +
		"QmX4zTUJa1vDXjw3mTxwXBdCd9gThbggaHFGhA1QpnKdK6x " +
		"<null> <null> <null> <null> <null> <null> <null> <hash> " +
		"QmarC75CYs3HLzgSXfUdZqJatFZb6Pmj4QVJmbBwX3R1K9x " +
		"<null> <null> <null> <null>]"

	assert.EqualValues(t, expect, n.String())
}
