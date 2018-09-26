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

package service

import (
	"encoding/hex"
	"testing"

	"github.com/stratumn/go-node/app/coin/pb"
	"github.com/stretchr/testify/assert"

	crypto "gx/ipfs/QmPvyPwuCgJ7pDmrKDxRtsScJgBaM5h4EpRL2qQJsmXf4n/go-libp2p-crypto"
	peer "gx/ipfs/QmQsErDt8Qgw1XrsXf2BpEzDgGWtB1YLsTAARBup5b6B9W/go-libp2p-peer"
)

func TestCoinConfig(t *testing.T) {
	_, pubKey, err := crypto.GenerateKeyPair(crypto.Ed25519, 0)
	assert.NoError(t, err, "crypto.GenerateKeyPair()")

	id, err := peer.IDFromPublicKey(pubKey)
	assert.NoError(t, err, "peer.IDFromPublicKey()")

	configEncoded := peer.IDB58Encode(id)

	t.Run("decode-miner-d", func(t *testing.T) {
		config := &Config{MinerID: configEncoded}

		decodedMinerID, err := config.GetMinerID()
		assert.NoError(t, err, "config.GetMinerID()")

		assert.Equal(t, id, decodedMinerID, "decodedMinerID")
	})

	t.Run("missing-miner-d", func(t *testing.T) {
		config := &Config{}

		_, err := config.GetMinerID()
		assert.EqualError(t, err, ErrMissingMinerID.Error())
	})

	t.Run("custom-genesis-block", func(t *testing.T) {
		block := &pb.Block{
			Header: &pb.Header{Nonce: 42, Version: 24},
			Transactions: []*pb.Transaction{
				&pb.Transaction{From: []byte("zou"), To: []byte("plap"), Value: 42},
				&pb.Transaction{From: []byte("pizza"), To: []byte("yolo"), Value: 43},
			},
		}
		b, err := block.Marshal()
		assert.NoError(t, err, "block.Marshal()")

		config := &Config{GenesisBlock: hex.EncodeToString(b)}

		decodedBlock, err := config.GetGenesisBlock()
		assert.NoError(t, err, "config.GetGenesisBlock()")

		assert.Equal(t, block, decodedBlock, "decodedGenesisBlock")
	})

	t.Run("default-genesis-block", func(t *testing.T) {
		config := &Config{GenesisBlock: ""}

		block, err := config.GetGenesisBlock()
		assert.NoError(t, err, "config.GetGenesisBlock()")

		gen, err := GetGenesisBlock()
		assert.NoError(t, err, "GetGenesisBlock()")

		assert.Equal(t, gen, block, "decodedGenesisBlock")
	})
}
