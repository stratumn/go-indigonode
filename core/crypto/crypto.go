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
	"context"

	"github.com/pkg/errors"
	"github.com/stratumn/go-indigonode/core/monitoring"

	ic "gx/ipfs/Qme1knMqwt1hKZbc1BmQFmnm9f36nyQGwXxPGVpVJ9rMK5/go-libp2p-crypto"
)

var (
	// ErrInvalidKeyType is returned when the cryptographic key scheme isn't supported.
	ErrInvalidKeyType = errors.New("given key type isn't supported yet")
)

// Sign signs the given payload with the given private key.
func Sign(ctx context.Context, sk ic.PrivKey, payload []byte) (*Signature, error) {
	_, span := monitoring.StartSpan(ctx, "crypto", "Sign")
	defer span.End()

	pkBytes, err := sk.GetPublic().Bytes()
	if err != nil {
		span.SetStatus(monitoring.NewStatus(monitoring.StatusCodeInvalidArgument, err.Error()))
		return nil, errors.WithStack(err)
	}

	signatureBytes, err := sk.Sign(payload)
	if err != nil {
		span.SetUnknownError(err)
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
		span.SetStatus(monitoring.NewStatus(monitoring.StatusCodeInvalidArgument, ErrInvalidKeyType.Error()))
		return nil, ErrInvalidKeyType
	}

	return &Signature{
		KeyType:   keyType,
		PublicKey: pkBytes,
		Signature: signatureBytes,
	}, nil
}

// Verify verifies a signature.
func (s *Signature) Verify(ctx context.Context, payload []byte) (ok bool) {
	_, span := monitoring.StartSpan(ctx, "crypto", "Verify")
	defer func() {
		span.AddBoolAttribute("signature_ok", ok)
		span.End()
	}()

	if s == nil {
		return false
	}

	pubKey, err := ic.UnmarshalPublicKey(s.PublicKey)
	if err != nil {
		return false
	}

	ok, err = pubKey.Verify(payload, s.Signature)
	if err != nil {
		return false
	}

	return ok
}
