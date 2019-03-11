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

package crypto

import (
	"context"

	"github.com/pkg/errors"
	"github.com/stratumn/go-node/core/monitoring"

	ic "github.com/libp2p/go-libp2p-crypto"
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
