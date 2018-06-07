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
	"gx/ipfs/QmZyZDi491cCNTLfAhwcaDii2Kg4pwKRkhqQzURGDvY6ua/go-multihash"
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
		name         string
		node         node
		marshalErr   error
		unmarshalErr error
	}{{
		"null",
		null{},
		nil,
		nil,
	}, {
		"leaf",
		&leaf{Value: []byte("Alice")},
		nil,
		nil,
	}, {
		"odd-leaf",
		&leaf{Value: []byte("Alice")},
		nil,
		nil,
	}, {
		"edge",
		&edge{Path: []uint8{1, 2, 3}, Hash: testHash1},
		nil,
		nil,
	}, {
		"branch",
		&branch{
			Value: []byte("Alice"),
			EmbeddedNodes: [...]node{
				null{},
				null{},
				null{},
				&edge{Path: []uint8{1, 2, 3}, Hash: testHash1},
				null{},
				null{},
				null{},
				null{},
				null{},
				null{},
				null{},
				&edge{Path: []uint8{10, 11, 12}, Hash: testHash2},
				null{},
				null{},
				null{},
				null{},
			},
		},
		nil,
		nil,
	}, {
		"branch-invalid-type",
		&branch{
			Value: []byte("Alice"),
			EmbeddedNodes: [...]node{
				null{},
				null{},
				null{},
				&edge{Path: []uint8{1, 2, 3}, Hash: testHash1},
				null{},
				null{},
				null{},
				null{},
				null{},
				null{},
				null{},
				&leaf{Value: []byte("Alice")},
				null{},
				null{},
				null{},
				null{},
			},
		},
		nil,
		ErrInvalidNodeType,
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf, err := tt.node.MarshalBinary()

			if tt.marshalErr != nil {
				assert.EqualError(t, err, tt.marshalErr.Error(), "tt.node.MarshalBinary()")
				return
			}

			require.NoError(t, err, "tt.node.MarshalBinary()")

			node, read, err := unmarshalNode(buf)

			if tt.unmarshalErr != nil {
				assert.EqualError(t, err, tt.unmarshalErr.Error(), "UnmarshalNode()")
				return
			}

			require.NoError(t, err, "UnmarshalNode()")
			assert.Equal(t, len(buf), read, "UnmarshalNode().read")
			assert.Equal(t, tt.node, node, "UnmarshalNode().node")
			assert.Equal(t, tt.node, tt.node.Clone(), "tt.node.Clone()")
		})
	}
}

func TestNode_String(t *testing.T) {
	n := branch{
		Value: []byte("Alice"),
		EmbeddedNodes: [...]node{
			null{},
			null{},
			null{},
			&edge{Path: []uint8{1, 2, 3}, Hash: testHash1},
			null{},
			null{},
			null{},
			null{},
			null{},
			null{},
			null{},
			&edge{Path: []uint8{10, 11, 12}, Hash: testHash2},
			null{},
			null{},
			null{},
			null{},
		},
	}

	expect := "<branch> 416c696365 [<null> <null> <null> <edge> 123 " +
		"QmX4zTUJa1vDXjw3mTxwXBdCd9gThbggaHFGhA1QpnKdK6 " +
		"<null> <null> <null> <null> <null> <null> <null> <edge> abc " +
		"QmarC75CYs3HLzgSXfUdZqJatFZb6Pmj4QVJmbBwX3R1K9 " +
		"<null> <null> <null> <null>]"

	assert.EqualValues(t, expect, n.String())
}