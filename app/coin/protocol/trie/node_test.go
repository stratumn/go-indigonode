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

package trie

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gx/ipfs/QmPnFwZ2JXKnXgMw8CdBPxn7FWh6LLdjUjxV1fKHuJnkr8/go-multihash"
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
