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

	"github.com/pkg/errors"

	"gx/ipfs/QmPUHzTLPZFYqv8WqcBTuMFYTgeom4uHHEaxzk7bd5GYZB/go-libp2p-transport"
	logging "gx/ipfs/QmSpJByNKFX1sCsHBEp3R73FL4NF6FnQTEGyNAXHm2GS52/go-log"
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
	event := log.EventBegin(context.Background(), "Fingerprint")
	defer event.Done()

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
	event.Append(logging.Metadata{"fingerprint": b58})

	return []byte(b58)
}

// Protect drops any connection attempt from or to a nonwhitelisted peer.
func (p *PrivateNetwork) Protect(conn transport.Conn) (transport.Conn, error) {
	event := log.EventBegin(context.Background(), "Protect")
	defer event.Done()

	remoteAddr := conn.RemoteMultiaddr()
	localAddr := conn.LocalMultiaddr()
	event.Append(logging.Metadata{
		"local":  localAddr.String(),
		"remote": remoteAddr.String(),
	})

	for _, addr := range p.AllowedAddrs() {
		if remoteAddr.Equal(addr) {
			return conn, nil
		}
	}

	event.SetError(ErrConnectionRefused)
	if err := conn.Close(); err != nil {
		event.SetError(errors.WithStack(err))
	}

	return conn, ErrConnectionRefused
}

// AllowedAddrs returns all addresses we allow connections to and from.
func (p *PrivateNetwork) AllowedAddrs() []multiaddr.Multiaddr {
	ctx := context.Background()
	defer log.EventBegin(ctx, "AllowedAddrs").Done()

	p.allowedPeersLock.RLock()
	defer p.allowedPeersLock.RUnlock()

	allAddrs := make([]multiaddr.Multiaddr, 0)
	for peer := range p.allowedPeers {
		addrs := p.peerStore.Addrs(peer)
		if len(addrs) == 0 {
			log.Event(ctx, "PeerAddressMissing", logging.Metadata{
				"peerID": peer.Pretty(),
			})
		} else {
			allAddrs = append(allAddrs, addrs...)
		}
	}

	return allAddrs
}

// AllowedPeers returns the list of whitelisted peers.
func (p *PrivateNetwork) AllowedPeers() []peer.ID {
	event := log.EventBegin(context.Background(), "AllowedPeers")
	defer event.Done()

	p.allowedPeersLock.RLock()
	defer p.allowedPeersLock.RUnlock()

	event.Append(logging.Metadata{"peers_count": len(p.allowedPeers)})

	allowed := make([]peer.ID, 0, len(p.allowedPeers))
	for peer := range p.allowedPeers {
		allowed = append(allowed, peer)
	}

	return allowed
}
