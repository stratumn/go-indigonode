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
	indigostore "github.com/stratumn/go-indigocore/store"
	"github.com/stretchr/testify/assert"

	ihost "gx/ipfs/QmNmJZL7FQySMtE2BQuLMuZg2EB2CLEunJJUSVSc9YnnbV/go-libp2p-host"
	bhost "gx/ipfs/QmQr1j6UvdhpponAaqSdswqRpdzsFwNop2N8kXLNw8afem/go-libp2p-blankhost"
	netutil "gx/ipfs/QmYVR3C8DWPHdHxvLtNFYfjsXgaRAdh6hPMNH3KiwCgu4o/go-libp2p-netutil"
	protocol "gx/ipfs/QmZNkThpqfVXs9GNbexPrfBbXSLNYeKrE7jwFM2oqHbyqN/go-libp2p-protocol"
)

func TestListMissingLinkHashes(t *testing.T) {
	ctx := context.Background()

	testCases := []struct {
		name  string
		setup func() (*cs.Link, indigostore.SegmentReader, []string)
	}{{
		"nothing-missing",
		func() (*cs.Link, indigostore.SegmentReader, []string) {
			prevLink := cstesting.NewLinkBuilder().Build()
			refLink := cstesting.NewLinkBuilder().Build()
			link := cstesting.NewLinkBuilder().
				WithParent(prevLink).
				WithRef(refLink).
				Build()

			s := dummystore.New(&dummystore.Config{})
			s.CreateLink(ctx, prevLink)
			s.CreateLink(ctx, refLink)

			return link, s, []string{}
		},
	}, {
		"some-missing",
		func() (*cs.Link, indigostore.SegmentReader, []string) {
			prevLink := cstesting.NewLinkBuilder().Build()
			prevLinkHash, _ := prevLink.HashString()
			refLink1 := cstesting.NewLinkBuilder().Build()
			refLink2 := cstesting.NewLinkBuilder().Build()
			refLinkHash2, _ := refLink2.HashString()
			link := cstesting.NewLinkBuilder().
				WithParent(prevLink).
				WithRef(refLink1).
				WithRef(refLink2).
				Build()

			s := dummystore.New(&dummystore.Config{})
			s.CreateLink(ctx, refLink1)

			return link, s, []string{prevLinkHash, refLinkHash2}
		},
	}, {
		"duplicate-link-hash",
		func() (*cs.Link, indigostore.SegmentReader, []string) {
			prevLink := cstesting.NewLinkBuilder().Build()
			prevLinkHash, _ := prevLink.HashString()
			link := cstesting.NewLinkBuilder().
				WithParent(prevLink).
				WithRef(prevLink).
				Build()

			s := dummystore.New(&dummystore.Config{})

			return link, s, []string{prevLinkHash}
		},
	}}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			link, reader, expected := tt.setup()
			lhs, err := store.ListMissingLinkHashes(context.Background(), link, reader)
			assert.NoError(t, err)
			assert.ElementsMatch(t, expected, lhs)
		})
	}
}

func TestSyncEngine_New(t *testing.T) {
	testCases := []struct {
		name      string
		protocol  protocol.ID
		newEngine func(ihost.Host) store.SyncEngine
	}{{
		"multi-node-sync",
		store.MultiNodeSyncProtocolID,
		func(h ihost.Host) store.SyncEngine {
			return store.NewMultiNodeSyncEngine(h, nil)
		},
	}, {
		"single-node-sync",
		store.SingleNodeSyncProtocolID,
		func(h ihost.Host) store.SyncEngine {
			return store.NewSingleNodeSyncEngine(h, nil)
		},
	}}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			ctx := context.Background()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			h := mockstore.NewMockHost(ctrl)
			h.EXPECT().SetStreamHandler(testCase.protocol, gomock.Any()).Times(1)
			h.EXPECT().RemoveStreamHandler(testCase.protocol).Times(1)

			engine := testCase.newEngine(h)
			engine.Close(ctx)
		})
	}
}

func TestMultiNodeSyncEngine_GetMissingLinks(t *testing.T) {
	t.Run("node-not-connected", func(t *testing.T) {
		ctx := context.Background()

		h := bhost.NewBlankHost(netutil.GenSwarmNetwork(t, ctx))
		defer h.Close()

		indigoStore := dummystore.New(&dummystore.Config{})
		engine := store.NewMultiNodeSyncEngine(h, indigoStore)
		defer engine.Close(ctx)

		_, err := engine.GetMissingLinks(
			ctx,
			"some peer",
			cstesting.NewLinkBuilder().Build(),
			indigoStore,
		)
		assert.EqualError(t, err, store.ErrNoConnectedPeers.Error())
	})

	t.Run("simple-one-pass-sync", func(t *testing.T) {
		ctx := context.Background()

		h1 := bhost.NewBlankHost(netutil.GenSwarmNetwork(t, ctx))
		h2 := bhost.NewBlankHost(netutil.GenSwarmNetwork(t, ctx))
		defer h1.Close()
		defer h2.Close()

		assert.NoError(t, h1.Connect(ctx, h2.Peerstore().PeerInfo(h2.ID())))

		// link ---> prevLink
		//   `-----> refLink1
		//   `-----> refLink2
		prevLink := cstesting.NewLinkBuilder().WithoutParent().Build()
		refLink1 := cstesting.NewLinkBuilder().WithoutParent().Build()
		refLink2 := cstesting.NewLinkBuilder().WithoutParent().Build()
		link := cstesting.NewLinkBuilder().
			WithParent(prevLink).
			WithRef(refLink1).
			WithRef(refLink2).
			Build()

		// Node 1 has refLink2
		store1 := dummystore.New(&dummystore.Config{})
		store1.CreateLink(ctx, refLink2)
		// Node 2 contains all references except one.
		store2 := dummystore.New(&dummystore.Config{})
		store2.CreateLink(ctx, prevLink)
		store2.CreateLink(ctx, refLink1)

		engine1 := store.NewMultiNodeSyncEngine(h1, store1)
		engine2 := store.NewMultiNodeSyncEngine(h2, store2)
		defer engine1.Close(ctx)
		defer engine2.Close(ctx)

		links, err := engine1.GetMissingLinks(ctx, h2.ID(), link, store1)
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

		// link ---> prevLink
		//   `-----> refLink
		prevLink := cstesting.NewLinkBuilder().WithoutParent().Build()
		refLink := cstesting.NewLinkBuilder().WithoutParent().Build()
		link := cstesting.NewLinkBuilder().
			WithParent(prevLink).
			WithRef(refLink).
			Build()

		// Node 1 has nothing locally.
		store1 := dummystore.New(&dummystore.Config{})
		// Node 2 contains all references except one.
		store2 := dummystore.New(&dummystore.Config{})
		store2.CreateLink(ctx, prevLink)

		engine1 := store.NewMultiNodeSyncEngine(h1, store1)
		engine2 := store.NewMultiNodeSyncEngine(h2, store2)
		defer engine1.Close(ctx)
		defer engine2.Close(ctx)

		_, err := engine1.GetMissingLinks(ctx, h2.ID(), link, store1)
		assert.EqualError(t, err, store.ErrLinkNotFound.Error())
	})

	t.Run("recursive-sync", func(t *testing.T) {
		ctx := context.Background()

		h1 := bhost.NewBlankHost(netutil.GenSwarmNetwork(t, ctx))
		h2 := bhost.NewBlankHost(netutil.GenSwarmNetwork(t, ctx))
		h3 := bhost.NewBlankHost(netutil.GenSwarmNetwork(t, ctx))
		defer h1.Close()
		defer h2.Close()
		defer h3.Close()

		assert.NoError(t, h1.Connect(ctx, h2.Peerstore().PeerInfo(h2.ID())))
		assert.NoError(t, h1.Connect(ctx, h3.Peerstore().PeerInfo(h3.ID())))

		// link ---> prevLink
		//   |           `-------> prevPrevLink
		//   `-----> refLink ----> prevRefLink
		prevPrevLink := cstesting.NewLinkBuilder().WithoutParent().Build()
		prevLink := cstesting.NewLinkBuilder().
			WithoutParent().
			WithRef(prevPrevLink).
			Build()
		prevRefLink := cstesting.NewLinkBuilder().WithoutParent().Build()
		refLink := cstesting.NewLinkBuilder().WithParent(prevRefLink).Build()
		link := cstesting.NewLinkBuilder().
			WithParent(prevLink).
			WithRef(refLink).
			Build()

		store1 := dummystore.New(&dummystore.Config{})
		store1.CreateLink(ctx, prevRefLink)

		store2 := dummystore.New(&dummystore.Config{})
		store2.CreateLink(ctx, prevLink)

		store3 := dummystore.New(&dummystore.Config{})
		store3.CreateLink(ctx, refLink)
		store3.CreateLink(ctx, prevPrevLink)

		engine1 := store.NewMultiNodeSyncEngine(h1, store1)
		engine2 := store.NewMultiNodeSyncEngine(h2, store2)
		engine3 := store.NewMultiNodeSyncEngine(h3, store3)
		defer engine1.Close(ctx)
		defer engine2.Close(ctx)
		defer engine3.Close(ctx)

		links, err := engine1.GetMissingLinks(ctx, h3.ID(), link, store1)
		assert.NoError(t, err)
		assert.Len(t, links, 3)
		assert.Contains(t, links, prevPrevLink)
		assert.Contains(t, links, prevLink)
		assert.Contains(t, links, refLink)

		// To comply with dependency ordering, prevPrevLink should appear
		// before prevLink.
		assert.True(t, getLinkIndex(prevPrevLink, links) < getLinkIndex(prevLink, links))
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
