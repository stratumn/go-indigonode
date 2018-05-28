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

package protector_test

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stratumn/alice/core/protector"
	"github.com/stratumn/alice/core/protector/mocks"
	"github.com/stratumn/alice/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gx/ipfs/QmWWQ2Txc2c6tqjsBpzg5Ar652cHPGNsQQp2SejkNmkUMb/go-multiaddr"
	"gx/ipfs/QmcJukH2sAFjY3HdBKq35WDzWoL3UUu2gt9wdfqZTUyM74/go-libp2p-peer"
	"gx/ipfs/Qmd3oYWVLCVWryDV6Pobv6whZcvDXAHqS3chemZ658y4a8/go-libp2p-interface-pnet"
)

func TestPrivateNetworkWithBootstrap_New(t *testing.T) {
	ipnet.ForcePrivateNetwork = true
	p := protector.NewPrivateNetworkWithBootstrap(nil)

	assert.False(t, ipnet.ForcePrivateNetwork)

	networkStateWriter, ok := p.(protector.NetworkStateWriter)
	require.True(t, ok, "p.(protector.NetworkStateWriter)")

	networkStateWriter.SetNetworkState(context.Background(), protector.Protected)
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

	// All connections are accepted during bootstrap.
	bootstrapConn := mocks.NewMockConn(ctrl)
	wrappedConn, err := p.Protect(bootstrapConn)
	assert.Equal(t, bootstrapConn, wrappedConn)

	// Notifying the bootstrap channel starts rejecting unauthorized requests.
	networkStateWriter, ok := p.(protector.NetworkStateWriter)
	require.True(t, ok, "p.(networkStateWriter)")
	networkStateWriter.SetNetworkState(ctx, protector.Protected)

	invalidConn := mocks.NewMockConn(ctrl)
	invalidConn.EXPECT().LocalMultiaddr().Return(test.GenerateMultiaddr(t)).Times(1)
	invalidConn.EXPECT().RemoteMultiaddr().Return(test.GenerateMultiaddr(t)).Times(1)
	invalidConn.EXPECT().Close().Times(1)

	_, err = p.Protect(invalidConn)
	assert.EqualError(t, err, protector.ErrConnectionRefused.Error())

	validConn := mocks.NewMockConn(ctrl)
	validConn.EXPECT().LocalMultiaddr().Return(testData.Peers[peer1][0]).Times(1)
	validConn.EXPECT().RemoteMultiaddr().Return(testData.Peers[peer2][0]).Times(1)

	wrappedConn, err = p.Protect(validConn)
	assert.Equal(t, validConn, wrappedConn)
}
