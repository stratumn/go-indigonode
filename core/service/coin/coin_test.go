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
	"testing"

	"github.com/stretchr/testify/assert"

	crypto "gx/ipfs/QmaPbCnUMBohSGo3KnxEa2bHqyJVVeEEcwtqJAYxerieBo/go-libp2p-crypto"
)

func TestCoinConfig(t *testing.T) {
	_, pubKey, err := crypto.GenerateKeyPair(crypto.Ed25519, 0)
	assert.NoError(t, err, "crypto.GenerateKeyPair()")

	pubKeyBytes, err := pubKey.Bytes()
	assert.NoError(t, err, "pubKey.Bytes()")

	t.Run("decode-public-key", func(t *testing.T) {
		config := &Config{MinerPublicKey: crypto.ConfigEncodeKey(pubKeyBytes)}

		decodedPublicKey, err := config.GetMinerPublicKey()
		assert.NoError(t, err, "config.GetMinerPublicKey()")

		decodedPublicKeyBytes, err := decodedPublicKey.Bytes()
		assert.NoError(t, err, "decodedPublicKey.Bytes()")
		assert.EqualValues(t, pubKeyBytes, decodedPublicKeyBytes)
	})

	t.Run("missing-public-key", func(t *testing.T) {
		config := &Config{}

		_, err := config.GetMinerPublicKey()
		assert.EqualError(t, err, ErrMissingMinerPublicKey.Error())
	})
}
