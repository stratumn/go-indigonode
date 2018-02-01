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

package chain

import (
	"crypto/sha256"
	"encoding/json"
	"testing"

	db "github.com/stratumn/alice/core/protocol/coin/db"
	pb "github.com/stratumn/alice/pb/coin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestChain(t *testing.T) {
	block1 := &pb.Block{Header: &pb.Header{BlockNumber: 1}}
	b1, err := json.Marshal(block1)
	assert.NoError(t, err, "json.Marshal(block1)")
	h1 := sha256.Sum256(b1)

	block2 := &pb.Block{Header: &pb.Header{BlockNumber: 2}}
	b2, err := json.Marshal(block2)
	assert.NoError(t, err, "json.Marshal(block2)")
	h2 := sha256.Sum256(b2)

	tests := []struct {
		name string
		run  func(*testing.T, Chain)
	}{{
		"empty-chain",
		func(t *testing.T, c Chain) {
			h, err := c.CurrentHeader()
			assert.EqualError(t, err, ErrBlockHashNotFound.Error(), "c.GetBlock(block1)")
			assert.Nil(t, h, "s.CurrentHeader()")
		},
	}, {
		"add-get",
		func(t *testing.T, c Chain) {
			assert.NoError(t, c.AddBlock(block1), "c.AddBlock()")
			b, err := c.GetBlock(h1[:], block1.Header.BlockNumber)
			assert.NoError(t, err, "c.GetBlock(block1)")
			assert.Equal(t, b, block1)
		},
	}, {
		"last-block",
		func(t *testing.T, c Chain) {
			assert.NoError(t, c.AddBlock(block1), "c.AddBlock()")
			assert.NoError(t, c.AddBlock(block2), "c.AddBlock()")

			h, err := c.CurrentHeader()
			assert.NoError(t, err, "c.CurrentHeader()")
			assert.Equal(t, h, block2.Header)
		},
	}, {
		"get-by-number",
		func(t *testing.T, c Chain) {
			assert.NoError(t, c.AddBlock(block1), "c.AddBlock()")
			assert.NoError(t, c.AddBlock(block2), "c.AddBlock()")

			h, err := c.GetHeaderByNumber(block1.Header.BlockNumber)
			assert.NoError(t, err, "c.GetHeaderByNumber()")
			assert.Equal(t, []*pb.Header{block1.Header}, h)

			h, err = c.GetHeaderByNumber(block2.Header.BlockNumber)
			assert.NoError(t, err, "c.GetHeaderByNumber()")
			assert.Equal(t, []*pb.Header{block2.Header}, h)
		},
	}, {
		"get-by-number-multiple",
		func(t *testing.T, c Chain) {
			block := &pb.Block{Header: &pb.Header{BlockNumber: 1, Nonce: 42}}

			assert.NoError(t, c.AddBlock(block1), "c.AddBlock()")
			assert.NoError(t, c.AddBlock(block), "c.AddBlock()")

			headers, err := c.GetHeaderByNumber(block1.Header.BlockNumber)
			assert.NoError(t, err, "c.GetHeaderByNumber()")
			assert.Len(t, headers, 2, "c.GetHeaderByNumber()")
			assert.Equal(t, []*pb.Header{block1.Header, block.Header}, headers)
		},
	}, {
		"get-by-bad-number",
		func(t *testing.T, c Chain) {
			_, err := c.GetHeaderByNumber(42)
			assert.EqualError(t, err, ErrBlockNumberNotFound.Error(), "c.GetHeaderByNumber()")
		},
	}, {
		"get-by-hash",
		func(t *testing.T, c Chain) {
			assert.NoError(t, c.AddBlock(block1), "c.AddBlock()")
			assert.NoError(t, c.AddBlock(block2), "c.AddBlock()")

			h, err := c.GetHeaderByHash(h1[:])
			assert.NoError(t, err, "c.GetHeaderByHash()")
			assert.Equal(t, h, block1.Header)

			h, err = c.GetHeaderByHash(h2[:])
			assert.NoError(t, err, "c.GetHeaderByHash()")
			assert.Equal(t, h, block2.Header)
		},
	}, {
		"get-by-bad-hash",
		func(t *testing.T, c Chain) {
			_, err := c.GetHeaderByHash(h1[:])
			assert.EqualError(t, err, ErrBlockHashNotFound.Error(), "c.GetHeaderByHash()")
		},
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			memdb, err := db.NewMemDB(nil)
			require.NoError(t, err, "db.NewMemDB()")
			defer memdb.Close()

			tt.run(t, NewDBChain(memdb))
		})
	}
}
