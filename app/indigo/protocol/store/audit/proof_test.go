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

package audit_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/pkg/errors"
	"github.com/stratumn/go-indigocore/cs"
	"github.com/stratumn/go-indigocore/cs/cstesting"
	"github.com/stratumn/go-indigonode/app/indigo/protocol/store/audit"
	"github.com/stratumn/go-indigonode/app/indigo/protocol/store/constants"
	"github.com/stratumn/go-indigonode/core/crypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	peer "gx/ipfs/QmcJukH2sAFjY3HdBKq35WDzWoL3UUu2gt9wdfqZTUyM74/go-libp2p-peer"
	ic "gx/ipfs/Qme1knMqwt1hKZbc1BmQFmnm9f36nyQGwXxPGVpVJ9rMK5/go-libp2p-crypto"
)

var (
	sk1     ic.PrivKey
	peerID1 peer.ID

	sk2     ic.PrivKey
	peerID2 peer.ID
)

func init() {
	var err error
	sk1Bytes, err := ic.ConfigDecodeKey("CAESYKnY+DjP117lX9R3c79j7aNy3nw/kgl9Qsj8UqDuQuV3njK+2GS0+CzP1DsDphBxyVLNJcy+f/Io+U9GKNkzL1KeMr7YZLT4LM/UOwOmEHHJUs0lzL5/8ij5T0Yo2TMvUg==")
	if err != nil {
		panic(err)
	}

	sk1, err = ic.UnmarshalPrivateKey(sk1Bytes)
	if err != nil {
		panic(err)
	}

	peerID1, err = peer.IDB58Decode("12D3KooWLTuTuFWdyPMqayf8oFxBTW1q2aSTLZHQ3HeaW84dMpKb")
	if err != nil {
		panic(err)
	}

	if !peerID1.MatchesPrivateKey(sk1) {
		panic("peerID / secret key mismatch")
	}

	sk2Bytes, err := ic.ConfigDecodeKey("CAESYHwyHBAE7vMstcqNNllwvuG9Fd9tzuoYhZx5kVZDlBe585qva8NR6XpPLOxtLxCLcb1TRJ5Ku0WgA0I9as4/fx7zmq9rw1Hpek8s7G0vEItxvVNEnkq7RaADQj1qzj9/Hg==")
	if err != nil {
		panic(err)
	}

	sk2, err = ic.UnmarshalPrivateKey(sk2Bytes)
	if err != nil {
		panic(err)
	}

	peerID2, err = peer.IDB58Decode("12D3KooWSDJ3VWhFsHJKkCTq7ALuv4v8SdnbjmAMS8PmrbDJqpih")
	if err != nil {
		panic(err)
	}

	if !peerID2.MatchesPrivateKey(sk2) {
		panic("peerID / secret key mismatch")
	}
}

func TestSignLink(t *testing.T) {
	ctx := context.Background()
	link := cstesting.NewLinkBuilder().
		WithMetadata(constants.NodeIDKey, peerID1.Pretty()).
		Build()

	tests := []struct {
		name     string
		sk       ic.PrivKey
		link     *cs.Link
		validate func(*testing.T, *cs.Segment)
		err      error
	}{{
		"missing-private-key",
		nil,
		link,
		nil,
		errors.New("secret key or link missing"),
	}, {
		"missing-link",
		sk1,
		nil,
		nil,
		errors.New("secret key or link missing"),
	}, {
		"missing-link-node-id",
		sk1,
		cstesting.RandomLink(),
		nil,
		constants.ErrInvalidMetaNodeID,
	}, {
		"add-valid-evidence",
		sk1,
		link,
		func(t *testing.T, segment *cs.Segment) {
			assert.Len(t, segment.Meta.Evidences, 1)

			e := segment.Meta.Evidences[0]
			assert.Equal(t, audit.PeerSignatureBackend, e.Backend)
			assert.Equal(t, peerID1.Pretty(), e.Provider)
			assert.True(t, e.Proof.Verify(segment.GetLinkHash()[:]))
		},
		nil,
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			segment, err := audit.SignLink(ctx, tt.sk, tt.link)
			if tt.err != nil {
				assert.EqualError(t, err, tt.err.Error())
			} else {
				assert.NoError(t, err)
				tt.validate(t, segment)
			}
		})
	}
}

func TestPeerSignature_New(t *testing.T) {
	link := cstesting.NewLinkBuilder().
		WithMetadata(constants.NodeIDKey, peerID1.Pretty()).
		Build()
	linkHash, _ := link.Hash()

	proof, err := audit.NewPeerSignature(context.Background(), sk1, link.Segmentify())
	assert.NoError(t, err)

	sigProof := proof.(*audit.PeerSignature)
	assert.Equal(t, linkHash[:], sigProof.LinkHash, "proof.LinkHash")
	assert.Equal(t, []byte(peerID1), sigProof.PeerID, "proof.PeerID")

	sig := sigProof.Signature
	assert.NotNil(t, sig, "proof.Signature")
	assert.Equal(t, crypto.KeyType_Ed25519, sig.KeyType, "sig.KeyType")
	assert.True(t, sig.Verify(context.Background(), linkHash[:]), "sig.Verify()")
	assert.True(t, proof.Verify(linkHash[:]), "proof.Verify()")
}

func TestPeerSignature_Verify(t *testing.T) {
	t.Run("peer-id-mismatch", func(t *testing.T) {
		segment := cstesting.NewLinkBuilder().
			WithMetadata(constants.NodeIDKey, peerID1.Pretty()).
			Build().
			Segmentify()
		proof, _ := audit.NewPeerSignature(context.Background(), sk1, segment)
		proof.(*audit.PeerSignature).PeerID = []byte(peerID2)

		assert.False(t, proof.Verify(segment.GetLinkHash()[:]))
	})

	t.Run("link-hash-mismatch", func(t *testing.T) {
		segment := cstesting.NewLinkBuilder().
			WithMetadata(constants.NodeIDKey, peerID1.Pretty()).
			Build().
			Segmentify()
		proof, _ := audit.NewPeerSignature(context.Background(), sk1, segment)

		assert.False(t, proof.Verify([]byte("hello")))
	})

	t.Run("signature-mismatch", func(t *testing.T) {
		s1 := cstesting.NewLinkBuilder().
			WithMetadata(constants.NodeIDKey, peerID1.Pretty()).
			Build().
			Segmentify()
		s2 := cstesting.NewLinkBuilder().
			WithMetadata(constants.NodeIDKey, peerID1.Pretty()).
			Build().
			Segmentify()

		proof1, _ := audit.NewPeerSignature(context.Background(), sk1, s1)
		proof2, _ := audit.NewPeerSignature(context.Background(), sk1, s2)

		// Swap the signatures. The signature will be valid, but won't
		// sign the right link hash.
		proof1.(*audit.PeerSignature).Signature = proof2.(*audit.PeerSignature).Signature

		assert.False(t, proof1.Verify(s1.GetLinkHash()[:]))
	})

	t.Run("valid-proof", func(t *testing.T) {
		segment := cstesting.NewLinkBuilder().
			WithMetadata(constants.NodeIDKey, peerID1.Pretty()).
			Build().
			Segmentify()
		proof, _ := audit.NewPeerSignature(context.Background(), sk1, segment)

		assert.True(t, proof.Verify(segment.GetLinkHash()[:]))
	})
}

func TestPeerSignature_Marshal(t *testing.T) {
	ctx := context.Background()
	link := cstesting.NewLinkBuilder().
		WithMetadata(constants.NodeIDKey, peerID2.Pretty()).
		Build()

	segment, err := audit.SignLink(ctx, sk2, link)
	require.NoError(t, err, "audit.SignLink()")

	segmentJSON, err := json.Marshal(segment)
	require.NoError(t, err, "json.Marshal()")

	var unmarshalled cs.Segment
	require.NoError(t, json.Unmarshal(segmentJSON, &unmarshalled), "json.Unmarshal()")

	assert.Len(t, unmarshalled.Meta.Evidences, 1)
	assert.Equal(t, *segment, unmarshalled)
}
