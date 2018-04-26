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
	"crypto/rand"
	"testing"
	"time"

	json "github.com/gibson042/canonicaljson-go"
	"github.com/satori/go.uuid"
	"github.com/stratumn/alice/core/p2p"
	"github.com/stratumn/alice/core/protocol/indigo/store"
	pb "github.com/stratumn/alice/pb/indigo/store"
	"github.com/stratumn/go-indigocore/cs/cstesting"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	bhost "gx/ipfs/QmQr1j6UvdhpponAaqSdswqRpdzsFwNop2N8kXLNw8afem/go-libp2p-blankhost"
	netutil "gx/ipfs/QmYVR3C8DWPHdHxvLtNFYfjsXgaRAdh6hPMNH3KiwCgu4o/go-libp2p-netutil"
	ic "gx/ipfs/QmaPbCnUMBohSGo3KnxEa2bHqyJVVeEEcwtqJAYxerieBo/go-libp2p-crypto"
)

func genPeerPrivateKey() ic.PrivKey {
	sk, _, _ := ic.GenerateEd25519Key(rand.Reader)
	return sk
}

func genNetworkID() string {
	return uuid.NewV4().String()
}

func TestNodeID(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	h := bhost.NewBlankHost(netutil.GenSwarmNetwork(t, ctx))
	defer h.Close()

	networkID := genNetworkID()
	networkMgr := store.NewNetworkManager(genPeerPrivateKey())

	// The NodeID is only available after joining the network.
	assert.NoError(t, networkMgr.Join(ctx, networkID, h))
	assert.Equal(t, h.ID().Pretty(), networkMgr.NodeID())
}

func TestJoinLeave(t *testing.T) {
	t.Run("missing-network-id", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		h := bhost.NewBlankHost(netutil.GenSwarmNetwork(t, ctx))
		defer h.Close()

		networkMgr := store.NewNetworkManager(genPeerPrivateKey())

		assert.EqualError(t, networkMgr.Join(ctx, "", h), store.ErrInvalidNetworkID.Error())
	})

	t.Run("join-multiple-times", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		h1 := bhost.NewBlankHost(netutil.GenSwarmNetwork(t, ctx))
		h2 := bhost.NewBlankHost(netutil.GenSwarmNetwork(t, ctx))
		defer h1.Close()
		defer h2.Close()

		networkID := genNetworkID()
		networkMgr := store.NewNetworkManager(genPeerPrivateKey())

		assert.NoError(t, networkMgr.Join(ctx, networkID, h1))
		// Subsequent joins should be no-ops.
		assert.NoError(t, networkMgr.Join(ctx, networkID, h1))
		assert.NoError(t, networkMgr.Join(ctx, networkID, h2))
		assert.NoError(t, networkMgr.Join(ctx, networkID, h1))
	})

	t.Run("leave-without-joining", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		h := bhost.NewBlankHost(netutil.GenSwarmNetwork(t, ctx))
		defer h.Close()

		networkMgr := store.NewNetworkManager(genPeerPrivateKey())
		networkID1 := genNetworkID()
		networkID2 := genNetworkID()

		assert.NoError(t, networkMgr.Join(ctx, networkID1, h))
		assert.EqualError(t, networkMgr.Leave(ctx, networkID2), store.ErrInvalidNetworkID.Error())
	})

	t.Run("join-then-leave", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		h := bhost.NewBlankHost(netutil.GenSwarmNetwork(t, ctx))
		defer h.Close()

		networkMgr := store.NewNetworkManager(genPeerPrivateKey())
		networkID := genNetworkID()

		assert.NoError(t, networkMgr.Join(ctx, networkID, h))
		assert.NoError(t, networkMgr.Leave(ctx, networkID))
	})
}

func TestPublishListen(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	nodeCount := 3
	hosts := make([]*p2p.Host, nodeCount)
	for i := 0; i < nodeCount; i++ {
		hosts[i] = p2p.NewHost(ctx, netutil.GenSwarmNetwork(t, ctx))
		defer hosts[i].Close()
	}

	listenCtx, cancelListen := context.WithCancel(ctx)
	networkID := genNetworkID()

	networkMgrs := make([]store.NetworkManager, nodeCount)
	listenChans := make([]chan error, nodeCount)
	listeners := make([]<-chan *pb.SignedLink, nodeCount)
	for i := 0; i < nodeCount; i++ {
		networkMgrs[i] = store.NewNetworkManager(genPeerPrivateKey())
		listeners[i] = networkMgrs[i].AddListener()

		listenChans[i] = make(chan error)
		assert.NoError(t, networkMgrs[i].Join(ctx, networkID, hosts[i]), "networkMgr.Join()")

		go func(i int) { listenChans[i] <- networkMgrs[i].Listen(listenCtx) }(i)
	}

	// Note: on test networks, we need all nodes to first join the floodsub
	// before connecting to each other.
	// Another tricky thing is that floodsub connections don't seem to be
	// bijective in tests: if A connects to B, A will have B as his pubsub peers
	// but B might not have A in its pubsub peers.
	// So we connect all nodes in both directions for the test to prevent
	// random failures.

	for i := 0; i < nodeCount; i++ {
		for j := 0; j < nodeCount; j++ {
			if i == j {
				continue
			}

			err := hosts[i].Connect(ctx, hosts[j].Peerstore().PeerInfo(hosts[j].ID()))
			require.NoErrorf(t, err, "h%d.Connect(h%d)", i, j)
		}
	}

	t.Run("publish-valid-message", func(t *testing.T) {
		link := cstesting.NewLinkBuilder().Build()
		linkBytes, _ := json.Marshal(link)
		assert.NoError(t, networkMgrs[0].Publish(context.Background(), link))

		for i := 1; i < nodeCount; i++ {
			select {
			case l := <-listeners[i]:
				assert.Equal(t, linkBytes, l.Link)
			case <-time.After(100 * time.Millisecond):
				assert.Failf(t, "<-listeners[i]", "listener %d didn't receive link", i)
			}
		}
	})

	cancelListen()
	for i := 0; i < nodeCount; i++ {
		assert.EqualError(t, <-listenChans[i], context.Canceled.Error())
	}

	for i := 0; i < nodeCount; i++ {
		assert.NoError(t, networkMgrs[i].Leave(ctx, networkID))
	}
}

func TestAddRemoveListeners(t *testing.T) {
	t.Run("remove-closes-channel", func(t *testing.T) {
		networkMgr := store.NewNetworkManager(genPeerPrivateKey())
		testChan := networkMgr.AddListener()
		networkMgr.RemoveListener(testChan)

		_, ok := <-testChan
		assert.False(t, ok, "<-testChan")
	})

	t.Run("remove-unknown-channel", func(t *testing.T) {
		networkMgr := store.NewNetworkManager(genPeerPrivateKey())
		privateChan := make(chan *pb.SignedLink)
		networkMgr.RemoveListener(privateChan)

		networkMgr.AddListener()
		networkMgr.RemoveListener(privateChan)
	})
}
