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

package bootstraptest

import (
	"context"
	"testing"

	"github.com/stratumn/alice/core/app/bootstrap/protocol"
	"github.com/stratumn/alice/core/app/bootstrap/protocol/proposal"
	"github.com/stratumn/alice/core/protector"
	protectorpb "github.com/stratumn/alice/core/protector/pb"
	"github.com/stratumn/alice/core/streamutil"
	"github.com/stretchr/testify/require"

	netutil "gx/ipfs/Qmb6BsZf6Y3kxffXMNTubGPF1w1bkHtpvhfYbmnwP3NQyw/go-libp2p-netutil"
	bhost "gx/ipfs/Qmc64U41EEB4nPG7wxjEqFwKJajS2f8kk5q2TvUrQf78Xu/go-libp2p-blankhost"
	"gx/ipfs/QmcJukH2sAFjY3HdBKq35WDzWoL3UUu2gt9wdfqZTUyM74/go-libp2p-peer"
	"gx/ipfs/QmdeiKhUy1TVGBaKxt7y1QmBDLBdisSrLJ1x58Eoj4PXUh/go-libp2p-peerstore"
	"gx/ipfs/Qme1knMqwt1hKZbc1BmQFmnm9f36nyQGwXxPGVpVJ9rMK5/go-libp2p-crypto"
	ihost "gx/ipfs/QmfZTdmunzKzAGJrSvXXQbQ5kLLUiEMX5vdwux7iXkdk7D/go-libp2p-host"
)

// TestNetwork lets you configure a test network.
type TestNetwork struct {
	ctx context.Context
	t   *testing.T

	coordinator      *bhost.BlankHost
	coordinatorCfg   protector.NetworkConfig
	coordinatorStore proposal.Store

	coordinated       map[peer.ID]*bhost.BlankHost
	coordinatedCfgs   map[peer.ID]protector.NetworkConfig
	coordinatedStores map[peer.ID]proposal.Store
}

// NewTestNetwork returns an empty TestNetwork.
func NewTestNetwork(ctx context.Context, t *testing.T) *TestNetwork {
	return &TestNetwork{
		ctx:               ctx,
		t:                 t,
		coordinated:       make(map[peer.ID]*bhost.BlankHost),
		coordinatedCfgs:   make(map[peer.ID]protector.NetworkConfig),
		coordinatedStores: make(map[peer.ID]proposal.Store),
	}
}

// AddCoordinatorNode adds a coordinator node to the network.
func (n *TestNetwork) AddCoordinatorNode() protocol.Handler {
	h := bhost.NewBlankHost(netutil.GenSwarmNetwork(n.t, n.ctx))
	coordinatorID := h.ID()
	coordinatorKey := h.Peerstore().PrivKey(coordinatorID)

	cfg, err := protector.NewInMemoryConfig(
		context.Background(),
		protectorpb.NewNetworkConfig(protectorpb.NetworkState_BOOTSTRAP),
	)
	require.NoError(n.t, err, "protector.NewInMemoryConfig()")

	err = cfg.AddPeer(context.Background(), h.ID(), h.Addrs())
	require.NoError(n.t, err)

	n.coordinator = h
	n.coordinatorStore = proposal.NewInMemoryStore()
	n.coordinatorCfg = cfg

	return protocol.NewCoordinatorHandler(
		n.coordinator,
		streamutil.NewStreamProvider(),
		protector.WrapWithSignature(
			n.coordinatorCfg,
			coordinatorKey,
		),
		n.coordinatorStore,
	)
}

// AddCoordinatedNode adds a coordinated node to the test network.
func (n *TestNetwork) AddCoordinatedNode() (protocol.Handler, peer.ID) {
	require.NotNil(n.t, n.coordinator, "n.coordinator")

	h := bhost.NewBlankHost(netutil.GenSwarmNetwork(n.t, n.ctx))
	h.Peerstore().AddAddrs(
		n.CoordinatorID(),
		n.CoordinatorHost().Addrs(),
		peerstore.PermanentAddrTTL,
	)

	cfg, err := protector.NewInMemoryConfig(
		context.Background(),
		protectorpb.NewNetworkConfig(protectorpb.NetworkState_BOOTSTRAP),
	)
	require.NoError(n.t, err, "protector.NewInMemoryConfig()")

	err = cfg.AddPeer(context.Background(), n.coordinator.ID(), n.coordinator.Addrs())
	require.NoError(n.t, err)

	propStore := proposal.NewInMemoryStore()

	n.coordinated[h.ID()] = h
	n.coordinatedCfgs[h.ID()] = cfg
	n.coordinatedStores[h.ID()] = propStore

	return protocol.NewCoordinatedHandler(
		h,
		streamutil.NewStreamProvider(),
		&protector.NetworkMode{
			CoordinatorID:  n.CoordinatorID(),
			ProtectionMode: protector.PrivateWithCoordinatorMode,
		},
		cfg,
		propStore,
	), h.ID()
}

// CoordinatorHost returns the underlying host of the coordinator.
func (n *TestNetwork) CoordinatorHost() ihost.Host {
	return n.coordinator
}

// CoordinatorID returns the ID of the coordinator.
func (n *TestNetwork) CoordinatorID() peer.ID {
	if n.coordinator != nil {
		return n.coordinator.ID()
	}

	return ""
}

// CoordinatorKey returns the private key of the coordinator.
func (n *TestNetwork) CoordinatorKey() crypto.PrivKey {
	if n.coordinator != nil {
		return n.coordinator.Peerstore().PrivKey(n.CoordinatorID())
	}

	return nil
}

// CoordinatorConfig returns the network config of the coordinator.
func (n *TestNetwork) CoordinatorConfig() protector.NetworkConfig {
	return n.coordinatorCfg
}

// CoordinatorStore returns the proposal store of the coordinator.
func (n *TestNetwork) CoordinatorStore() proposal.Store {
	return n.coordinatorStore
}

// CoordinatedHost returns the underlying host of a given coordinated node.
func (n *TestNetwork) CoordinatedHost(peerID peer.ID) ihost.Host {
	return n.coordinated[peerID]
}

// CoordinatedConfig returns the network config of a given coordinated node.
func (n *TestNetwork) CoordinatedConfig(peerID peer.ID) protector.NetworkConfig {
	return n.coordinatedCfgs[peerID]
}

// CoordinatedStore returns the proposal store of a given coordinated node.
func (n *TestNetwork) CoordinatedStore(peerID peer.ID) proposal.Store {
	return n.coordinatedStores[peerID]
}

// Close tears down the network components.
func (n *TestNetwork) Close() {
	if n.coordinator != nil {
		require.NoError(n.t, n.coordinator.Close(), "coordinator.Close()")
	}

	for _, h := range n.coordinated {
		require.NoError(n.t, h.Close(), "h.Close()")
	}
}
