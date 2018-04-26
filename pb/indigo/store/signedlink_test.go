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

package store_test

import (
	"testing"

	"github.com/stratumn/alice/pb/indigo/store"
	"github.com/stratumn/go-indigocore/cs/cstesting"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	peer "gx/ipfs/QmZoWKhxUmZ2seW4BzX6fJkNR8hh9PsGModr7q171yq2SS/go-libp2p-peer"
	ic "gx/ipfs/QmaPbCnUMBohSGo3KnxEa2bHqyJVVeEEcwtqJAYxerieBo/go-libp2p-crypto"
)

var (
	sk1     ic.PrivKey
	peerID1 peer.ID

	sk2     ic.PrivKey
	peerID2 peer.ID
)

func init() {
	var err error
	sk1Bytes, err := ic.ConfigDecodeKey("CAESYA8c9Ei8atnEXpODNga9OrXUBDjPhpEr3Zf6DYwsVmLfzBY65CWfDdpbDHOZZE+nB7W/K4b2RLuohkGx/a1JwYzMFjrkJZ8N2lsMc5lkT6cHtb8rhvZEu6iGQbH9rUnBjA==")
	if err != nil {
		panic(err)
	}

	sk1, err = ic.UnmarshalPrivateKey(sk1Bytes)
	if err != nil {
		panic(err)
	}

	peerID1, err = peer.IDB58Decode("QmPEeCgxxX6YbQWqkKuF42YCUpy4GdrqGLPMAFZ8A3A35d")
	if err != nil {
		panic(err)
	}

	if !peerID1.MatchesPrivateKey(sk1) {
		panic("peerID / secret key mismatch")
	}

	sk2Bytes, err := ic.ConfigDecodeKey("CAESYKecc4tj7XAXruOYfd4m61d3mvxJUUdUVwIuFbB/PYFAtAoPM/Pbft/aS3mc5jFkb2dScZS61XOl9PnU3uDWuPq0Cg8z89t+39pLeZzmMWRvZ1JxlLrVc6X0+dTe4Na4+g==")
	if err != nil {
		panic(err)
	}

	sk2, err = ic.UnmarshalPrivateKey(sk2Bytes)
	if err != nil {
		panic(err)
	}

	peerID2, err = peer.IDB58Decode("QmeZjNhdKPNNEtCbmL6THvMfTRPZMgC1wfYe9s3DdoQZcM")
	if err != nil {
		panic(err)
	}

	if !peerID2.MatchesPrivateKey(sk2) {
		panic("peerID / secret key mismatch")
	}
}

func TestNewSignedLink(t *testing.T) {
	link := cstesting.RandomLink()

	tests := []struct {
		name           string
		signLink       func(*testing.T) *store.SignedLink
		testSignedLink func(*testing.T, *store.SignedLink)
		valid          bool
	}{{
		"from-peer",
		func(t *testing.T) *store.SignedLink {
			sl, err := store.NewSignedLink(sk1, link)
			require.NoError(t, err, "store.NewSignedLink()")
			return sl
		},
		func(t *testing.T, sl *store.SignedLink) {
			assert.Equal(t, []byte(peerID1), sl.From)
		},
		true,
	}, {
		"from-mismatch",
		func(t *testing.T) *store.SignedLink {
			sl, err := store.NewSignedLink(sk1, link)
			require.NoError(t, err, "store.NewSignedLink()")

			sl.From = []byte(peerID2)
			return sl
		},
		func(*testing.T, *store.SignedLink) {},
		false,
	}, {
		"signer-mismatch",
		func(t *testing.T) *store.SignedLink {
			sl, err := store.NewSignedLink(sk1, link)
			require.NoError(t, err, "store.NewSignedLink()")

			sl.From = []byte(peerID2)
			sl.Signature.PublicKey, _ = sk2.GetPublic().Bytes()

			return sl
		},
		func(*testing.T, *store.SignedLink) {},
		false,
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sl := tt.signLink(t)
			tt.testSignedLink(t, sl)
			assert.Equal(t, tt.valid, sl.VerifySignature(), "VerifySignature()")
		})
	}
}
