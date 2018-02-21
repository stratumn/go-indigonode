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

package coin_test

import (
	"encoding/hex"
	"testing"

	pb "github.com/stratumn/alice/pb/coin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	peer "gx/ipfs/Qma7H6RW8wRrfZpNSXwxYGcd1E149s42FpWNpDNieSVrnU/go-libp2p-peer"
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
