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

package bootstraptesting

import (
	"context"
	"testing"

	"github.com/stratumn/alice/core/protector"
	"github.com/stratumn/alice/core/protocol/bootstrap"
	"github.com/stratumn/alice/core/protocol/bootstrap/proposal"
	"github.com/stretchr/testify/require"

	netutil "gx/ipfs/Qmb6BsZf6Y3kxffXMNTubGPF1w1bkHtpvhfYbmnwP3NQyw/go-libp2p-netutil"
	bhost "gx/ipfs/Qmc64U41EEB4nPG7wxjEqFwKJajS2f8kk5q2TvUrQf78Xu/go-libp2p-blankhost"
	"gx/ipfs/QmcJukH2sAFjY3HdBKq35WDzWoL3UUu2gt9wdfqZTUyM74/go-libp2p-peer"
	"gx/ipfs/QmdeiKhUy1TVGBaKxt7y1QmBDLBdisSrLJ1x58Eoj4PXUh/go-libp2p-peerstore"
	"gx/ipfs/Qme1knMqwt1hKZbc1BmQFmnm9f36nyQGwXxPGVpVJ9rMK5/go-libp2p-crypto"
	ihost "gx/ipfs/QmfZTdmunzKzAGJrSvXXQbQ5kLLUiEMX5vdwux7iXkdk7D/go-libp2p-host"
)

// ConnectAction connects a node to a network.
type ConnectAction func() (bootstrap.Handler, error)

// TestNetwork lets you configure a test network.
type TestNetwork struct {
	ctx         context.Context
	t           *testing.T
	coordinator *bhost.BlankHost
	coordinated []*bhost.BlankHost
}

// NewTestNetwork returns an empty TestNetwork.
func NewTestNetwork(ctx context.Context, t *testing.T) *TestNetwork {
	return &TestNetwork{
		ctx: ctx,
		t:   t,
	}
}

// PrepareCoordinatedNode prepares a coordinated node, giving you access to its peerID
// before connecting it to the network.
// Use the returned ConnectAction to actually add the node to the network.
func (n *TestNetwork) PrepareCoordinatedNode(coordinatorID peer.ID, networkConfig protector.NetworkConfig) (ihost.Host, ConnectAction) {
	h := bhost.NewBlankHost(netutil.GenSwarmNetwork(n.t, n.ctx))

	connect := func() (bootstrap.Handler, error) {
		if n.coordinator != nil {
			err := h.Connect(n.ctx, n.coordinator.Peerstore().PeerInfo(n.CoordinatorID()))
			require.NoError(n.t, err, "h.Connect()")

			n.coordinator.Peerstore().AddAddrs(h.ID(), h.Addrs(), peerstore.PermanentAddrTTL)
		}

		n.coordinated = append(n.coordinated, h)
		return bootstrap.NewCoordinatedHandler(
			n.ctx,
			h,
			&protector.NetworkMode{
				CoordinatorID:  coordinatorID,
				ProtectionMode: protector.PrivateWithCoordinatorMode,
			},
			networkConfig,
		)
	}

	return h, connect
}

// AddCoordinatedNode adds a coordinated node to the test network.
func (n *TestNetwork) AddCoordinatedNode(coordinatorID peer.ID, networkConfig protector.NetworkConfig) (bootstrap.Handler, error) {
	_, connect := n.PrepareCoordinatedNode(coordinatorID, networkConfig)
	return connect()
}

// AddCoordinatorNode adds a coordinator node to the network.
func (n *TestNetwork) AddCoordinatorNode(networkConfig protector.NetworkConfig) (bootstrap.Handler, error) {
	n.coordinator = bhost.NewBlankHost(netutil.GenSwarmNetwork(n.t, n.ctx))
	return bootstrap.NewCoordinatorHandler(
		n.coordinator,
		protector.WrapWithSignature(
			networkConfig,
			n.coordinator.Peerstore().PrivKey(n.coordinator.ID()),
		),
		proposal.NewInMemoryStore(),
	)
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

// Close tears down the network components.
func (n *TestNetwork) Close() {
	if n.coordinator != nil {
		require.NoError(n.t, n.coordinator.Close(), "coordinator.Close()")
	}

	for _, h := range n.coordinated {
		require.NoError(n.t, h.Close(), "h.Close()")
	}
}
