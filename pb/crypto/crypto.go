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

package crypto

import (
	"github.com/pkg/errors"

	ic "gx/ipfs/QmaPbCnUMBohSGo3KnxEa2bHqyJVVeEEcwtqJAYxerieBo/go-libp2p-crypto"
)

var (
	// ErrInvalidKeyType is returned when the cryptographic key scheme isn't supported.
	ErrInvalidKeyType = errors.New("given key type isn't supported yet")
)

// Sign signs the given payload with the given private key.
func Sign(sk ic.PrivKey, payload []byte) (*Signature, error) {
	pkBytes, err := sk.GetPublic().Bytes()
	if err != nil {
		return nil, errors.WithStack(err)
	}

	signatureBytes, err := sk.Sign(payload)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	var keyType KeyType
	switch sk.(type) {
	case *ic.Ed25519PrivateKey:
		keyType = KeyType_Ed25519
	case *ic.Secp256k1PrivateKey:
		keyType = KeyType_Secp256k1
	case *ic.RsaPrivateKey:
		keyType = KeyType_RSA
	default:
		return nil, ErrInvalidKeyType
	}

	return &Signature{
		KeyType:   keyType,
		PublicKey: pkBytes,
		Signature: signatureBytes,
	}, nil
}

// Verify verifies a signature.
func (s *Signature) Verify(payload []byte) bool {
	if s == nil {
		return false
	}

	pubKey, err := ic.UnmarshalPublicKey(s.PublicKey)
	if err != nil {
		return false
	}

	valid, err := pubKey.Verify(payload, s.Signature)
	if err != nil {
		return false
	}

	return valid
}
