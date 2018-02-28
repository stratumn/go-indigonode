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

package coin

import (
	"encoding/hex"
	"testing"

	pb "github.com/stratumn/alice/pb/coin"
	"github.com/stretchr/testify/assert"

	peer "gx/ipfs/Qma7H6RW8wRrfZpNSXwxYGcd1E149s42FpWNpDNieSVrnU/go-libp2p-peer"
	crypto "gx/ipfs/QmaPbCnUMBohSGo3KnxEa2bHqyJVVeEEcwtqJAYxerieBo/go-libp2p-crypto"
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

	t.Run("decode-genesis-block", func(t *testing.T) {
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

		assert.Equal(t, block, decodedBlock, "decodedMinerID")
	})
}
