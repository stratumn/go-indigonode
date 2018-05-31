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

	"github.com/golang/mock/gomock"
	"github.com/stratumn/alice/core/protector"
	cryptopb "github.com/stratumn/alice/pb/crypto"
	protectorpb "github.com/stratumn/alice/pb/protector"
	"github.com/stratumn/alice/test"
	"github.com/stratumn/alice/test/mocks"
	"github.com/stretchr/testify/require"

	"gx/ipfs/QmRDePEiL4Yupq5EkcK3L3ko3iMgYaqUdLu7xc1kqs7dnV/go-multicodec"
	"gx/ipfs/QmWWQ2Txc2c6tqjsBpzg5Ar652cHPGNsQQp2SejkNmkUMb/go-multiaddr"
	"gx/ipfs/QmcJukH2sAFjY3HdBKq35WDzWoL3UUu2gt9wdfqZTUyM74/go-libp2p-peer"
	"gx/ipfs/Qme1knMqwt1hKZbc1BmQFmnm9f36nyQGwXxPGVpVJ9rMK5/go-libp2p-crypto"
)

// NetworkConfigBuilder lets you configure a mock NetworkConfig.
type NetworkConfigBuilder struct {
	t          *testing.T
	mockConfig *mocks.MockNetworkConfig
	configData *protectorpb.NetworkConfig
}

// NewNetworkConfigBuilder initializes a default NetworkConfig.
func NewNetworkConfigBuilder(t *testing.T, ctrl *gomock.Controller) *NetworkConfigBuilder {
	return &NetworkConfigBuilder{
		t:          t,
		mockConfig: mocks.NewMockNetworkConfig(ctrl),
		configData: protectorpb.NewNetworkConfig(protectorpb.NetworkState_BOOTSTRAP),
	}
}

// WithAllowedPeer allows a peer with a random address.
func (c *NetworkConfigBuilder) WithAllowedPeer(peerID peer.ID) *NetworkConfigBuilder {
	_, ok := c.configData.Participants[peerID.Pretty()]
	if !ok {
		c.configData.Participants[peerID.Pretty()] = &protectorpb.PeerAddrs{
			Addresses: []string{test.GeneratePeerMultiaddr(c.t, peerID).String()},
		}
	}

	return c
}

// WithAllowedPeerAddrs allows a peer with given addresses.
func (c *NetworkConfigBuilder) WithAllowedPeerAddrs(peerID peer.ID, peerAddrs ...multiaddr.Multiaddr) *NetworkConfigBuilder {
	addrs, ok := c.configData.Participants[peerID.Pretty()]
	if !ok {
		addrs = &protectorpb.PeerAddrs{}
	}

	for _, addr := range peerAddrs {
		addrs.Addresses = append(addrs.Addresses, addr.String())
	}

	c.configData.Participants[peerID.Pretty()] = addrs
	return c
}

// WithNetworkState sets the network state.
func (c *NetworkConfigBuilder) WithNetworkState(networkState protectorpb.NetworkState) *NetworkConfigBuilder {
	c.configData.NetworkState = networkState
	return c
}

// WithInvalidSignature sets an invalid signature on the config.
func (c *NetworkConfigBuilder) WithInvalidSignature() *NetworkConfigBuilder {
	c.configData.Signature = &cryptopb.Signature{
		KeyType:   cryptopb.KeyType_Ed25519,
		PublicKey: []byte("hello world"),
		Signature: []byte("random junk"),
	}
	return c
}

// WithValidSignature makes sure the configuration is correctly
// signed by the given peer.
func (c *NetworkConfigBuilder) WithValidSignature(signerKey crypto.PrivKey) *NetworkConfigBuilder {
	err := c.configData.Sign(context.Background(), signerKey)
	require.NoError(c.t, err, "configData.Sign()")
	return c
}

// Build builds the NetworkConfig.
func (c *NetworkConfigBuilder) Build() protector.NetworkConfig {
	c.mockConfig.EXPECT().NetworkState(gomock.Any()).AnyTimes().Return(c.configData.NetworkState)

	var allowedPeers []peer.ID
	for peerStr := range c.configData.Participants {
		peerID, _ := peer.IDB58Decode(peerStr)
		c.mockConfig.EXPECT().IsAllowed(gomock.Any(), peerID).AnyTimes().Return(true)
		allowedPeers = append(allowedPeers, peerID)
	}

	c.mockConfig.EXPECT().IsAllowed(gomock.Any(), gomock.Any()).AnyTimes().Return(false)
	c.mockConfig.EXPECT().AllowedPeers(gomock.Any()).AnyTimes().Return(allowedPeers)

	c.mockConfig.EXPECT().Encode(gomock.Any()).AnyTimes().DoAndReturn(func(enc multicodec.Encoder) error {
		return enc.Encode(c.configData)
	})

	return c.mockConfig
}
