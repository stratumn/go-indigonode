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

	"github.com/golang/mock/gomock"
	"github.com/stratumn/go-node/core/protector"
	"github.com/stratumn/go-node/core/protector/pb"
	"github.com/stratumn/go-node/test"
	libp2pmocks "github.com/stratumn/go-node/test/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/libp2p/go-libp2p-peer"
	manet "github.com/multiformats/go-multiaddr-net"
	"github.com/libp2p/go-libp2p-interface-pnet"
	"github.com/multiformats/go-multiaddr"
)

func TestPrivateNetworkWithBootstrap_New(t *testing.T) {
	ipnet.ForcePrivateNetwork = true
	p := protector.NewPrivateNetworkWithBootstrap(nil)

	assert.False(t, ipnet.ForcePrivateNetwork)

	networkStateWriter, ok := p.(protector.NetworkStateWriter)
	require.True(t, ok, "p.(protector.NetworkStateWriter)")

	networkStateWriter.SetNetworkState(context.Background(), pb.NetworkState_PROTECTED)
	assert.True(t, ipnet.ForcePrivateNetwork)
}

func TestPrivateNetworkWithBootstrap_Protect(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	updateChan := make(chan protector.NetworkUpdate)
	defer close(updateChan)

	peer1 := test.GeneratePeerID(t)
	peer2 := test.GeneratePeerID(t)

	testData := &PeerStoreData{
		Peers: map[peer.ID][]multiaddr.Multiaddr{
			peer1: []multiaddr.Multiaddr{test.GenerateNetAddr(t)},
			peer2: []multiaddr.Multiaddr{test.GenerateNetAddr(t)},
		},
	}

	p := protector.NewPrivateNetworkWithBootstrap(testData.PeerStore(ctrl))
	go p.ListenForUpdates(updateChan)

	updateChan <- protector.CreateAddNetworkUpdate(peer1)
	updateChan <- protector.CreateAddNetworkUpdate(peer2)

	waitUntilAllowed(t, p, peer2, 2)

	// All connections are accepted during bootstrap.
	bootstrapConn := libp2pmocks.NewMockNetConn(ctrl)
	wrappedConn, err := p.Protect(bootstrapConn)
	assert.Equal(t, bootstrapConn, wrappedConn)

	// Notifying the bootstrap channel starts rejecting unauthorized requests.
	networkStateWriter, ok := p.(protector.NetworkStateWriter)
	require.True(t, ok, "p.(networkStateWriter)")
	networkStateWriter.SetNetworkState(ctx, pb.NetworkState_PROTECTED)

	invalidConn := libp2pmocks.NewMockNetConn(ctrl)
	invalidLocalAddr, _ := manet.ToNetAddr(test.GenerateNetAddr(t))
	invalidRemoteAddr, _ := manet.ToNetAddr(test.GenerateNetAddr(t))
	invalidConn.EXPECT().LocalAddr().Return(invalidLocalAddr)
	invalidConn.EXPECT().RemoteAddr().Return(invalidRemoteAddr)
	invalidConn.EXPECT().Close()

	_, err = p.Protect(invalidConn)
	assert.EqualError(t, err, protector.ErrConnectionRefused.Error())

	validConn := libp2pmocks.NewMockNetConn(ctrl)
	validLocalAddr, _ := manet.ToNetAddr(testData.Peers[peer1][0])
	validRemoteAddr, _ := manet.ToNetAddr(testData.Peers[peer2][0])
	validConn.EXPECT().LocalAddr().Return(validLocalAddr)
	validConn.EXPECT().RemoteAddr().Return(validRemoteAddr)

	wrappedConn, err = p.Protect(validConn)
	assert.Equal(t, validConn, wrappedConn)
}
