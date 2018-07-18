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

	"github.com/stratumn/go-indigonode/app/indigo/protocol/store/audit"
	"github.com/stratumn/go-indigonode/app/indigo/protocol/store/constants"
	"github.com/stratumn/go-indigocore/cs"
	"github.com/stratumn/go-indigocore/cs/cstesting"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	peer "gx/ipfs/QmcJukH2sAFjY3HdBKq35WDzWoL3UUu2gt9wdfqZTUyM74/go-libp2p-peer"
	ic "gx/ipfs/Qme1knMqwt1hKZbc1BmQFmnm9f36nyQGwXxPGVpVJ9rMK5/go-libp2p-crypto"
)

// TestAddSegment runs tests on the writer.AddSegment method.
func (f Factory) TestAddSegment(t *testing.T) {
	ctx := context.Background()

	store, err := f.New()
	require.NoError(t, err)
	defer f.Free(store)

	sk, _, _ := ic.GenerateEd25519Key(rand.Reader)
	peerID, _ := peer.IDFromPrivateKey(sk)

	t.Run("saves a segment and its evidence", func(t *testing.T) {
		link1 := cstesting.NewLinkBuilder().
			WithMetadata(constants.NodeIDKey, peerID.Pretty()).
			Build()
		segment1, _ := audit.SignLink(ctx, sk, link1)

		link2 := cstesting.NewLinkBuilder().
			WithMetadata(constants.NodeIDKey, peerID.Pretty()).
			Build()
		segment2, _ := audit.SignLink(ctx, sk, link2)

		assert.NoError(t, store.AddSegment(ctx, segment1), "store.AddSegment()")
		assert.NoError(t, store.AddSegment(ctx, segment2), "store.AddSegment()")

		segments, err := store.GetByPeer(ctx, peerID, audit.Pagination{})
		require.NoError(t, err, "store.GetByPeer()")
		require.Len(t, segments, 2)
		assert.Equal(t, segment1, segments[0])
		assert.Equal(t, segment2, segments[1])
	})

	t.Run("updates a segment when a new evidence is received", func(t *testing.T) {
		sk2, _, _ := ic.GenerateEd25519Key(rand.Reader)
		peerID2, _ := peer.IDFromPrivateKey(sk2)

		link1 := cstesting.NewLinkBuilder().
			WithMetadata(constants.NodeIDKey, peerID2.Pretty()).
			Build()
		segment1, _ := audit.SignLink(ctx, sk2, link1)
		assert.NoError(t, store.AddSegment(ctx, segment1), "store.AddSegment()")

		assert.NoError(t, segment1.Meta.AddEvidence(cs.Evidence{
			Backend:  "generic",
			Provider: "test",
		}))
		assert.NoError(t, store.AddSegment(ctx, segment1), "store.AddSegment()")

		segments, err := store.GetByPeer(ctx, peerID2, audit.NewDefaultPagination())
		require.NoError(t, err, "store.GetByPeer()")
		require.Len(t, segments, 1)
		require.Len(t, segments[0].Meta.Evidences, 2)
	})

	t.Run("returns an error when peerID is missing from the link", func(t *testing.T) {
		link3 := cstesting.NewLinkBuilder().
			WithMetadata(constants.NodeIDKey, peerID.Pretty()).
			Build()
		segment3, _ := audit.SignLink(ctx, sk, link3)
		delete(segment3.Link.Meta.Data, constants.NodeIDKey)
		assert.EqualError(t, store.AddSegment(ctx, segment3), constants.ErrInvalidMetaNodeID.Error())
	})

}
