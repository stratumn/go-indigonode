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

package protector_test

import (
	"context"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/pkg/errors"
	"github.com/stratumn/go-node/core/protector"
	"github.com/stratumn/go-node/test"
	libp2pmocks "github.com/stratumn/go-node/test/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/libp2p/go-libp2p-peer"
	manet "github.com/multiformats/go-multiaddr-net"
	"github.com/libp2p/go-libp2p-interface-pnet"
	"github.com/multiformats/go-multiaddr"
)

// waitUntilAllowed waits until the given peer is found in the allowed list
// and the allowed list contains allowedCount elements.
// It fails after a short polling period.
func waitUntilAllowed(t *testing.T, p protector.Protector, peer peer.ID, allowedCount int) {
	test.WaitUntil(t, 50*time.Millisecond, 5*time.Millisecond, func() error {
		allowedPeers := p.AllowedPeers(context.Background())
		if allowedCount == 0 && len(allowedPeers) == 0 {
			return nil
		}

		if len(allowedPeers) != allowedCount {
			return errors.New("invalid number of peers")
		}

		for _, allowedPeer := range allowedPeers {
			if allowedPeer == peer {
				return nil
			}
		}

		return errors.New("peer not found in allowed list")
	}, "peer not accepted in time")
}

type PeerStoreData struct {
	Peers map[peer.ID][]multiaddr.Multiaddr
}

func (p *PeerStoreData) PeerStore(ctrl *gomock.Controller) *libp2pmocks.MockPeerstore {
	peerStore := libp2pmocks.NewMockPeerstore(ctrl)

	for peerID, addrs := range p.Peers {
		peerStore.EXPECT().Addrs(peerID).Return(addrs).AnyTimes()
	}

	// Return empty address list for unknown peers.
	peerStore.EXPECT().Addrs(gomock.Any()).Return(nil).AnyTimes()

	return peerStore
}

func TestPrivateNetwork_New(t *testing.T) {
	ipnet.ForcePrivateNetwork = false
	protector.NewPrivateNetwork(nil)

	assert.True(t, ipnet.ForcePrivateNetwork)
}

func TestPrivateNetwork_ListenForUpdates(t *testing.T) {
	updateChan := make(chan protector.NetworkUpdate)
	p := protector.NewPrivateNetwork(nil)

	exitChan := make(chan struct{})
	go func() {
		p.ListenForUpdates(updateChan)
		exitChan <- struct{}{}
	}()

	close(updateChan)

	select {
	case <-time.After(5 * time.Millisecond):
		assert.Fail(t, "ListenForUpdates")
	case <-exitChan:
	}
}

func TestPrivateNetwork_Fingerprint(t *testing.T) {
	peer := test.GeneratePeerID(t)

	t.Run("stable-network", func(t *testing.T) {
		chan1 := make(chan protector.NetworkUpdate)
		defer close(chan1)

		chan2 := make(chan protector.NetworkUpdate)
		defer close(chan2)

		p1 := protector.NewPrivateNetwork(nil)
		go p1.ListenForUpdates(chan1)
		p2 := protector.NewPrivateNetwork(nil)
		go p2.ListenForUpdates(chan2)

		chan1 <- protector.CreateAddNetworkUpdate(peer)
		chan2 <- protector.CreateAddNetworkUpdate(peer)
		waitUntilAllowed(t, p1, peer, 1)
		waitUntilAllowed(t, p2, peer, 1)

		f1 := p1.Fingerprint()
		f2 := p2.Fingerprint()

		require.NotNil(t, f1)
		assert.Equal(t, f1, f2)
	})

	t.Run("changes-on-peer-added", func(t *testing.T) {
		updateChan := make(chan protector.NetworkUpdate)
		defer close(updateChan)

		p := protector.NewPrivateNetwork(nil)
		go p.ListenForUpdates(updateChan)
		f1 := p.Fingerprint()

		updateChan <- protector.CreateAddNetworkUpdate(peer)
		waitUntilAllowed(t, p, peer, 1)
		f2 := p.Fingerprint()

		assert.NotEqual(t, f1, f2)
	})

	t.Run("changes-on-peer-removed", func(t *testing.T) {
		updateChan := make(chan protector.NetworkUpdate)
		defer close(updateChan)

		p := protector.NewPrivateNetwork(nil)
		go p.ListenForUpdates(updateChan)
		f1 := p.Fingerprint()

		updateChan <- protector.CreateAddNetworkUpdate(peer)
		waitUntilAllowed(t, p, peer, 1)
		f2 := p.Fingerprint()

		updateChan <- protector.CreateRemoveNetworkUpdate(peer)
		waitUntilAllowed(t, p, "", 0)
		f3 := p.Fingerprint()

		assert.NotEqual(t, f1, f2)
		assert.Equal(t, f1, f3)
	})
}

func TestPrivateNetwork_Protect(t *testing.T) {
	peer1 := test.GeneratePeerID(t)
	peer2 := test.GeneratePeerID(t)
	peer3 := test.GeneratePeerID(t)

	testData := &PeerStoreData{
		Peers: map[peer.ID][]multiaddr.Multiaddr{
			peer1: []multiaddr.Multiaddr{
				test.GenerateNetAddr(t),
				test.GenerateNetAddr(t),
			},
			peer2: []multiaddr.Multiaddr{
				test.GenerateNetAddr(t),
				test.GenerateNetAddr(t),
			},
		},
	}

	testCases := []struct {
		name           string
		networkUpdates func(protector.Protector, chan<- protector.NetworkUpdate)
		local          multiaddr.Multiaddr
		remote         multiaddr.Multiaddr
		reject         bool
	}{{
		"allowed-peer-not-in-peer-store",
		func(p protector.Protector, updateChan chan<- protector.NetworkUpdate) {
			updateChan <- protector.CreateAddNetworkUpdate(peer1)
			updateChan <- protector.CreateAddNetworkUpdate(peer3)
			waitUntilAllowed(t, p, peer3, 2)
		},
		testData.Peers[peer1][0], // accepted (peer1)
		testData.Peers[peer2][0], // rejected (peer2, not peer3)
		true,
	}, {
		"no-peer-allowed",
		func(protector.Protector, chan<- protector.NetworkUpdate) {},
		testData.Peers[peer1][0],
		testData.Peers[peer2][1],
		true,
	}, {
		"invalid-remote-addr",
		func(p protector.Protector, updateChan chan<- protector.NetworkUpdate) {
			updateChan <- protector.CreateAddNetworkUpdate(peer1)
			waitUntilAllowed(t, p, peer1, 1)
		},
		testData.Peers[peer1][0],
		testData.Peers[peer2][0],
		true,
	}, {
		"removed-peer",
		func(p protector.Protector, updateChan chan<- protector.NetworkUpdate) {
			updateChan <- protector.CreateAddNetworkUpdate(peer2)
			waitUntilAllowed(t, p, peer2, 1)
			updateChan <- protector.CreateAddNetworkUpdate(peer1)
			updateChan <- protector.CreateRemoveNetworkUpdate(peer2)
			waitUntilAllowed(t, p, peer1, 1)
		},
		testData.Peers[peer1][0],
		testData.Peers[peer2][0],
		true,
	}, {
		"invalid-local-addr-ignored",
		func(p protector.Protector, updateChan chan<- protector.NetworkUpdate) {
			updateChan <- protector.CreateAddNetworkUpdate(peer2)
			waitUntilAllowed(t, p, peer2, 1)
		},
		// Peer1 isn't whitelisted but we only validate the remote address (peer2).
		testData.Peers[peer1][0],
		testData.Peers[peer2][0],
		false,
	}, {
		"valid-to-valid-from",
		func(p protector.Protector, updateChan chan<- protector.NetworkUpdate) {
			updateChan <- protector.CreateAddNetworkUpdate(peer1)
			updateChan <- protector.CreateAddNetworkUpdate(peer2)
			waitUntilAllowed(t, p, peer2, 2)
		},
		testData.Peers[peer1][0],
		testData.Peers[peer2][1],
		false,
	}}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			updateChan := make(chan protector.NetworkUpdate)
			defer close(updateChan)

			p := protector.NewPrivateNetwork(testData.PeerStore(ctrl))
			go p.ListenForUpdates(updateChan)

			tt.networkUpdates(p, updateChan)

			conn := libp2pmocks.NewMockNetConn(ctrl)
			local, _ := manet.ToNetAddr(tt.local)
			remote, _ := manet.ToNetAddr(tt.remote)
			conn.EXPECT().LocalAddr().Return(local)
			conn.EXPECT().RemoteAddr().Return(remote)
			if tt.reject {
				conn.EXPECT().Close()
			}

			wrappedConn, err := p.Protect(conn)
			if tt.reject {
				assert.EqualError(t, err, protector.ErrConnectionRefused.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, conn, wrappedConn)
			}
		})
	}
}

func TestPrivateNetwork_AllowedAddrs(t *testing.T) {
	peer1 := test.GeneratePeerID(t)
	peer2 := test.GeneratePeerID(t)
	peer3 := test.GeneratePeerID(t)

	testData := &PeerStoreData{
		Peers: map[peer.ID][]multiaddr.Multiaddr{
			peer1: []multiaddr.Multiaddr{
				test.GeneratePeerMultiaddr(t, peer1),
				test.GeneratePeerMultiaddr(t, peer1),
			},
			peer2: []multiaddr.Multiaddr{
				test.GeneratePeerMultiaddr(t, peer2),
				test.GeneratePeerMultiaddr(t, peer2),
			},
		},
	}

	testCases := []struct {
		name           string
		networkUpdates func(protector.Protector, chan<- protector.NetworkUpdate)
		addrs          []multiaddr.Multiaddr
	}{{
		"ignores-missing-peer",
		func(p protector.Protector, updateChan chan<- protector.NetworkUpdate) {
			updateChan <- protector.CreateAddNetworkUpdate(peer1)
			updateChan <- protector.CreateAddNetworkUpdate(peer3)
			waitUntilAllowed(t, p, peer3, 2)
		},
		testData.Peers[peer1],
	}, {
		"ignores-removed-peer",
		func(p protector.Protector, updateChan chan<- protector.NetworkUpdate) {
			updateChan <- protector.CreateAddNetworkUpdate(peer1)
			updateChan <- protector.CreateAddNetworkUpdate(peer2)
			updateChan <- protector.CreateRemoveNetworkUpdate(peer1)
			waitUntilAllowed(t, p, peer2, 1)
		},
		testData.Peers[peer2],
	}, {
		"returns-all-addresses",
		func(p protector.Protector, updateChan chan<- protector.NetworkUpdate) {
			updateChan <- protector.CreateAddNetworkUpdate(peer1)
			updateChan <- protector.CreateAddNetworkUpdate(peer2)
			updateChan <- protector.CreateAddNetworkUpdate(peer3)
			waitUntilAllowed(t, p, peer3, 3)
		},
		append(testData.Peers[peer1], testData.Peers[peer2]...),
	}}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			updateChan := make(chan protector.NetworkUpdate)
			defer close(updateChan)

			p := protector.NewPrivateNetwork(testData.PeerStore(ctrl))
			go p.ListenForUpdates(updateChan)

			tt.networkUpdates(p, updateChan)

			assert.ElementsMatch(t, tt.addrs, p.AllowedAddrs(context.Background()))
		})
	}
}

func TestPrivateNetwork_AllowedPeers(t *testing.T) {
	peer1 := test.GeneratePeerID(t)
	peer2 := test.GeneratePeerID(t)
	peer3 := test.GeneratePeerID(t)

	testCases := []struct {
		name           string
		networkUpdates func(chan<- protector.NetworkUpdate)
		peers          []peer.ID
	}{{
		"empty-network",
		func(chan<- protector.NetworkUpdate) {},
		nil,
	}, {
		"add-remove-peer",
		func(updateChan chan<- protector.NetworkUpdate) {
			updateChan <- protector.CreateAddNetworkUpdate(peer2)
			updateChan <- protector.CreateAddNetworkUpdate(peer1)
			updateChan <- protector.CreateRemoveNetworkUpdate(peer2)
			updateChan <- protector.CreateAddNetworkUpdate(peer3)
		},
		[]peer.ID{peer1, peer3},
	}}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			updateChan := make(chan protector.NetworkUpdate)
			defer close(updateChan)

			p := protector.NewPrivateNetwork(nil)
			go p.ListenForUpdates(updateChan)

			tt.networkUpdates(updateChan)

			for _, peer := range tt.peers {
				waitUntilAllowed(t, p, peer, len(tt.peers))
			}
		})
	}
}
