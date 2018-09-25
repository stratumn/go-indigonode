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

package protector

import (
	"context"
	"net"
	"sync"

	"github.com/stratumn/go-node/core/protector/pb"

	"gx/ipfs/QmQsErDt8Qgw1XrsXf2BpEzDgGWtB1YLsTAARBup5b6B9W/go-libp2p-peer"
	"gx/ipfs/QmW7Ump7YyBMr712Ta3iEVh3ZYcfVvJaPryfbCnyE826b4/go-libp2p-interface-pnet"
	"gx/ipfs/QmYmsdtJ3HsodkePE3eU3TsCaP2YvPZJ4LoXnNkDE5Tpt7/go-multiaddr"
	"gx/ipfs/Qmda4cPRvSRyox3SqgJN6DfSZGU5TtHufPTp9uXjFj71X6/go-libp2p-peerstore"
)

// PrivateNetworkWithBootstrap implements the github.com/libp2p/go-libp2p-interface-pnet/ipnet.Protector interface.
// It protects a network by only allowing whitelisted peers to connect once the
// bootstrap phase is complete.
// During the bootstrap phase, it accepts all requests.
type PrivateNetworkWithBootstrap struct {
	privateNetwork Protector

	networkStateLock sync.RWMutex
	networkState     pb.NetworkState
}

// NewPrivateNetworkWithBootstrap creates a protector for private networks
// supporting an open bootstrapping phase.
// The protector accepts all connections during the bootstrap phase.
// Once the network state changes and ends the bootstrap phase, the protector
// starts rejecting every non-white-listed request.
func NewPrivateNetworkWithBootstrap(peerStore peerstore.Peerstore) Protector {
	p := PrivateNetworkWithBootstrap{
		privateNetwork: NewPrivateNetwork(peerStore),
		networkState:   pb.NetworkState_BOOTSTRAP,
	}

	// We initially allow all requests.
	ipnet.ForcePrivateNetwork = false

	return &p
}

// Protect accepts all connections until the bootstrap channel is notified.
// Then it switches to private network mode.
func (p *PrivateNetworkWithBootstrap) Protect(conn net.Conn) (net.Conn, error) {
	p.networkStateLock.RLock()
	bootstrapDone := p.networkState != pb.NetworkState_BOOTSTRAP
	p.networkStateLock.RUnlock()

	if !bootstrapDone {
		return conn, nil
	}

	return p.privateNetwork.Protect(conn)
}

// ListenForUpdates listens for network updates.
// This is a blocking call that should be made in a dedicated go routine.
// Closing the channel will stop the listener.
func (p *PrivateNetworkWithBootstrap) ListenForUpdates(updateChan <-chan NetworkUpdate) {
	p.privateNetwork.ListenForUpdates(updateChan)
}

// Fingerprint returns a hash of the participants list.
func (p *PrivateNetworkWithBootstrap) Fingerprint() []byte {
	return p.privateNetwork.Fingerprint()
}

// AllowedAddrs returns the list of whitelisted addresses.
func (p *PrivateNetworkWithBootstrap) AllowedAddrs(ctx context.Context) []multiaddr.Multiaddr {
	return p.privateNetwork.AllowedAddrs(ctx)
}

// AllowedPeers returns the list of whitelisted peers.
func (p *PrivateNetworkWithBootstrap) AllowedPeers(ctx context.Context) []peer.ID {
	return p.privateNetwork.AllowedPeers(ctx)
}

// SetNetworkState sets the network state. The protector adapts to the
// network state, so this method should be called when it changes.
func (p *PrivateNetworkWithBootstrap) SetNetworkState(_ context.Context, networkState pb.NetworkState) error {
	p.networkStateLock.Lock()
	defer p.networkStateLock.Unlock()

	p.networkState = networkState
	switch p.networkState {
	case pb.NetworkState_BOOTSTRAP:
		ipnet.ForcePrivateNetwork = false
	default:
		ipnet.ForcePrivateNetwork = true
	}

	return nil
}
