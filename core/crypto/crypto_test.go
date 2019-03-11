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

package crypto_test

import (
	"context"
	"crypto/rand"
	"testing"

	"github.com/stratumn/go-node/core/crypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	ic "github.com/libp2p/go-libp2p-crypto"
)

func TestSignVerify(t *testing.T) {
	ctx := context.Background()

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
			sig, err := crypto.Sign(ctx, sk, []byte("h3ll0_42"))
			require.NoError(t, err, "crypto.Sign()")
			assert.Equal(t, tt.expectedKeyType, sig.KeyType)

			valid := sig.Verify(ctx, []byte("h3ll0_42"))
			assert.True(t, valid, "sig.Verify()")
		})
	}
}

func TestVerify(t *testing.T) {
	ctx := context.Background()

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

			sig, _ := crypto.Sign(ctx, sk1, []byte("h3ll0"))
			sig.PublicKey, _ = pk2.Bytes()

			return sig
		},
		[]byte("h3ll0"),
		false,
	}, {
		"payload-mismatch",
		func() *crypto.Signature {
			sk1, _, _ := ic.GenerateEd25519Key(rand.Reader)
			sig, _ := crypto.Sign(ctx, sk1, []byte("g00dby3"))
			return sig
		},
		[]byte("h3ll0"),
		false,
	}, {
		"invalid-signature",
		func() *crypto.Signature {
			sk1, _, _ := ic.GenerateEd25519Key(rand.Reader)
			sig, _ := crypto.Sign(ctx, sk1, []byte("h3ll0"))
			swap := sig.Signature[3]
			for i := 4; i < len(sig.Signature); i++ {
				if sig.Signature[3] != sig.Signature[i] {
					sig.Signature[3] = sig.Signature[i]
					sig.Signature[i] = swap
					break
				}
			}
			return sig
		},
		[]byte("h3ll0"),
		false,
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sig := tt.signature()
			valid := sig.Verify(ctx, tt.payload)
			assert.Equal(t, tt.expected, valid)
		})
	}
}
