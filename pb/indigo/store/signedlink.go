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

package store

import (
	json "github.com/gibson042/canonicaljson-go"
	"github.com/pkg/errors"
	"github.com/stratumn/alice/pb/crypto"
	"github.com/stratumn/go-indigocore/cs"

	peer "gx/ipfs/QmZoWKhxUmZ2seW4BzX6fJkNR8hh9PsGModr7q171yq2SS/go-libp2p-peer"
	ic "gx/ipfs/QmaPbCnUMBohSGo3KnxEa2bHqyJVVeEEcwtqJAYxerieBo/go-libp2p-crypto"
)

// NewSignedLink creates a new signed link from a private key and a link.
func NewSignedLink(sk ic.PrivKey, link *cs.Link) (*SignedLink, error) {
	peerID, err := peer.IDFromPrivateKey(sk)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	linkBytes, err := json.Marshal(link)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	sig, err := crypto.Sign(sk, linkBytes)
	if err != nil {
		return nil, err
	}

	signed := &SignedLink{
		From:      []byte(peerID),
		Link:      linkBytes,
		Signature: sig,
	}

	return signed, nil
}

// VerifySignature verifies the signature of a signed link.
func (sl *SignedLink) VerifySignature() bool {
	if sl == nil || sl.Signature == nil {
		return false
	}

	fromPeerID, err := peer.IDFromBytes(sl.From)
	if err != nil {
		return false
	}

	pubKey, err := ic.UnmarshalPublicKey(sl.Signature.PublicKey)
	if err != nil {
		return false
	}

	if !fromPeerID.MatchesPublicKey(pubKey) {
		return false
	}

	return sl.Signature.Verify(sl.Link)
}
