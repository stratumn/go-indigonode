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

package bootstraptest

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stratumn/go-indigonode/core/app/bootstrap/protocol/proposal"
	"github.com/stratumn/go-indigonode/core/protector"
	protectorpb "github.com/stratumn/go-indigonode/core/protector/pb"
	"github.com/stratumn/go-indigonode/test"
	"github.com/stretchr/testify/require"

	inet "gx/ipfs/QmXoz9o2PT3tEzf7hicegwex5UgVP54n3k82K7jrWFyN86/go-libp2p-net"
	"gx/ipfs/QmcJukH2sAFjY3HdBKq35WDzWoL3UUu2gt9wdfqZTUyM74/go-libp2p-peer"
)

func waitUntilNetworkState(t *testing.T, state protectorpb.NetworkState, networkConfig protector.NetworkConfig) {
	test.WaitUntil(t, 100*time.Millisecond, 10*time.Millisecond,
		func() error {
			if networkConfig.NetworkState(context.Background()) != state {
				return errors.New("unexpected network state")
			}

			return nil
		}, "network state not set in time")
}

func waitUntilAllowed(t *testing.T, peerID peer.ID, networkConfig protector.NetworkConfig) {
	test.WaitUntil(t, 100*time.Millisecond, 10*time.Millisecond,
		func() error {
			if !networkConfig.IsAllowed(context.Background(), peerID) {
				return errors.New("peer not allowed")
			}

			return nil
		}, "peer not allowed in time")
}

func waitUntilNotAllowed(t *testing.T, peerID peer.ID, networkConfig protector.NetworkConfig) {
	test.WaitUntil(t, 100*time.Millisecond, 10*time.Millisecond,
		func() error {
			if networkConfig.IsAllowed(context.Background(), peerID) {
				return errors.New("still allowed")
			}

			return nil
		}, "peer not removed in time")
}

func waitUntilProposed(t *testing.T, s proposal.Store, peerID peer.ID) {
	test.WaitUntil(t, 100*time.Millisecond, 10*time.Millisecond,
		func() error {
			r, _ := s.Get(context.Background(), peerID)
			if r == nil {
				return errors.New("proposal not received yet")
			}

			return nil
		}, "proposal not received in time")
}

func waitUntilNotProposed(t *testing.T, s proposal.Store, peerID peer.ID) {
	test.WaitUntil(t, 100*time.Millisecond, 10*time.Millisecond,
		func() error {
			r, _ := s.Get(context.Background(), peerID)
			if r == nil {
				return nil
			}

			return errors.New("proposal not dismissed yet")
		}, "proposal not dismissed in time")
}

func TestPrivateNetworkWithCoordinator(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	testnet := NewTestNetwork(ctx, t)
	coordinator := testnet.AddCoordinatorNode()

	c1, peer1 := testnet.AddCoordinatedNode()
	require.NoError(t, c1.Handshake(ctx))

	c2, peer2 := testnet.AddCoordinatedNode()
	require.NoError(t, c2.Handshake(ctx))

	c3, peer3 := testnet.AddCoordinatedNode()
	require.NoError(t, c3.Handshake(ctx))

	// Coordinated nodes should propose adding themselves when joining the network.

	waitUntilProposed(t, testnet.CoordinatorStore(), peer1)
	waitUntilProposed(t, testnet.CoordinatorStore(), peer2)
	waitUntilProposed(t, testnet.CoordinatorStore(), peer3)

	require.Equal(t, inet.Connected, testnet.CoordinatorHost().Network().Connectedness(peer1))
	require.Equal(t, inet.Connected, testnet.CoordinatorHost().Network().Connectedness(peer2))
	require.Equal(t, inet.Connected, testnet.CoordinatorHost().Network().Connectedness(peer3))

	require.NoError(t, coordinator.Accept(ctx, peer1))
	require.NoError(t, coordinator.Accept(ctx, peer2))
	require.NoError(t, coordinator.Reject(ctx, peer3))

	waitUntilAllowed(t, peer1, testnet.CoordinatorConfig())
	waitUntilAllowed(t, peer2, testnet.CoordinatorConfig())
	waitUntilNotProposed(t, testnet.CoordinatorStore(), peer3)

	// Authorized peers should receive the network configuration,
	// and unauthorized peers should be disconnected from the network.

	require.NoError(t, coordinator.CompleteBootstrap(ctx))
	waitUntilNetworkState(t, protectorpb.NetworkState_PROTECTED, testnet.CoordinatorConfig())
	waitUntilNetworkState(t, protectorpb.NetworkState_PROTECTED, testnet.CoordinatedConfig(peer1))
	waitUntilNetworkState(t, protectorpb.NetworkState_PROTECTED, testnet.CoordinatedConfig(peer2))

	require.ElementsMatch(t, []peer.ID{testnet.CoordinatorID(), peer1, peer2}, testnet.CoordinatorConfig().AllowedPeers(ctx))
	require.ElementsMatch(t, []peer.ID{testnet.CoordinatorID(), peer1, peer2}, testnet.CoordinatedConfig(peer1).AllowedPeers(ctx))
	require.ElementsMatch(t, []peer.ID{testnet.CoordinatorID(), peer1, peer2}, testnet.CoordinatedConfig(peer2).AllowedPeers(ctx))

	require.Equal(t, inet.NotConnected, testnet.CoordinatorHost().Network().Connectedness(peer3))

	// Add Peer3 to the network.

	err := c2.AddNode(ctx, peer3, testnet.CoordinatedHost(peer3).Addrs()[0], []byte("much pr00f"))
	require.NoError(t, err)

	waitUntilProposed(t, testnet.CoordinatorStore(), peer3)

	require.NoError(t, coordinator.Accept(ctx, peer3))

	waitUntilAllowed(t, peer3, testnet.CoordinatorConfig())
	waitUntilAllowed(t, peer3, testnet.CoordinatedConfig(peer1))
	waitUntilAllowed(t, peer3, testnet.CoordinatedConfig(peer2))
	waitUntilAllowed(t, peer3, testnet.CoordinatedConfig(peer3))

	test.WaitUntilConnected(t, testnet.CoordinatorHost(), peer3)

	// Remove Peer1 from the network.

	require.NoError(t, c3.RemoveNode(ctx, peer1))

	waitUntilProposed(t, testnet.CoordinatorStore(), peer1)
	waitUntilProposed(t, testnet.CoordinatedStore(peer1), peer1)
	waitUntilProposed(t, testnet.CoordinatedStore(peer2), peer1)
	waitUntilProposed(t, testnet.CoordinatedStore(peer3), peer1)

	require.NoError(t, c2.Accept(ctx, peer1))
	require.NoError(t, c3.Accept(ctx, peer1))

	waitUntilNotAllowed(t, peer1, testnet.CoordinatorConfig())
	waitUntilNotAllowed(t, peer1, testnet.CoordinatedConfig(peer2))
	waitUntilNotAllowed(t, peer1, testnet.CoordinatedConfig(peer3))

	waitUntilNotProposed(t, testnet.CoordinatorStore(), peer1)
	waitUntilNotProposed(t, testnet.CoordinatedStore(peer2), peer1)
	waitUntilNotProposed(t, testnet.CoordinatedStore(peer3), peer1)

	test.WaitUntilDisconnected(t, testnet.CoordinatorHost(), peer1)
	test.WaitUntilDisconnected(t, testnet.CoordinatedHost(peer2), peer1)
	test.WaitUntilDisconnected(t, testnet.CoordinatedHost(peer3), peer1)

	require.ElementsMatch(t, []peer.ID{testnet.CoordinatorID(), peer2, peer3}, testnet.CoordinatorConfig().AllowedPeers(ctx))
	require.ElementsMatch(t, []peer.ID{testnet.CoordinatorID(), peer2, peer3}, testnet.CoordinatedConfig(peer2).AllowedPeers(ctx))
	require.ElementsMatch(t, []peer.ID{testnet.CoordinatorID(), peer2, peer3}, testnet.CoordinatedConfig(peer3).AllowedPeers(ctx))
}
