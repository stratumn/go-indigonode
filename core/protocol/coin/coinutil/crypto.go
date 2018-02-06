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

package coinutil

import (
	pb "github.com/stratumn/alice/pb/coin"

	ic "gx/ipfs/QmaPbCnUMBohSGo3KnxEa2bHqyJVVeEEcwtqJAYxerieBo/go-libp2p-crypto"
)

// PrivateKey is a private key.
type PrivateKey struct {
	privKey ic.PrivKey
	keyType pb.KeyType
}

// NewPrivateKey creates a new PrivateKey from a libp2p one.
func NewPrivateKey(privKey ic.PrivKey, keyType pb.KeyType) *PrivateKey {
	return &PrivateKey{
		privKey: privKey,
		keyType: keyType,
	}
}

// Bytes returns the key bytes.
func (k PrivateKey) Bytes() ([]byte, error) {
	return k.privKey.Bytes()
}

// GetPublicKey returns the corresponding PublicKey.
func (k PrivateKey) GetPublicKey() *PublicKey {
	pubKey := k.privKey.GetPublic()
	return NewPublicKey(pubKey, k.keyType)
}

// Sign signs the given payload.
func (k PrivateKey) Sign(payload []byte) ([]byte, error) {
	return k.privKey.Sign(payload)
}

// PublicKey is a public key.
type PublicKey struct {
	pubKey  ic.PubKey
	keyType pb.KeyType
}

// NewPublicKey creates a PublicKey from a libp2p one.
func NewPublicKey(pubKey ic.PubKey, keyType pb.KeyType) *PublicKey {
	return &PublicKey{
		pubKey:  pubKey,
		keyType: keyType,
	}
}

// Bytes returns the key bytes.
func (k PublicKey) Bytes() ([]byte, error) {
	return k.pubKey.Bytes()
}

// Verify verifies a signature.
func (k PublicKey) Verify(payload, sig []byte) (bool, error) {
	return k.pubKey.Verify(payload, sig)
}
