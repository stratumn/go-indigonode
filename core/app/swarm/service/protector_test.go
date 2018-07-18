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

package service_test

import (
	"context"
	"io/ioutil"
	"path"
	"testing"
	"time"

	"github.com/pkg/errors"
	swarm "github.com/stratumn/go-indigonode/core/app/swarm/service"
	"github.com/stratumn/go-indigonode/core/protector"
	"github.com/stratumn/go-indigonode/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gx/ipfs/QmWWQ2Txc2c6tqjsBpzg5Ar652cHPGNsQQp2SejkNmkUMb/go-multiaddr"
	"gx/ipfs/QmcJukH2sAFjY3HdBKq35WDzWoL3UUu2gt9wdfqZTUyM74/go-libp2p-peer"
	"gx/ipfs/QmdeiKhUy1TVGBaKxt7y1QmBDLBdisSrLJ1x58Eoj4PXUh/go-libp2p-peerstore"
)

func TestInvalidConfig(t *testing.T) {
	_, err := swarm.NewProtectorConfig(&swarm.Config{ProtectionMode: "over-9000"})
	assert.EqualError(t, err, swarm.ErrInvalidProtectionMode.Error())
}

func TestNoProtectorConfig(t *testing.T) {
	cfg, err := swarm.NewProtectorConfig(&swarm.Config{})
	require.NoError(t, err)

	p, c, err := cfg.Configure(context.Background(), nil, nil)
	require.NoError(t, err)
	assert.Nil(t, p)
	assert.Nil(t, c)
}

func waitUntilAllowed(t *testing.T, peerID peer.ID, networkConfig protector.NetworkConfig) {
	test.WaitUntil(
		t,
		10*time.Millisecond,
		3*time.Millisecond,
		func() error {
			allowed := networkConfig.AllowedPeers(context.Background())
			if len(allowed) == 0 {
				return errors.New("no peer allowed")
			}

			if allowed[0] != peerID {
				return errors.New("peer not allowed")
			}

			return nil
		}, "Peer not allowed yet")
}

func TestPrivateCoordinatorConfig(t *testing.T) {
	ctx := context.Background()
	configDir, _ := ioutil.TempDir("", "indigo-node")

	s := &swarm.Service{}
	config := s.Config().(swarm.Config)
	config.Addresses = []string{"/ip4/42.42.42.42/tcp/8903"}
	config.ProtectionMode = protector.PrivateWithCoordinatorMode
	config.CoordinatorConfig = &swarm.CoordinatorConfig{
		IsCoordinator: true,
		ConfigPath:    path.Join(configDir, "config.json"),
	}
	s.SetConfig(config)

	peerID, _ := peer.IDB58Decode(config.PeerID)

	cfg, err := swarm.NewProtectorConfig(&config)
	require.NoError(t, err)

	p, networkConfig, err := cfg.Configure(ctx, s, peerstore.NewPeerstore())
	assert.IsType(t, &protector.PrivateNetworkWithBootstrap{}, p)
	require.NotNil(t, networkConfig)

	waitUntilAllowed(t, peerID, networkConfig)
	assert.ElementsMatch(t, []peer.ID{peerID}, networkConfig.AllowedPeers(ctx))
}

func TestPrivateWithCoordinatorConfig(t *testing.T) {
	ctx := context.Background()
	configDir, _ := ioutil.TempDir("", "indigo-node")

	coordinatorID := test.GeneratePeerID(t)
	coordinatorAddr := test.GeneratePeerMultiaddr(t, coordinatorID)

	s := &swarm.Service{}
	config := s.Config().(swarm.Config)
	config.ProtectionMode = protector.PrivateWithCoordinatorMode
	config.CoordinatorConfig = &swarm.CoordinatorConfig{
		ConfigPath:           path.Join(configDir, "config.json"),
		CoordinatorID:        coordinatorID.Pretty(),
		CoordinatorAddresses: []string{coordinatorAddr.String()},
	}
	s.SetConfig(config)

	cfg, err := swarm.NewProtectorConfig(&config)
	require.NoError(t, err)

	pstore := peerstore.NewPeerstore()
	p, networkConfig, err := cfg.Configure(ctx, s, pstore)
	assert.IsType(t, &protector.PrivateNetwork{}, p)

	// Coordinator should be added to peer store.
	coordinatorInfo := pstore.PeerInfo(coordinatorID)
	require.NotNil(t, coordinatorInfo)
	assert.ElementsMatch(t, []multiaddr.Multiaddr{coordinatorAddr}, coordinatorInfo.Addrs)

	require.NotNil(t, networkConfig)
	waitUntilAllowed(t, coordinatorID, networkConfig)
	assert.ElementsMatch(t, []peer.ID{coordinatorID}, networkConfig.AllowedPeers(ctx))
}
