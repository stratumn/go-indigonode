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

package protector

import (
	"context"
	"sort"
	"sync"

	"github.com/stratumn/go-indigonode/core/monitoring"

	"gx/ipfs/QmPUHzTLPZFYqv8WqcBTuMFYTgeom4uHHEaxzk7bd5GYZB/go-libp2p-transport"
	"gx/ipfs/QmWWQ2Txc2c6tqjsBpzg5Ar652cHPGNsQQp2SejkNmkUMb/go-multiaddr"
	"gx/ipfs/QmZyZDi491cCNTLfAhwcaDii2Kg4pwKRkhqQzURGDvY6ua/go-multihash"
	"gx/ipfs/QmcJukH2sAFjY3HdBKq35WDzWoL3UUu2gt9wdfqZTUyM74/go-libp2p-peer"
	"gx/ipfs/Qmd3oYWVLCVWryDV6Pobv6whZcvDXAHqS3chemZ658y4a8/go-libp2p-interface-pnet"
	"gx/ipfs/QmdeiKhUy1TVGBaKxt7y1QmBDLBdisSrLJ1x58Eoj4PXUh/go-libp2p-peerstore"
)

// PrivateNetwork implements the github.com/libp2p/go-libp2p-interface-pnet/ipnet.Protector interface.
// It protects a network by only allowing whitelisted peers to connect.
type PrivateNetwork struct {
	peerStore peerstore.Peerstore

	allowedPeersLock sync.RWMutex
	allowedPeers     map[peer.ID]struct{}
}

// NewPrivateNetwork creates a protector for private networks.
// It needs the PeerStore used by all network connections.
// It is not the responsibility of this protector to add whitelisted peers
// to the PeerStore, that needs to be done by another component for the
// end-to-end flow to work properly.
func NewPrivateNetwork(peerStore peerstore.Peerstore) Protector {
	ipnet.ForcePrivateNetwork = true

	return &PrivateNetwork{
		peerStore:    peerStore,
		allowedPeers: make(map[peer.ID]struct{}),
	}
}

// ListenForUpdates listens for network updates.
// This is a blocking call that should be made in a dedicated go routine.
// Closing the channel will stop the listener.
func (p *PrivateNetwork) ListenForUpdates(updateChan <-chan NetworkUpdate) {
	for {
		update, ok := <-updateChan
		if !ok {
			return
		}

		p.allowedPeersLock.Lock()

		switch update.Type {
		case Add:
			p.allowedPeers[update.PeerID] = struct{}{}
		case Remove:
			delete(p.allowedPeers, update.PeerID)
		}

		p.allowedPeersLock.Unlock()
	}
}

// Fingerprint returns a hash of the participants list.
func (p *PrivateNetwork) Fingerprint() []byte {
	_, span := monitoring.StartSpan(context.Background(), "protector", "Fingerprint")
	defer span.End()

	p.allowedPeersLock.RLock()
	allowed := make([]string, 0, len(p.allowedPeers))
	for peer := range p.allowedPeers {
		allowed = append(allowed, peer.Pretty())
	}
	p.allowedPeersLock.RUnlock()

	// We need to sort to get a stable fingerprint.
	sort.Strings(allowed)

	var allowedBytes []byte
	for _, peer := range allowed {
		allowedBytes = append(allowedBytes, []byte(peer)...)
	}

	mh, _ := multihash.Sum(allowedBytes, multihash.SHA2_256, -1)
	b58 := mh.B58String()
	span.AddStringAttribute("fingerpint", b58)

	return []byte(b58)
}

// Protect drops any connection attempt from or to a nonwhitelisted peer.
func (p *PrivateNetwork) Protect(conn transport.Conn) (transport.Conn, error) {
	ctx, span := monitoring.StartSpan(context.Background(), "protector", "Protect")
	defer span.End()

	localAddr := conn.LocalMultiaddr()
	span.AddStringAttribute("local", localAddr.String())
	remoteAddr := conn.RemoteMultiaddr()
	span.AddStringAttribute("remote", remoteAddr.String())

	for _, addr := range p.AllowedAddrs(ctx) {
		if remoteAddr.Equal(addr) {
			return conn, nil
		}
	}

	if err := conn.Close(); err != nil {
		span.AddStringAttribute("close_err", err.Error())
	}

	span.SetStatus(monitoring.NewStatus(monitoring.StatusCodePermissionDenied, ErrConnectionRefused.Error()))
	return conn, ErrConnectionRefused
}

// AllowedAddrs returns all addresses we allow connections to and from.
func (p *PrivateNetwork) AllowedAddrs(ctx context.Context) []multiaddr.Multiaddr {
	ctx, span := monitoring.StartSpan(ctx, "protector", "AllowedAddrs")
	defer span.End()

	p.allowedPeersLock.RLock()
	defer p.allowedPeersLock.RUnlock()

	allAddrs := make([]multiaddr.Multiaddr, 0)
	for peer := range p.allowedPeers {
		addrs := p.peerStore.Addrs(peer)
		if len(addrs) == 0 {
			span.Annotate(ctx, peer.Pretty(), "peer address missing")
		} else {
			allAddrs = append(allAddrs, addrs...)
		}
	}

	return allAddrs
}

// AllowedPeers returns the list of whitelisted peers.
func (p *PrivateNetwork) AllowedPeers(ctx context.Context) []peer.ID {
	ctx, span := monitoring.StartSpan(ctx, "protector", "AllowedPeers")
	defer span.End()

	p.allowedPeersLock.RLock()
	defer p.allowedPeersLock.RUnlock()

	span.AddIntAttribute("peers_count", int64(len(p.allowedPeers)))

	allowed := make([]peer.ID, 0, len(p.allowedPeers))
	for peer := range p.allowedPeers {
		allowed = append(allowed, peer)
	}

	return allowed
}
