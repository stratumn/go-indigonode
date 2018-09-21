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
	"github.com/stratumn/go-indigonode/core/protector"
	"github.com/stratumn/go-indigonode/core/protector/pb"
	"github.com/stratumn/go-indigonode/test"
	libp2pmocks "github.com/stratumn/go-indigonode/test/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gx/ipfs/QmQsErDt8Qgw1XrsXf2BpEzDgGWtB1YLsTAARBup5b6B9W/go-libp2p-peer"
	"gx/ipfs/QmW7Ump7YyBMr712Ta3iEVh3ZYcfVvJaPryfbCnyE826b4/go-libp2p-interface-pnet"
	"gx/ipfs/QmYmsdtJ3HsodkePE3eU3TsCaP2YvPZJ4LoXnNkDE5Tpt7/go-multiaddr"
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
			peer1: []multiaddr.Multiaddr{test.GeneratePeerMultiaddr(t, peer1)},
			peer2: []multiaddr.Multiaddr{test.GeneratePeerMultiaddr(t, peer2)},
		},
	}

	p := protector.NewPrivateNetworkWithBootstrap(testData.PeerStore(ctrl))
	go p.ListenForUpdates(updateChan)

	updateChan <- protector.CreateAddNetworkUpdate(peer1)
	updateChan <- protector.CreateAddNetworkUpdate(peer2)

	waitUntilAllowed(t, p, peer2, 2)

	// All connections are accepted during bootstrap.
	bootstrapConn := libp2pmocks.NewMockTransportConn(ctrl)
	wrappedConn, err := p.Protect(bootstrapConn)
	assert.Equal(t, bootstrapConn, wrappedConn)

	// Notifying the bootstrap channel starts rejecting unauthorized requests.
	networkStateWriter, ok := p.(protector.NetworkStateWriter)
	require.True(t, ok, "p.(networkStateWriter)")
	networkStateWriter.SetNetworkState(ctx, pb.NetworkState_PROTECTED)

	invalidConn := libp2pmocks.NewMockTransportConn(ctrl)
	invalidConn.EXPECT().LocalMultiaddr().Return(test.GenerateMultiaddr(t)).Times(1)
	invalidConn.EXPECT().RemoteMultiaddr().Return(test.GenerateMultiaddr(t)).Times(1)
	invalidConn.EXPECT().Close().Times(1)

	_, err = p.Protect(invalidConn)
	assert.EqualError(t, err, protector.ErrConnectionRefused.Error())

	validConn := libp2pmocks.NewMockTransportConn(ctrl)
	validConn.EXPECT().LocalMultiaddr().Return(testData.Peers[peer1][0]).Times(1)
	validConn.EXPECT().RemoteMultiaddr().Return(testData.Peers[peer2][0]).Times(1)

	wrappedConn, err = p.Protect(validConn)
	assert.Equal(t, validConn, wrappedConn)
}
