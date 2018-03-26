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

	db "github.com/stratumn/alice/core/db"
	"github.com/stratumn/alice/core/protocol/coin/coinutil"
	pb "github.com/stratumn/alice/pb/coin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestChain(t *testing.T) {
	genesisBlock := &pb.Block{Header: &pb.Header{BlockNumber: 0}}
	genesisHash, err := coinutil.HashHeader(genesisBlock.Header)
	assert.NoError(t, err, "coinutil.HashHeader()")

	block1 := &pb.Block{Header: &pb.Header{BlockNumber: 1, PreviousHash: genesisHash}}
	h1, err := coinutil.HashHeader(block1.Header)
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
			assert.EqualError(t, err, ErrBlockNotFound.Error(), "c.GetBlock(block1)")
			assert.Nil(t, h, "s.CurrentHeader()")
		},
	}, {
		"add-invalid-previous-hash",
		func(t *testing.T, c Chain) {
			block := &pb.Block{Header: &pb.Header{BlockNumber: 0, PreviousHash: h1}}
			assert.EqualError(t, c.AddBlock(block), ErrBlockNotFound.Error(), "c.AddBlock()")
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
			assert.NoError(t, c.AddBlock(block1), "c.AddBlock()")
			assert.NoError(t, c.SetHead(block1), "c.SetHead()")

			h, err := c.CurrentHeader()
			assert.NoError(t, err, "c.CurrentHeader()")
			assert.Equal(t, h, block1.Header)

			b, err := c.CurrentBlock()
			assert.NoError(t, err, "c.CurrentBlock()")
			assert.Equal(t, b, block1)
		},
	}, {
		"set-bad-head",
		func(t *testing.T, c Chain) {
			assert.EqualError(t, c.SetHead(block1), ErrBlockNotFound.Error(), "c.SetHead()")
		},
	}, {
		"get-by-number",
		func(t *testing.T, c Chain) {
			assert.NoError(t, c.AddBlock(block1), "c.AddBlock()")
			// Get header.
			h, err := c.GetHeadersByNumber(genesisBlock.Header.BlockNumber)
			assert.NoError(t, err, "c.GetHeadersByNumber()")
			assert.Equal(t, []*pb.Header{genesisBlock.Header}, h)

			h, err = c.GetHeadersByNumber(block1.Header.BlockNumber)
			assert.NoError(t, err, "c.GetHeadersByNumber()")
			assert.Equal(t, []*pb.Header{block1.Header}, h)
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
			assert.EqualError(t, err, ErrBlockNotFound.Error(), "c.GetHeadersByNumber()")
		},
	}, {
		"get-by-hash",
		func(t *testing.T, c Chain) {
			assert.NoError(t, c.AddBlock(block1), "c.AddBlock()")

			h, err := c.GetHeaderByHash(genesisHash)
			assert.NoError(t, err, "c.GetHeaderByHash()")
			assert.Equal(t, h, genesisBlock.Header)

			h, err = c.GetHeaderByHash(h1)
			assert.NoError(t, err, "c.GetHeaderByHash()")
			assert.Equal(t, h, block1.Header)
		},
	}, {
		"get-by-bad-hash",
		func(t *testing.T, c Chain) {
			_, err := c.GetHeaderByHash(h1)
			assert.EqualError(t, err, ErrBlockNotFound.Error(), "c.GetHeaderByHash()")
		},
	}, {
		"update-main-branch",
		func(t *testing.T, c Chain) {
			// Update the head and check that the main branch references are updated.
			block1bis := &pb.Block{Header: &pb.Header{BlockNumber: 1, PreviousHash: genesisHash, Nonce: 42}}
			h1bis, err := coinutil.HashHeader(block1bis.Header)
			assert.NoError(t, err, "coinutil.HashHeader()")
			block2 := &pb.Block{Header: &pb.Header{BlockNumber: 2, PreviousHash: h1, Nonce: 42}}
			block2bis := &pb.Block{Header: &pb.Header{BlockNumber: 2, PreviousHash: h1bis, Nonce: 43}}

			assert.NoError(t, c.AddBlock(block1), "c.AddBlock()")
			assert.NoError(t, c.SetHead(block1), "c.SetHead()")
			assert.NoError(t, c.AddBlock(block2), "c.AddBlock()")
			assert.NoError(t, c.SetHead(block2), "c.SetHead()")

			assert.NoError(t, c.AddBlock(block1bis), "c.AddBlock()")
			assert.NoError(t, c.AddBlock(block2bis), "c.AddBlock()")

			// Get block.
			b, err := c.GetBlockByNumber(1)
			assert.NoError(t, err, "c.GetHeaderByNumber()")
			assert.Equal(t, b, block1)
			b, err = c.GetBlockByNumber(2)
			assert.NoError(t, err, "c.GetHeaderByNumber()")
			assert.Equal(t, b, block2)

			// Get header.
			h, err := c.GetHeaderByNumber(1)
			assert.NoError(t, err, "c.GetHeaderByNumber()")
			assert.Equal(t, h, block1.Header)
			h, err = c.GetHeaderByNumber(2)
			assert.NoError(t, err, "c.GetHeaderByNumber()")
			assert.Equal(t, h, block2.Header)

			assert.NoError(t, c.SetHead(block2bis), "c.SetHead()")

			// Get block.
			b, err = c.GetBlockByNumber(1)
			assert.NoError(t, err, "c.GetHeaderByNumber()")
			assert.Equal(t, b, block1bis)
			b, err = c.GetBlockByNumber(2)
			assert.NoError(t, err, "c.GetHeaderByNumber()")
			assert.Equal(t, b, block2bis)

			// Get header.
			h, err = c.GetHeaderByNumber(1)
			assert.NoError(t, err, "c.GetHeaderByNumber()")
			assert.Equal(t, h, block1bis.Header)
			h, err = c.GetHeaderByNumber(2)
			assert.NoError(t, err, "c.GetHeaderByNumber()")
			assert.Equal(t, h, block2bis.Header)
		},
	}, {
		"get-by-number-main-chain",
		func(t *testing.T, c Chain) {
			// Check that the refs to the main chain are updated.
			block2 := &pb.Block{Header: &pb.Header{BlockNumber: 2, PreviousHash: h1, Nonce: 42}}

			assert.NoError(t, c.AddBlock(block1), "c.AddBlock()")
			assert.NoError(t, c.SetHead(block1), "c.SetHead()")

			b, err := c.GetBlockByNumber(1)
			assert.NoError(t, err, "c.GetHeaderByNumber()")
			assert.Equal(t, b, block1)
			h, err := c.GetHeaderByNumber(1)
			assert.NoError(t, err, "c.GetHeaderByNumber()")
			assert.Equal(t, h, block1.Header)

			assert.NoError(t, c.AddBlock(block2), "c.AddBlock()")
			assert.NoError(t, c.SetHead(block2), "c.SetHead()")

			b, err = c.GetBlockByNumber(2)
			assert.NoError(t, err, "c.GetHeaderByNumber()")
			assert.Equal(t, b, block2)
			h, err = c.GetHeaderByNumber(2)
			assert.NoError(t, err, "c.GetHeaderByNumber()")
			assert.Equal(t, h, block2.Header)
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

func TestPath(t *testing.T) {
	// This test will retrieve the path between the head b4 and new blocks:
	// b4Child a child of b4 rollbacks[] replays[]
	// b5Child a child of b5 rollbacks[b4, b3, b2, b1] replays [b5]
	// b7Child a child of b7 rollbacks[b4, b3, b2] replays [b6, b7]
	// b9Child a child of b9 rollbacks[b4] replays[b8, b9]
	//
	// The chain looks like this
	// genesis - b1 - b2 - b3 - b4
	//         \    \         \
	//           b5   b6        b8
	//                   \         \
	//                     b7        b9

	genesisBlock := &pb.Block{Header: &pb.Header{BlockNumber: 0}}
	genesisHash, err := coinutil.HashHeader(genesisBlock.Header)
	require.NoError(t, err)

	// Main branch
	b1 := &pb.Block{Header: &pb.Header{BlockNumber: 1, PreviousHash: genesisHash, Nonce: 1}}
	b1Hash, err := coinutil.HashHeader(b1.Header)
	require.NoError(t, err)

	b2 := &pb.Block{Header: &pb.Header{BlockNumber: 2, PreviousHash: b1Hash, Nonce: 2}}
	b2Hash, err := coinutil.HashHeader(b2.Header)
	require.NoError(t, err)

	b3 := &pb.Block{Header: &pb.Header{BlockNumber: 3, PreviousHash: b2Hash, Nonce: 3}}
	b3Hash, err := coinutil.HashHeader(b3.Header)
	require.NoError(t, err)

	b4 := &pb.Block{Header: &pb.Header{BlockNumber: 4, PreviousHash: b3Hash, Nonce: 4}}
	b4Hash, err := coinutil.HashHeader(b4.Header)
	require.NoError(t, err)

	b4Child := &pb.Block{Header: &pb.Header{BlockNumber: 5, PreviousHash: b4Hash}}

	// First fork
	b5 := &pb.Block{Header: &pb.Header{BlockNumber: 1, PreviousHash: genesisHash, Nonce: 5}}
	b5Hash, err := coinutil.HashHeader(b5.Header)
	require.NoError(t, err)

	b5Child := &pb.Block{Header: &pb.Header{BlockNumber: 2, PreviousHash: b5Hash}}

	// Second fork
	b6 := &pb.Block{Header: &pb.Header{BlockNumber: 2, PreviousHash: b1Hash, Nonce: 6}}
	b6Hash, err := coinutil.HashHeader(b6.Header)
	require.NoError(t, err)

	b7 := &pb.Block{Header: &pb.Header{BlockNumber: 3, PreviousHash: b6Hash, Nonce: 7}}
	b7Hash, err := coinutil.HashHeader(b7.Header)
	require.NoError(t, err)

	b7Child := &pb.Block{Header: &pb.Header{BlockNumber: 4, PreviousHash: b7Hash}}

	// Third fork
	b8 := &pb.Block{Header: &pb.Header{BlockNumber: 4, PreviousHash: b3Hash, Nonce: 8}}
	b8Hash, err := coinutil.HashHeader(b8.Header)
	require.NoError(t, err)

	b9 := &pb.Block{Header: &pb.Header{BlockNumber: 5, PreviousHash: b8Hash, Nonce: 9}}
	b9Hash, err := coinutil.HashHeader(b9.Header)
	require.NoError(t, err)

	b9Child := &pb.Block{Header: &pb.Header{BlockNumber: 6, PreviousHash: b9Hash}}

	// Chain setup
	memdb, err := db.NewMemDB(nil)
	require.NoError(t, err)

	c := NewChainDB(memdb, OptGenesisBlock(genesisBlock))

	for _, b := range []*pb.Block{b1, b2, b3, b4, b5, b6, b7, b8, b9} {
		err := c.AddBlock(b)
		require.NoError(t, err)
	}

	err = c.SetHead(b4)
	require.NoError(t, err)

	rollbacks, replays, err := GetPath(c, b4, b4Child)
	require.NoError(t, err)
	assert.Len(t, rollbacks, 0)
	assert.Len(t, replays, 0)

	rollbacks, replays, err = GetPath(c, b4, b5Child)
	require.NoError(t, err)
	assert.Equal(t, []*pb.Block{b4, b3, b2, b1}, rollbacks)
	assert.Equal(t, []*pb.Block{b5}, replays)

	rollbacks, replays, err = GetPath(c, b4, b7Child)
	require.NoError(t, err)
	assert.Equal(t, []*pb.Block{b4, b3, b2}, rollbacks)
	assert.Equal(t, []*pb.Block{b6, b7}, replays)

	rollbacks, replays, err = GetPath(c, b4, b9Child)
	require.NoError(t, err)
	assert.Equal(t, []*pb.Block{b4}, rollbacks)
	assert.Equal(t, []*pb.Block{b8, b9}, replays)

	rollbacks, replays, err = GetPath(c, b4, b2)
	require.NoError(t, err)
	assert.Equal(t, []*pb.Block{b4, b3, b2}, rollbacks)
	assert.Len(t, replays, 0)

	rollbacks, replays, err = GetPath(c, b4, b7)
	require.NoError(t, err)
	assert.Equal(t, []*pb.Block{b4, b3, b2}, rollbacks)
	assert.Equal(t, []*pb.Block{b6}, replays)
}
