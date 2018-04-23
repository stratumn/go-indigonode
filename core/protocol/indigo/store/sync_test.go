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
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stratumn/alice/core/protocol/indigo/store"
	"github.com/stratumn/alice/core/service/indigo/store/mockstore"
	"github.com/stratumn/go-indigocore/cs"
	"github.com/stratumn/go-indigocore/cs/cstesting"
	"github.com/stratumn/go-indigocore/dummystore"
	"github.com/stretchr/testify/assert"

	bhost "gx/ipfs/QmQr1j6UvdhpponAaqSdswqRpdzsFwNop2N8kXLNw8afem/go-libp2p-blankhost"
	netutil "gx/ipfs/QmYVR3C8DWPHdHxvLtNFYfjsXgaRAdh6hPMNH3KiwCgu4o/go-libp2p-netutil"
)

func TestSynchronousSyncEngine_New(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	h := mockstore.NewMockHost(ctrl)
	h.EXPECT().SetStreamHandler(store.SynchronousSyncProtocolID, gomock.Any()).Times(1)
	h.EXPECT().RemoveStreamHandler(store.SynchronousSyncProtocolID).Times(1)

	engine := store.NewSynchronousSyncEngine(h, nil)
	engine.Close(ctx)
}

func TestSynchronousSyncEngine_GetMissingLinks(t *testing.T) {
	t.Run("node-not-connected", func(t *testing.T) {
		ctx := context.Background()

		h := bhost.NewBlankHost(netutil.GenSwarmNetwork(t, ctx))
		defer h.Close()

		indigoStore := dummystore.New(&dummystore.Config{})
		engine := store.NewSynchronousSyncEngine(h, indigoStore)
		defer engine.Close(ctx)

		_, err := engine.GetMissingLinks(ctx, cstesting.RandomLink(), indigoStore)
		assert.EqualError(t, err, store.ErrNoConnectedPeers.Error())
	})

	t.Run("simple-one-pass-sync", func(t *testing.T) {
		ctx := context.Background()

		h1 := bhost.NewBlankHost(netutil.GenSwarmNetwork(t, ctx))
		h2 := bhost.NewBlankHost(netutil.GenSwarmNetwork(t, ctx))
		defer h1.Close()
		defer h2.Close()

		assert.NoError(t, h1.Connect(ctx, h2.Peerstore().PeerInfo(h2.ID())))

		// prevLink <--------- link
		// refLink1 <------|
		// refLink2 <------|
		prevLink := cstesting.RandomLink()
		prevLink.Meta.PrevLinkHash = ""
		prevLinkHash, _ := prevLink.Hash()

		refLink1 := cstesting.RandomLink()
		refLink1.Meta.PrevLinkHash = ""
		refLinkHash1, _ := refLink1.Hash()

		refLink2 := cstesting.RandomLink()
		refLink2.Meta.PrevLinkHash = ""
		refLinkHash2, _ := refLink2.Hash()

		link := cstesting.RandomLink()
		link.Meta.PrevLinkHash = prevLinkHash.String()
		link.Meta.Refs = []cs.SegmentReference{
			{LinkHash: refLinkHash1.String()},
			{LinkHash: refLinkHash2.String()},
		}

		// Node 1 has refLink2
		store1 := dummystore.New(&dummystore.Config{})
		store1.CreateLink(ctx, refLink2)
		// Node 2 contains all references except one.
		store2 := dummystore.New(&dummystore.Config{})
		store2.CreateLink(ctx, prevLink)
		store2.CreateLink(ctx, refLink1)

		engine1 := store.NewSynchronousSyncEngine(h1, store1)
		engine2 := store.NewSynchronousSyncEngine(h2, store2)
		defer engine1.Close(ctx)
		defer engine2.Close(ctx)

		links, err := engine1.GetMissingLinks(ctx, link, store1)
		assert.NoError(t, err)
		assert.Len(t, links, 2)
		assert.Contains(t, links, prevLink)
		assert.Contains(t, links, refLink1)
	})

	t.Run("link-missing-from-all-peers", func(t *testing.T) {
		ctx := context.Background()

		h1 := bhost.NewBlankHost(netutil.GenSwarmNetwork(t, ctx))
		h2 := bhost.NewBlankHost(netutil.GenSwarmNetwork(t, ctx))
		defer h1.Close()
		defer h2.Close()

		assert.NoError(t, h1.Connect(ctx, h2.Peerstore().PeerInfo(h2.ID())))

		// prevLink <--------- link
		// refLink <------|
		prevLink := cstesting.RandomLink()
		prevLink.Meta.PrevLinkHash = ""
		prevLinkHash, _ := prevLink.Hash()

		refLink := cstesting.RandomLink()
		refLink.Meta.PrevLinkHash = ""
		refLinkHash, _ := refLink.Hash()

		link := cstesting.RandomLink()
		link.Meta.PrevLinkHash = prevLinkHash.String()
		link.Meta.Refs = []cs.SegmentReference{
			{LinkHash: refLinkHash.String()},
		}

		// Node 1 has nothing locally.
		store1 := dummystore.New(&dummystore.Config{})
		// Node 2 contains all references except one.
		store2 := dummystore.New(&dummystore.Config{})
		store2.CreateLink(ctx, prevLink)

		engine1 := store.NewSynchronousSyncEngine(h1, store1)
		engine2 := store.NewSynchronousSyncEngine(h2, store2)
		defer engine1.Close(ctx)
		defer engine2.Close(ctx)

		_, err := engine1.GetMissingLinks(ctx, link, store1)
		assert.EqualError(t, err, store.ErrLinkNotFound.Error())
	})

	t.Run("recursive-multinode-sync", func(t *testing.T) {
		ctx := context.Background()

		h1 := bhost.NewBlankHost(netutil.GenSwarmNetwork(t, ctx))
		h2 := bhost.NewBlankHost(netutil.GenSwarmNetwork(t, ctx))
		h3 := bhost.NewBlankHost(netutil.GenSwarmNetwork(t, ctx))
		defer h1.Close()
		defer h2.Close()
		defer h3.Close()

		assert.NoError(t, h1.Connect(ctx, h2.Peerstore().PeerInfo(h2.ID())))
		assert.NoError(t, h1.Connect(ctx, h3.Peerstore().PeerInfo(h3.ID())))

		prevPrevLink := cstesting.RandomLink()
		prevPrevLink.Meta.PrevLinkHash = ""
		prevPrevLinkHash, _ := prevPrevLink.Hash()

		prevLink := cstesting.RandomLink()
		prevLink.Meta.PrevLinkHash = ""
		prevLink.Meta.Refs = []cs.SegmentReference{{LinkHash: prevPrevLinkHash.String()}}
		prevLinkHash, _ := prevLink.Hash()

		prevRefLink := cstesting.RandomLink()
		prevRefLink.Meta.PrevLinkHash = ""
		prevRefLinkHash, _ := prevRefLink.Hash()

		refLink := cstesting.RandomLink()
		refLink.Meta.PrevLinkHash = prevRefLinkHash.String()
		refLinkHash, _ := refLink.Hash()

		link := cstesting.RandomLink()
		link.Meta.PrevLinkHash = prevLinkHash.String()
		link.Meta.Refs = []cs.SegmentReference{{LinkHash: refLinkHash.String()}}

		store1 := dummystore.New(&dummystore.Config{})
		store1.CreateLink(ctx, prevRefLink)

		store2 := dummystore.New(&dummystore.Config{})
		store2.CreateLink(ctx, prevLink)

		store3 := dummystore.New(&dummystore.Config{})
		store3.CreateLink(ctx, refLink)
		store3.CreateLink(ctx, prevPrevLink)

		engine1 := store.NewSynchronousSyncEngine(h1, store1)
		engine2 := store.NewSynchronousSyncEngine(h2, store2)
		engine3 := store.NewSynchronousSyncEngine(h3, store3)
		defer engine1.Close(ctx)
		defer engine2.Close(ctx)
		defer engine3.Close(ctx)

		links, err := engine1.GetMissingLinks(ctx, link, store1)
		assert.NoError(t, err)
		assert.Len(t, links, 3)
		assert.Contains(t, links, prevPrevLink)
		assert.Contains(t, links, prevLink)
		assert.Contains(t, links, refLink)

		// To comply with dependency ordering, prevPrevLink should appear
		// before prevLink.
		assert.True(t, getLinkIndex(prevPrevLink, links) < getLinkIndex(prevLink, links))
	})

	t.Run("reference-segment-included", func(t *testing.T) {
		ctx := context.Background()

		h1 := bhost.NewBlankHost(netutil.GenSwarmNetwork(t, ctx))
		h2 := bhost.NewBlankHost(netutil.GenSwarmNetwork(t, ctx))
		defer h1.Close()
		defer h2.Close()

		assert.NoError(t, h1.Connect(ctx, h2.Peerstore().PeerInfo(h2.ID())))

		prevRefLink := cstesting.RandomLink()
		prevRefLink.Meta.PrevLinkHash = ""
		prevRefLinkHash, _ := prevRefLink.Hash()

		// Referenced link fully included in prevLink.
		// Will not be stored anywhere.
		refLink := cstesting.RandomLink()
		refLink.Meta.PrevLinkHash = prevRefLinkHash.String()

		prevLink := cstesting.RandomLink()
		prevLink.Meta.PrevLinkHash = ""
		prevLink.Meta.Refs = []cs.SegmentReference{
			{Segment: refLink.Segmentify()},
		}
		prevLinkHash, _ := prevLink.Hash()

		link := cstesting.RandomLink()
		link.Meta.PrevLinkHash = prevLinkHash.String()

		store1 := dummystore.New(&dummystore.Config{})
		store2 := dummystore.New(&dummystore.Config{})
		store2.CreateLink(ctx, prevLink)
		store2.CreateLink(ctx, prevRefLink)

		engine1 := store.NewSynchronousSyncEngine(h1, store1)
		engine2 := store.NewSynchronousSyncEngine(h2, store2)
		defer engine1.Close(ctx)
		defer engine2.Close(ctx)

		links, err := engine1.GetMissingLinks(ctx, link, store1)
		assert.NoError(t, err)
		assert.Len(t, links, 3)
		assert.Equal(t, prevRefLink, links[0])
		assert.Equal(t, refLink, links[1])
		assert.Equal(t, prevLink, links[2])
	})
}

func getLinkIndex(link *cs.Link, links []*cs.Link) int {
	linkHash, _ := link.HashString()
	for i, l := range links {
		lh, _ := l.HashString()
		if linkHash == lh {
			return i
		}
	}

	return -1
}
