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
	"testing"

	"github.com/stratumn/alice/core/protocol/coin/coinutil"
	db "github.com/stratumn/alice/core/protocol/coin/db"
	pb "github.com/stratumn/alice/pb/coin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestChain(t *testing.T) {
	genesisBlock := &pb.Block{Header: &pb.Header{BlockNumber: 0}}
	genesisHash, err := coinutil.HashHeader(genesisBlock.Header)
	assert.NoError(t, err, "coinutil.HashHeader()")

	block2 := &pb.Block{Header: &pb.Header{BlockNumber: 1, PreviousHash: genesisHash}}
	h2, err := coinutil.HashHeader(block2.Header)
	assert.NoError(t, err, "coinutil.HashHeader()")

	tests := []struct {
		name string
		run  func(*testing.T, Chain)
	}{{
		"empty-chain",
		func(t *testing.T, _ Chain) {
			memdb, err := db.NewMemDB(nil)
			require.NoError(t, err, "db.NewMemDB()")
			c := NewChainDB(memdb)
			h, err := c.CurrentHeader()
			assert.EqualError(t, err, ErrBlockHashNotFound.Error(), "c.GetBlock(block1)")
			assert.Nil(t, h, "s.CurrentHeader()")
		},
	}, {
		"add-invalid-previous-hash",
		func(t *testing.T, c Chain) {
			block := &pb.Block{Header: &pb.Header{BlockNumber: 0, PreviousHash: h2}}
			assert.EqualError(t, c.AddBlock(block), ErrBlockHashNotFound.Error(), "c.AddBlock()")
		},
	}, {
		"add-invalid-number",
		func(t *testing.T, c Chain) {
			block := &pb.Block{Header: &pb.Header{BlockNumber: 42, PreviousHash: genesisHash}}
			assert.EqualError(t, c.AddBlock(block), ErrInvalidPreviousBlock.Error(), "c.AddBlock()")
		},
	}, {
		"add-get",
		func(t *testing.T, c Chain) {
			b, err := c.GetBlock(genesisHash, genesisBlock.Header.BlockNumber)
			assert.NoError(t, err, "c.GetBlock(genesisBlock)")
			assert.Equal(t, b, genesisBlock)
		},
	}, {
		"set-head",
		func(t *testing.T, c Chain) {
			assert.NoError(t, c.AddBlock(block2), "c.AddBlock()")
			assert.NoError(t, c.SetHead(block2), "c.SetHead()")

			h, err := c.CurrentHeader()
			assert.NoError(t, err, "c.CurrentHeader()")
			assert.Equal(t, h, block2.Header)

			b, err := c.CurrentBlock()
			assert.NoError(t, err, "c.CurrentBlock()")
			assert.Equal(t, b, block2)
		},
	}, {
		"set-bad-head",
		func(t *testing.T, c Chain) {
			assert.EqualError(t, c.SetHead(block2), ErrBlockHashNotFound.Error(), "c.SetHead()")
		},
	}, {
		"get-by-number",
		func(t *testing.T, c Chain) {
			assert.NoError(t, c.AddBlock(block2), "c.AddBlock()")

			h, err := c.GetHeadersByNumber(genesisBlock.Header.BlockNumber)
			assert.NoError(t, err, "c.GetHeadersByNumber()")
			assert.Equal(t, []*pb.Header{genesisBlock.Header}, h)

			h, err = c.GetHeadersByNumber(block2.Header.BlockNumber)
			assert.NoError(t, err, "c.GetHeadersByNumber()")
			assert.Equal(t, []*pb.Header{block2.Header}, h)
		},
	}, {
		"get-by-number-multiple",
		func(t *testing.T, c Chain) {
			block := &pb.Block{Header: &pb.Header{BlockNumber: 0, Nonce: 42}}

			assert.NoError(t, c.AddBlock(block), "c.AddBlock()")

			headers, err := c.GetHeadersByNumber(genesisBlock.Header.BlockNumber)
			assert.NoError(t, err, "c.GetHeadersByNumber()")
			assert.Len(t, headers, 2, "c.GetHeadersByNumber()")
			assert.Equal(t, []*pb.Header{genesisBlock.Header, block.Header}, headers)
		},
	}, {
		"get-by-bad-number",
		func(t *testing.T, c Chain) {
			_, err := c.GetHeadersByNumber(42)
			assert.EqualError(t, err, ErrBlockNumberNotFound.Error(), "c.GetHeadersByNumber()")
		},
	}, {
		"get-by-hash",
		func(t *testing.T, c Chain) {
			assert.NoError(t, c.AddBlock(block2), "c.AddBlock()")

			h, err := c.GetHeaderByHash(genesisHash)
			assert.NoError(t, err, "c.GetHeaderByHash()")
			assert.Equal(t, h, genesisBlock.Header)

			h, err = c.GetHeaderByHash(h2)
			assert.NoError(t, err, "c.GetHeaderByHash()")
			assert.Equal(t, h, block2.Header)
		},
	}, {
		"get-by-bad-hash",
		func(t *testing.T, c Chain) {
			_, err := c.GetHeaderByHash(h2)
			assert.EqualError(t, err, ErrBlockHashNotFound.Error(), "c.GetHeaderByHash()")
		},
	}, {
		"update-main-branch",
		func(t *testing.T, c Chain) {
			// Update the head and check that the main branch references are updated.
			block2bis := &pb.Block{Header: &pb.Header{BlockNumber: 1, PreviousHash: genesisHash, Nonce: 42}}
			h2bis, err := coinutil.HashHeader(block2bis.Header)
			assert.NoError(t, err, "coinutil.HashHeader()")
			block3 := &pb.Block{Header: &pb.Header{BlockNumber: 2, PreviousHash: h2, Nonce: 42}}
			block3bis := &pb.Block{Header: &pb.Header{BlockNumber: 2, PreviousHash: h2bis, Nonce: 43}}

			assert.NoError(t, c.AddBlock(block2), "c.AddBlock()")
			assert.NoError(t, c.SetHead(block2), "c.SetHead()")
			assert.NoError(t, c.AddBlock(block3), "c.AddBlock()")
			assert.NoError(t, c.SetHead(block3), "c.SetHead()")

			assert.NoError(t, c.AddBlock(block2bis), "c.AddBlock()")
			assert.NoError(t, c.AddBlock(block3bis), "c.AddBlock()")

			h, err := c.GetHeaderByNumber(1)
			assert.NoError(t, err, "c.GetHeaderByNumber()")
			assert.Equal(t, h, block2.Header)
			h, err = c.GetHeaderByNumber(2)
			assert.NoError(t, err, "c.GetHeaderByNumber()")
			assert.Equal(t, h, block3.Header)

			assert.NoError(t, c.SetHead(block3bis), "c.SetHead()")

			h, err = c.GetHeaderByNumber(1)
			assert.NoError(t, err, "c.GetHeaderByNumber()")
			assert.Equal(t, h, block2bis.Header)
			h, err = c.GetHeaderByNumber(2)
			assert.NoError(t, err, "c.GetHeaderByNumber()")
			assert.Equal(t, h, block3bis.Header)
		},
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			memdb, err := db.NewMemDB(nil)
			require.NoError(t, err, "db.NewMemDB()")
			defer memdb.Close()

			tt.run(t, NewChainDB(memdb, OptGenesisBlock(genesisBlock)))
		})
	}
}
