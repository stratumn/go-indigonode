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

package coinutil

import (
	"github.com/stratumn/go-indigonode/app/coin/pb"

	ic "gx/ipfs/QmPvyPwuCgJ7pDmrKDxRtsScJgBaM5h4EpRL2qQJsmXf4n/go-libp2p-crypto"
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
