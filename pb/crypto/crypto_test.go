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

package crypto_test

import (
	"crypto/rand"
	"testing"

	"github.com/stratumn/alice/pb/crypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	ic "gx/ipfs/Qme1knMqwt1hKZbc1BmQFmnm9f36nyQGwXxPGVpVJ9rMK5/go-libp2p-crypto"
)

func TestSignVerify(t *testing.T) {
	tests := []struct {
		name            string
		generateKey     func() ic.PrivKey
		expectedKeyType crypto.KeyType
	}{{
		"ed25519",
		func() ic.PrivKey {
			sk, _, _ := ic.GenerateEd25519Key(rand.Reader)
			return sk
		},
		crypto.KeyType_Ed25519,
	}, {
		"secp256k1",
		func() ic.PrivKey {
			sk, _, _ := ic.GenerateSecp256k1Key(rand.Reader)
			return sk
		},
		crypto.KeyType_Secp256k1,
	}, {
		"RSA",
		func() ic.PrivKey {
			sk, _, _ := ic.GenerateKeyPair(0, 1024)
			return sk
		},
		crypto.KeyType_RSA,
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sk := tt.generateKey()
			sig, err := crypto.Sign(sk, []byte("h3ll0_42"))
			require.NoError(t, err, "crypto.Sign()")
			assert.Equal(t, tt.expectedKeyType, sig.KeyType)

			valid := sig.Verify([]byte("h3ll0_42"))
			assert.True(t, valid, "sig.Verify()")
		})
	}
}

func TestVerify(t *testing.T) {
	tests := []struct {
		name      string
		signature func() *crypto.Signature
		payload   []byte
		expected  bool
	}{{
		"nil-signature",
		func() *crypto.Signature {
			return nil
		},
		[]byte("h3ll0"),
		false,
	}, {
		"public-key-mismatch",
		func() *crypto.Signature {
			sk1, _, _ := ic.GenerateEd25519Key(rand.Reader)
			_, pk2, _ := ic.GenerateEd25519Key(rand.Reader)

			sig, _ := crypto.Sign(sk1, []byte("h3ll0"))
			sig.PublicKey, _ = pk2.Bytes()

			return sig
		},
		[]byte("h3ll0"),
		false,
	}, {
		"payload-mismatch",
		func() *crypto.Signature {
			sk1, _, _ := ic.GenerateEd25519Key(rand.Reader)
			sig, _ := crypto.Sign(sk1, []byte("g00dby3"))
			return sig
		},
		[]byte("h3ll0"),
		false,
	}, {
		"invalid-signature",
		func() *crypto.Signature {
			sk1, _, _ := ic.GenerateEd25519Key(rand.Reader)
			sig, _ := crypto.Sign(sk1, []byte("h3ll0"))
			sig.Signature[3] = sig.Signature[7]
			return sig
		},
		[]byte("h3ll0"),
		false,
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sig := tt.signature()
			valid := sig.Verify(tt.payload)
			assert.Equal(t, tt.expected, valid)
		})
	}
}
