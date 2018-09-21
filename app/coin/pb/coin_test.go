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

package pb_test

import (
	"encoding/hex"
	"testing"

	pb "github.com/stratumn/go-indigonode/app/coin/pb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	peer "gx/ipfs/QmQsErDt8Qgw1XrsXf2BpEzDgGWtB1YLsTAARBup5b6B9W/go-libp2p-peer"
)

func TestBlockModel(t *testing.T) {
	b := &pb.Block{}
	assert.Equal(t, uint64(0), b.BlockNumber(), "b.BlockNumber()")
	assert.Equal(t, []byte{}, b.PreviousHash(), "b.PreviousHash()")

	b = &pb.Block{Header: &pb.Header{BlockNumber: 42, PreviousHash: []byte("pizza")}}
	assert.Equal(t, b.Header.BlockNumber, b.BlockNumber(), "b.BlockNumber()")
	assert.Equal(t, b.Header.PreviousHash, b.PreviousHash(), "b.PreviousHash()")
}

func TestLoggable(t *testing.T) {
	fromID := "QmQnYf23kQ7SvuPZ3mQcg3RuJMr9E39fBvm89Nz4bevJdt"
	from, err := peer.IDB58Decode(fromID)
	require.NoError(t, err)

	toID := "QmYxjwtGm1mKL61Cc6NhBRgMWFz39r3k5tRRqWWiQoux7V"
	to, err := peer.IDB58Decode(toID)
	require.NoError(t, err)

	tx := &pb.Transaction{
		From:  []byte(from),
		To:    []byte(to),
		Value: 42,
		Fee:   2,
		Nonce: 1,
	}

	block := &pb.Block{
		Header: &pb.Header{
			BlockNumber:  42,
			Nonce:        3,
			PreviousHash: []byte("spongebob"),
			MerkleRoot:   []byte("patrick"),
		},
		Transactions: []*pb.Transaction{tx},
	}

	t.Run("Transaction", func(t *testing.T) {
		loggable := tx.Loggable()
		assert.Equal(t, fromID, loggable["from"], "from")
		assert.Equal(t, toID, loggable["to"], "to")
		assert.Equal(t, uint64(42), loggable["value"], "value")
		assert.Equal(t, uint64(2), loggable["fee"], "fee")
		assert.Equal(t, uint64(1), loggable["nonce"], "nonce")
	})

	t.Run("Block", func(t *testing.T) {
		loggable := block.Loggable()
		assert.Len(t, loggable["txs"], 1, "txs")
		header := loggable["header"].(map[string]interface{})
		assert.Equal(t, uint64(42), header["block_number"], "block_number")
		assert.Equal(t, uint64(3), header["nonce"], "nonce")
		assert.Equal(t, hex.EncodeToString(block.Header.MerkleRoot), header["merkle_root"], "merkle_root")
		assert.Equal(t, hex.EncodeToString(block.Header.PreviousHash), header["previous_hash"], "previous_hash")
	})
}
