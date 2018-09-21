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

package store_test

import (
	"context"
	"testing"
	"time"

	"github.com/satori/go.uuid"
	"github.com/stratumn/go-indigocore/cs"
	"github.com/stratumn/go-indigocore/cs/cstesting"
	"github.com/stratumn/go-indigonode/app/indigo/protocol/store"
	"github.com/stratumn/go-indigonode/app/indigo/protocol/store/constants"
	"github.com/stratumn/go-indigonode/core/p2p"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	bhost "gx/ipfs/QmQ4bjZSEC5drCRqssuXRymCswHPmW3Z46ibgBtg9XGd34/go-libp2p-blankhost"
	netutil "gx/ipfs/QmfDapjsRAfzVpjeEm2tSmX19QpCrkLDXRCDDWJcbbUsFn/go-libp2p-netutil"
)

func genNetworkID() string {
	return uuid.NewV4().String()
}

func TestPubSub_NodeID(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	h := bhost.NewBlankHost(netutil.GenSwarmNetwork(t, ctx))
	defer h.Close()

	networkID := genNetworkID()
	networkMgr := store.NewPubSubNetworkManager()

	// The NodeID is only available after joining the network.
	assert.NoError(t, networkMgr.Join(ctx, networkID, h))
	assert.Equal(t, h.ID(), networkMgr.NodeID())
}

func TestPubSub_JoinLeave(t *testing.T) {
	t.Run("missing-network-id", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		h := bhost.NewBlankHost(netutil.GenSwarmNetwork(t, ctx))
		defer h.Close()

		networkMgr := store.NewPubSubNetworkManager()

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
		networkMgr := store.NewPubSubNetworkManager()

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

		networkMgr := store.NewPubSubNetworkManager()
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

		networkMgr := store.NewPubSubNetworkManager()
		networkID := genNetworkID()

		assert.NoError(t, networkMgr.Join(ctx, networkID, h))
		assert.NoError(t, networkMgr.Leave(ctx, networkID))
	})
}

func TestPubSub_PublishListen(t *testing.T) {
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
	listeners := make([]<-chan *cs.Segment, nodeCount)
	for i := 0; i < nodeCount; i++ {
		networkMgrs[i] = store.NewPubSubNetworkManager()
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
		constants.SetLinkNodeID(link, hosts[0].ID())
		require.NoError(t, networkMgrs[0].Publish(context.Background(), link))

		for i := 1; i < nodeCount; i++ {
			select {
			case s := <-listeners[i]:
				assert.Equal(t, link, &s.Link)
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

func TestPubSub_AddRemoveListeners(t *testing.T) {
	t.Run("remove-closes-channel", func(t *testing.T) {
		networkMgr := store.NewPubSubNetworkManager()
		testChan := networkMgr.AddListener()
		networkMgr.RemoveListener(testChan)

		_, ok := <-testChan
		assert.False(t, ok, "<-testChan")
	})

	t.Run("remove-unknown-channel", func(t *testing.T) {
		networkMgr := store.NewPubSubNetworkManager()
		privateChan := make(chan *cs.Segment)
		networkMgr.RemoveListener(privateChan)

		networkMgr.AddListener()
		networkMgr.RemoveListener(privateChan)
	})
}
