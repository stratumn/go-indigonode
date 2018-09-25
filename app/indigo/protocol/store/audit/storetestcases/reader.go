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

package storetestcases

import (
	"context"
	"crypto/rand"
	"testing"

	"github.com/stratumn/go-indigocore/cs"
	"github.com/stratumn/go-indigocore/cs/cstesting"
	"github.com/stratumn/go-node/app/indigo/protocol/store/audit"
	"github.com/stratumn/go-node/app/indigo/protocol/store/constants"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	ic "gx/ipfs/QmPvyPwuCgJ7pDmrKDxRtsScJgBaM5h4EpRL2qQJsmXf4n/go-libp2p-crypto"
	peer "gx/ipfs/QmQsErDt8Qgw1XrsXf2BpEzDgGWtB1YLsTAARBup5b6B9W/go-libp2p-peer"
)

// TestGetByPeer runs tests on the reader.GetByPeer method.
func (f Factory) TestGetByPeer(t *testing.T) {
	ctx := context.Background()

	store, err := f.New()
	require.NoError(t, err)
	defer f.Free(store)

	sk1, _, _ := ic.GenerateEd25519Key(rand.Reader)
	peer1, _ := peer.IDFromPrivateKey(sk1)

	sk2, _, _ := ic.GenerateEd25519Key(rand.Reader)
	peer2, _ := peer.IDFromPrivateKey(sk2)

	for i := 0; i < 10; i++ {
		key := sk1
		peerID := peer1
		if i%2 == 0 {
			key = sk2
			peerID = peer2
		}

		link := cstesting.NewLinkBuilder().Build()
		constants.SetLinkNodeID(link, peerID)
		s, _ := audit.SignLink(ctx, key, link)
		require.NoError(
			t,
			store.AddSegment(ctx, s),
			"store.AddSegment()",
		)
	}

	tests := []struct {
		name string
		run  func(*testing.T)
	}{{
		"peer-not-found",
		func(t *testing.T) {
			segments, err := store.GetByPeer(ctx, "Sp0nG3b0B", audit.Pagination{})
			assert.NoError(t, err, "store.GetByPeer()")
			assert.Empty(t, segments, "store.GetByPeer()")
		},
	}, {
		"small-pagination",
		func(t *testing.T) {
			segmentsPage1, err := store.GetByPeer(ctx, peer1, audit.Pagination{Top: 3})
			assert.NoError(t, err, "store.GetByPeer()")
			require.Len(t, segmentsPage1, 3, "store.GetByPeer()")

			segmentsPage2, err := store.GetByPeer(ctx, peer1, audit.Pagination{Top: 3, Skip: 2})
			assert.NoError(t, err, "store.GetByPeer()")
			require.Len(t, segmentsPage2, 3, "store.GetByPeer()")

			assert.Equal(t, segmentsPage1[2], segmentsPage2[0])
			assert.False(t, ContainsSegment(segmentsPage2, segmentsPage1[0]))
			assert.False(t, ContainsSegment(segmentsPage2, segmentsPage1[1]))
		},
	}, {
		"big-pagination",
		func(t *testing.T) {
			segments, err := store.GetByPeer(ctx, peer2, audit.Pagination{Top: 10})
			assert.NoError(t, err, "store.GetByPeer()")
			assert.NotEmpty(t, segments, "store.GetByPeer()")
			assert.Len(t, segments, 5, "store.GetByPeer()")
		},
	}, {
		"out-of-bounds",
		func(t *testing.T) {
			segments, err := store.GetByPeer(ctx, peer1, audit.Pagination{Skip: 10, Top: 3})
			assert.NoError(t, err, "store.GetByPeer()")
			assert.Empty(t, segments, "store.GetByPeer()")
		},
	}, {
		"no-pagination",
		func(t *testing.T) {
			segments, err := store.GetByPeer(ctx, peer1, audit.Pagination{})
			assert.NoError(t, err, "store.GetByPeer()")
			assert.Len(t, segments, 5, "store.GetByPeer()")
		},
	}, {
		"all-evidence-types",
		func(t *testing.T) {
			sk3, _, _ := ic.GenerateEd25519Key(rand.Reader)
			peer3, _ := peer.IDFromPrivateKey(sk3)
			link := cstesting.NewLinkBuilder().Build()
			constants.SetLinkNodeID(link, peer3)
			s, _ := audit.SignLink(ctx, sk3, link)

			assert.NoError(t, s.Meta.AddEvidence(cs.Evidence{Backend: "generic", Provider: "1"}))
			assert.NoError(t, s.Meta.AddEvidence(cs.Evidence{Backend: "batch", Provider: "2"}))
			assert.NoError(t, s.Meta.AddEvidence(cs.Evidence{Backend: "bcbatch", Provider: "3"}))
			assert.NoError(t, s.Meta.AddEvidence(cs.Evidence{Backend: "dummy", Provider: "4"}))

			require.NoError(
				t,
				store.AddSegment(ctx, s),
				"store.AddSegment()",
			)

			segments, err := store.GetByPeer(ctx, peer3, audit.NewDefaultPagination())
			assert.NoError(t, err, "store.GetByPeer()")
			require.Len(t, segments, 1, "store.GetByPeer()")
			require.Len(t, segments[0].Meta.Evidences, 5, "store.GetByPeer()")
		},
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.run(t)
		})
	}
}
