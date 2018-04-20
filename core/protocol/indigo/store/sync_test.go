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

		// prevLink <--- link
		//                 |
		// refLink <-------|
		prevLink := cstesting.RandomLink()
		prevLink.Meta.PrevLinkHash = ""
		prevLinkHash, _ := prevLink.Hash()
		refLink := cstesting.RandomLink()
		refLink.Meta.PrevLinkHash = ""
		refLinkHash, _ := refLink.Hash()
		link := cstesting.RandomLink()
		link.Meta.PrevLinkHash = prevLinkHash.String()
		link.Meta.Refs = []cs.SegmentReference{{LinkHash: refLinkHash.String()}}

		// Node 1 doesn't have anything stored yet
		store1 := dummystore.New(&dummystore.Config{})
		// Node 2 contains all the references needed
		store2 := dummystore.New(&dummystore.Config{})
		store2.CreateLink(ctx, prevLink)
		store2.CreateLink(ctx, refLink)

		engine1 := store.NewSynchronousSyncEngine(h1, store1)
		engine2 := store.NewSynchronousSyncEngine(h2, store2)
		defer engine1.Close(ctx)
		defer engine2.Close(ctx)

		links, err := engine1.GetMissingLinks(ctx, link, store1)
		assert.NoError(t, err)
		assert.Equal(t, []*cs.Link{prevLink, refLink}, links)
	})

	// More complex case with 3 hosts and dependencies split between these hosts, recursion needed.
	// Test references including entire segment.
	// Verify results dependency ordering.
	// Case where there is a missing link. Sync should error out.
	// Unexpected errors cases (network, signatures, etc).
}
