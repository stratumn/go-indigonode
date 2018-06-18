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

package sync_test

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	pb "github.com/stratumn/alice/app/indigo/pb/store"
	"github.com/stratumn/alice/app/indigo/protocol/store/constants"
	"github.com/stratumn/alice/app/indigo/protocol/store/sync"
	"github.com/stratumn/alice/test/mocks"
	"github.com/stratumn/go-indigocore/cs"
	"github.com/stratumn/go-indigocore/cs/cstesting"
	"github.com/stratumn/go-indigocore/dummystore"
	indigostore "github.com/stratumn/go-indigocore/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	protobuf "gx/ipfs/QmRDePEiL4Yupq5EkcK3L3ko3iMgYaqUdLu7xc1kqs7dnV/go-multicodec/protobuf"
	inet "gx/ipfs/QmXoz9o2PT3tEzf7hicegwex5UgVP54n3k82K7jrWFyN86/go-libp2p-net"
	protocol "gx/ipfs/QmZNkThpqfVXs9GNbexPrfBbXSLNYeKrE7jwFM2oqHbyqN/go-libp2p-protocol"
	netutil "gx/ipfs/Qmb6BsZf6Y3kxffXMNTubGPF1w1bkHtpvhfYbmnwP3NQyw/go-libp2p-netutil"
	bhost "gx/ipfs/Qmc64U41EEB4nPG7wxjEqFwKJajS2f8kk5q2TvUrQf78Xu/go-libp2p-blankhost"
	ihost "gx/ipfs/QmfZTdmunzKzAGJrSvXXQbQ5kLLUiEMX5vdwux7iXkdk7D/go-libp2p-host"
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
			lhs, err := sync.ListMissingLinkHashes(ctx, link, reader)
			assert.NoError(t, err)
			assert.ElementsMatch(t, expected, lhs)
		})
	}
}

func TestOrderLinks(t *testing.T) {
	ctx := context.Background()

	testCases := []struct {
		name  string
		setup func() (*cs.Link, map[string]*cs.Link, indigostore.SegmentReader, []*cs.Link)
		err   error
	}{{
		"tree-dependency",
		func() (*cs.Link, map[string]*cs.Link, indigostore.SegmentReader, []*cs.Link) {
			// start ---> prevLink ---> prevPrevLink
			//   |            `-------> refPrevLink1
			//   |            `-------> refPrevLink2
			//   |
			//   `------> refLink ----> prevRefLink
			//                `-------> refRefLink
			prevPrevLink := cstesting.NewLinkBuilder().WithoutParent().Build()
			refPrevLink1 := cstesting.NewLinkBuilder().WithoutParent().Build()
			refPrevLink2 := cstesting.NewLinkBuilder().WithoutParent().Build()
			prevLink := cstesting.NewLinkBuilder().
				WithParent(prevPrevLink).
				WithRef(refPrevLink1).
				WithRef(refPrevLink2).
				Build()

			prevRefLink := cstesting.NewLinkBuilder().WithoutParent().Build()
			refRefLink := cstesting.NewLinkBuilder().WithoutParent().Build()
			refLink := cstesting.NewLinkBuilder().
				WithParent(prevRefLink).
				WithRef(refRefLink).
				Build()

			start := cstesting.NewLinkBuilder().
				WithParent(prevLink).
				WithRef(refLink).
				Build()

			expected := []*cs.Link{
				prevPrevLink,
				refPrevLink1,
				refPrevLink2,
				prevLink,
				prevRefLink,
				refRefLink,
				refLink,
			}

			return start, toLinksMap(expected), dummystore.New(&dummystore.Config{}), expected
		},
		nil,
	}, {
		"diamond-dependencies",
		func() (*cs.Link, map[string]*cs.Link, indigostore.SegmentReader, []*cs.Link) {
			// start ---> prevLink ---> prevPrevLink <-,
			//   `--------------------------^          |
			//   `------> refLink ----------^          |
			//               `--------> refRefLink ----'
			prevPrevLink := cstesting.NewLinkBuilder().WithoutParent().Build()
			prevLink := cstesting.NewLinkBuilder().WithParent(prevPrevLink).Build()

			refRefLink := cstesting.NewLinkBuilder().WithParent(prevPrevLink).Build()
			refLink := cstesting.NewLinkBuilder().
				WithParent(prevPrevLink).
				WithRef(refRefLink).
				Build()

			start := cstesting.NewLinkBuilder().
				WithParent(prevLink).
				WithRef(prevPrevLink).
				WithRef(refLink).
				Build()

			expected := []*cs.Link{
				prevPrevLink,
				prevLink,
				refRefLink,
				refLink,
			}

			return start, toLinksMap(expected), dummystore.New(&dummystore.Config{}), expected
		},
		nil,
	}, {
		"some-in-store",
		func() (*cs.Link, map[string]*cs.Link, indigostore.SegmentReader, []*cs.Link) {
			// start ---> prevLink ---> prevPrevLink (S) <-,
			//   `--------------------------^              |
			//   `------> refLink ----------^              |
			//               `--------> refRefLink (S) ----'
			prevPrevLink := cstesting.NewLinkBuilder().WithoutParent().Build()
			prevLink := cstesting.NewLinkBuilder().WithParent(prevPrevLink).Build()

			refRefLink := cstesting.NewLinkBuilder().WithParent(prevPrevLink).Build()
			refLink := cstesting.NewLinkBuilder().
				WithParent(prevPrevLink).
				WithRef(refRefLink).
				Build()

			start := cstesting.NewLinkBuilder().
				WithParent(prevLink).
				WithRef(prevPrevLink).
				WithRef(refLink).
				Build()

			testStore := dummystore.New(&dummystore.Config{})
			testStore.CreateLink(ctx, prevPrevLink)
			testStore.CreateLink(ctx, refRefLink)

			expected := []*cs.Link{
				prevLink,
				refLink,
			}

			return start, toLinksMap(expected), testStore, expected
		},
		nil,
	}, {
		"missing-link",
		func() (*cs.Link, map[string]*cs.Link, indigostore.SegmentReader, []*cs.Link) {
			// start ---> prevLink ---> X
			//   `------> refLink ------^
			missingLink := cstesting.NewLinkBuilder().WithoutParent().Build()
			prevLink := cstesting.NewLinkBuilder().WithParent(missingLink).Build()
			refLink := cstesting.NewLinkBuilder().WithParent(missingLink).Build()
			start := cstesting.NewLinkBuilder().
				WithParent(prevLink).
				WithRef(refLink).
				Build()

			linksMap := toLinksMap([]*cs.Link{prevLink, refLink})

			return start, linksMap, dummystore.New(&dummystore.Config{}), nil
		},
		sync.ErrLinkNotFound,
	}}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			link, linksMap, reader, expected := tt.setup()
			links, err := sync.OrderLinks(ctx, link, linksMap, reader)
			if tt.err != nil {
				assert.EqualError(t, err, tt.err.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, expected, links)
			}
		})
	}
}

func TestSyncEngine_New(t *testing.T) {
	testCases := []struct {
		name      string
		protocol  protocol.ID
		newEngine func(ihost.Host) sync.Engine
	}{{
		"multi-node-sync",
		sync.MultiNodeProtocolID,
		func(h ihost.Host) sync.Engine {
			return sync.NewMultiNodeEngine(h, nil)
		},
	}, {
		"single-node-sync",
		sync.SingleNodeProtocolID,
		func(h ihost.Host) sync.Engine {
			return sync.NewSingleNodeEngine(h, nil)
		},
	}}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			ctx := context.Background()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			h := mocks.NewMockHost(ctrl)
			h.EXPECT().SetStreamHandler(testCase.protocol, gomock.Any()).Times(1)
			h.EXPECT().RemoveStreamHandler(testCase.protocol).Times(1)

			engine := testCase.newEngine(h)
			engine.Close(ctx)
		})
	}
}

func TestMultiNodeEngine_GetMissingLinks(t *testing.T) {
	t.Run("node-not-connected", func(t *testing.T) {
		ctx := context.Background()

		h := bhost.NewBlankHost(netutil.GenSwarmNetwork(t, ctx))
		defer h.Close()

		indigoStore := dummystore.New(&dummystore.Config{})
		engine := sync.NewMultiNodeEngine(h, indigoStore)
		defer engine.Close(ctx)

		_, err := engine.GetMissingLinks(
			ctx,
			cstesting.NewLinkBuilder().Build(),
			indigoStore,
		)
		assert.EqualError(t, err, sync.ErrNoConnectedPeers.Error())
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

		engine1 := sync.NewMultiNodeEngine(h1, store1)
		engine2 := sync.NewMultiNodeEngine(h2, store2)
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

		engine1 := sync.NewMultiNodeEngine(h1, store1)
		engine2 := sync.NewMultiNodeEngine(h2, store2)
		defer engine1.Close(ctx)
		defer engine2.Close(ctx)

		_, err := engine1.GetMissingLinks(ctx, link, store1)
		assert.EqualError(t, err, sync.ErrLinkNotFound.Error())
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

		engine1 := sync.NewMultiNodeEngine(h1, store1)
		engine2 := sync.NewMultiNodeEngine(h2, store2)
		engine3 := sync.NewMultiNodeEngine(h3, store3)
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
}

func TestSingleNodeEngine_GetMissingLinks(t *testing.T) {
	t.Run("no-missing-links", func(t *testing.T) {
		ctx := context.Background()

		h := bhost.NewBlankHost(netutil.GenSwarmNetwork(t, ctx))
		defer h.Close()

		prevLink := cstesting.NewLinkBuilder().WithoutParent().Build()
		link := cstesting.NewLinkBuilder().WithParent(prevLink).Build()

		testStore := dummystore.New(&dummystore.Config{})
		testStore.CreateLink(ctx, prevLink)

		engine := sync.NewSingleNodeEngine(h, nil)
		defer engine.Close(ctx)

		links, err := engine.GetMissingLinks(ctx, link, testStore)
		assert.NoError(t, err)
		assert.Nil(t, links)
	})

	t.Run("sender-connection-failure", func(t *testing.T) {
		ctx := context.Background()

		h := bhost.NewBlankHost(netutil.GenSwarmNetwork(t, ctx))
		defer h.Close()

		engine := sync.NewSingleNodeEngine(h, nil)
		defer engine.Close(ctx)

		_, err := engine.GetMissingLinks(
			ctx,
			cstesting.NewLinkBuilder().
				WithMetadata(constants.NodeIDKey, h.ID().Pretty()).
				Build(),
			dummystore.New(&dummystore.Config{}),
		)
		assert.EqualError(t, err, sync.ErrNoConnectedPeers.Error())
	})

	t.Run("sender-protocol-error", func(t *testing.T) {
		ctx := context.Background()

		h1 := bhost.NewBlankHost(netutil.GenSwarmNetwork(t, ctx))
		h2 := bhost.NewBlankHost(netutil.GenSwarmNetwork(t, ctx))
		defer h1.Close()
		defer h2.Close()

		assert.NoError(t, h1.Connect(ctx, h2.Peerstore().PeerInfo(h2.ID())))

		// link ---> prevLink
		//   `--------> linkRef
		linkRef := cstesting.NewLinkBuilder().WithoutParent().Build()
		linkRefHash, _ := linkRef.Hash()
		prevLink := cstesting.NewLinkBuilder().WithoutParent().Build()
		prevLinkHash, _ := prevLink.Hash()
		link := cstesting.NewLinkBuilder().
			WithParent(prevLink).
			WithRef(linkRef).
			WithMetadata(constants.NodeIDKey, h2.ID().Pretty()).
			Build()

		// Set h2 to return only one of the two required segments.
		h2.SetStreamHandler(sync.SingleNodeProtocolID, func(stream inet.Stream) {
			dec := protobuf.Multicodec(nil).Decoder(stream)
			var linkHashes pb.LinkHashes
			require.NoError(t, dec.Decode(&linkHashes))
			require.Len(t, linkHashes.LinkHashes, 2)

			linkHash1, _ := linkHashes.LinkHashes[0].ToLinkHash()
			assert.Equal(t, prevLinkHash, linkHash1)
			linkHash2, _ := linkHashes.LinkHashes[1].ToLinkHash()
			assert.Equal(t, linkRefHash, linkHash2)

			enc := protobuf.Multicodec(nil).Encoder(stream)
			segmentsMessage, _ := pb.FromSegments(cs.SegmentSlice{linkRef.Segmentify()})
			require.NoError(t, enc.Encode(segmentsMessage))
		})

		testStore := dummystore.New(&dummystore.Config{})
		engine := sync.NewSingleNodeEngine(h1, testStore)
		defer engine.Close(ctx)

		_, err := engine.GetMissingLinks(ctx, link, testStore)
		assert.EqualError(t, err, sync.ErrInvalidLinkCount.Error())
	})

	t.Run("missing-previous-link", func(t *testing.T) {
		ctx := context.Background()

		h1 := bhost.NewBlankHost(netutil.GenSwarmNetwork(t, ctx))
		h2 := bhost.NewBlankHost(netutil.GenSwarmNetwork(t, ctx))
		defer h1.Close()
		defer h2.Close()

		assert.NoError(t, h1.Connect(ctx, h2.Peerstore().PeerInfo(h2.ID())))

		// link ---> prevLink
		//              `--------> prevLinkRef
		prevLinkRef := cstesting.NewLinkBuilder().WithoutParent().Build()
		prevLink := cstesting.NewLinkBuilder().WithRef(prevLinkRef).Build()
		link := cstesting.NewLinkBuilder().
			WithParent(prevLink).
			WithMetadata(constants.NodeIDKey, h2.ID().Pretty()).
			Build()

		store1 := dummystore.New(&dummystore.Config{})
		store2 := dummystore.New(&dummystore.Config{})
		store2.CreateLink(ctx, prevLink)
		store2.CreateLink(ctx, link)

		engine1 := sync.NewSingleNodeEngine(h1, store1)
		engine2 := sync.NewSingleNodeEngine(h2, store2)
		defer engine1.Close(ctx)
		defer engine2.Close(ctx)

		_, err := engine1.GetMissingLinks(ctx, link, store1)
		assert.EqualError(t, err, sync.ErrInvalidLinkCount.Error())
	})

	t.Run("diamond-dependency", func(t *testing.T) {
		ctx := context.Background()

		h1 := bhost.NewBlankHost(netutil.GenSwarmNetwork(t, ctx))
		h2 := bhost.NewBlankHost(netutil.GenSwarmNetwork(t, ctx))
		defer h1.Close()
		defer h2.Close()

		assert.NoError(t, h1.Connect(ctx, h2.Peerstore().PeerInfo(h2.ID())))

		// link ---> prevLink ---> prevPrevLink
		//   `------------------------^
		prevPrevLink := cstesting.NewLinkBuilder().WithoutParent().Build()
		prevLink := cstesting.NewLinkBuilder().WithParent(prevPrevLink).Build()
		link := cstesting.NewLinkBuilder().
			WithParent(prevLink).
			WithRef(prevPrevLink).
			WithMetadata(constants.NodeIDKey, h2.ID().Pretty()).
			Build()

		store1 := dummystore.New(&dummystore.Config{})
		store2 := dummystore.New(&dummystore.Config{})
		store2.CreateLink(ctx, prevPrevLink)
		store2.CreateLink(ctx, prevLink)
		store2.CreateLink(ctx, link)

		engine1 := sync.NewSingleNodeEngine(h1, store1)
		engine2 := sync.NewSingleNodeEngine(h2, store2)
		defer engine1.Close(ctx)
		defer engine2.Close(ctx)

		links, err := engine1.GetMissingLinks(ctx, link, store1)
		assert.NoError(t, err)
		require.Len(t, links, 2)
		assert.Equal(t, prevPrevLink, links[0])
		assert.Equal(t, prevLink, links[1])
	})

	t.Run("recursive-sync", func(t *testing.T) {
		ctx := context.Background()

		h1 := bhost.NewBlankHost(netutil.GenSwarmNetwork(t, ctx))
		h2 := bhost.NewBlankHost(netutil.GenSwarmNetwork(t, ctx))
		defer h1.Close()
		defer h2.Close()

		assert.NoError(t, h1.Connect(ctx, h2.Peerstore().PeerInfo(h2.ID())))

		// link ---> prevLink ---> prevPrevLink
		//   |          `--------> prevLinkRef
		//   |
		//   `-----> refLink
		//              `--------> refRefLink
		prevPrevLink := cstesting.NewLinkBuilder().WithoutParent().Build()
		prevLinkRef := cstesting.NewLinkBuilder().WithoutParent().Build()
		prevLink := cstesting.NewLinkBuilder().
			WithParent(prevPrevLink).
			WithRef(prevLinkRef).
			Build()

		refRefLink := cstesting.NewLinkBuilder().WithoutParent().Build()
		refLink := cstesting.NewLinkBuilder().
			WithoutParent().
			WithRef(refRefLink).
			Build()

		link := cstesting.NewLinkBuilder().
			WithParent(prevLink).
			WithRef(refLink).
			WithMetadata(constants.NodeIDKey, h2.ID().Pretty()).
			Build()

		store1 := dummystore.New(&dummystore.Config{})
		store2 := dummystore.New(&dummystore.Config{})
		store2.CreateLink(ctx, prevPrevLink)
		store2.CreateLink(ctx, prevLinkRef)
		store2.CreateLink(ctx, prevLink)
		store2.CreateLink(ctx, refRefLink)
		store2.CreateLink(ctx, refLink)
		store2.CreateLink(ctx, link)

		engine1 := sync.NewSingleNodeEngine(h1, store1)
		engine2 := sync.NewSingleNodeEngine(h2, store2)
		defer engine1.Close(ctx)
		defer engine2.Close(ctx)

		links, err := engine1.GetMissingLinks(ctx, link, store1)
		assert.NoError(t, err)
		assert.Len(t, links, 5)
		assert.Contains(t, links, prevPrevLink)
		assert.Contains(t, links, prevLinkRef)
		assert.Contains(t, links, prevLink)
		assert.Contains(t, links, refRefLink)
		assert.Contains(t, links, refLink)

		assert.True(t, getLinkIndex(prevPrevLink, links) < getLinkIndex(prevLink, links))
		assert.True(t, getLinkIndex(prevLinkRef, links) < getLinkIndex(prevLink, links))
		assert.True(t, getLinkIndex(refRefLink, links) < getLinkIndex(refLink, links))
	})
}

func toLinksMap(links []*cs.Link) map[string]*cs.Link {
	linksMap := make(map[string]*cs.Link)

	for _, link := range links {
		lh, _ := link.HashString()
		linksMap[lh] = link
	}

	return linksMap
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
