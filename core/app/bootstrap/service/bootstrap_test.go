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

package service

import (
	"context"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/pkg/errors"
	"github.com/stratumn/go-node/core/app/bootstrap/protocol"
	swarm "github.com/stratumn/go-node/core/app/swarm/service"
	"github.com/stratumn/go-node/core/manager/testservice"
	"github.com/stratumn/go-node/core/protector"
	mockprotector "github.com/stratumn/go-node/core/protector/mocks"
	protectorpb "github.com/stratumn/go-node/core/protector/pb"
	"github.com/stratumn/go-node/test"
	"github.com/stratumn/go-node/test/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	inet "github.com/libp2p/go-libp2p-net"
	peer "github.com/libp2p/go-libp2p-peer"
	pstore "github.com/libp2p/go-libp2p-peerstore"
	"github.com/libp2p/go-libp2p-peerstore/pstoremem"
	"github.com/multiformats/go-multiaddr"
)

const (
	testPID  = "QmXMTDMJht9Q1KCYgJqVRzNG9kpXoujLXJkPouSowTiwwr"
	testAddr = "/ip4/127.0.0.1/tcp/54983/ipfs/" + testPID
)

func testService(
	ctx context.Context,
	t *testing.T,
	host Host,
	networkMode *protector.NetworkMode,
	networkConfig protector.NetworkConfig,
) *Service {
	serv := &Service{}
	config := serv.Config().(Config)

	config.Addresses = []string{testAddr}
	config.MinPeerThreshold = 1

	require.NoError(t, serv.SetConfig(config), "serv.SetConfig(config)")

	deps := map[string]interface{}{
		"host": host,
		"swarm": &swarm.Swarm{
			NetworkMode:   networkMode,
			NetworkConfig: networkConfig,
		},
	}

	require.NoError(t, serv.Plug(deps), "serv.Plug(deps)")

	return serv
}

func expectPublicHost(ctx context.Context, t *testing.T, net *mocks.MockNetwork, host *mocks.MockHost) {
	seedID, err := peer.IDB58Decode(testPID)
	require.NoError(t, err, "peer.IDB58Decode(testPID)")

	ps := pstoremem.NewPeerstore()

	host.EXPECT().Network().Return(net).AnyTimes()
	host.EXPECT().Peerstore().Return(ps).AnyTimes()
	host.EXPECT().Connect(gomock.Any(), gomock.Any()).Return(nil)

	net.EXPECT().Peers().Return(ps.Peers())
	net.EXPECT().Connectedness(seedID).Return(inet.NotConnected)
	net.EXPECT().Peers().Return([]peer.ID{seedID}).Times(2)
}

func TestService_strings(t *testing.T) {
	testservice.CheckStrings(t, &Service{})
}

func TestService_Expose(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	host := mocks.NewMockHost(ctrl)
	net := mocks.NewMockNetwork(ctrl)
	expectPublicHost(ctx, t, net, host)

	serv := testService(ctx, t, host, &protector.NetworkMode{}, nil)
	exposed := testservice.Expose(ctx, t, serv, time.Second)

	assert.Equal(t, struct{}{}, exposed, "exposed type")
}

func TestService_Run(t *testing.T) {
	t.Run("public-network", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		host := mocks.NewMockHost(ctrl)
		net := mocks.NewMockNetwork(ctrl)
		expectPublicHost(ctx, t, net, host)

		serv := testService(ctx, t, host, &protector.NetworkMode{}, nil)
		testservice.TestRun(ctx, t, serv, time.Second)
	})

	t.Run("private-network", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		hostID := test.GeneratePeerID(t)
		peer1 := test.GeneratePeerID(t)
		peer2 := test.GeneratePeerID(t)
		peer2addr := test.GeneratePeerMultiaddr(t, peer2)

		net := mocks.NewMockNetwork(ctrl)
		net.EXPECT().Peers().Return([]peer.ID{peer1, peer2}).AnyTimes()
		net.EXPECT().Connectedness(peer1).Return(inet.Connected).AnyTimes()
		net.EXPECT().Connectedness(peer2).Return(inet.NotConnected).AnyTimes()

		peerStore := mocks.NewMockPeerstore(ctrl)
		peerStore.EXPECT().PeerInfo(peer2).Return(pstore.PeerInfo{ID: peer2}).AnyTimes()
		// Simulate that peer2's address expired and is re-added
		peerStore.EXPECT().Addrs(peer2).Return(nil).AnyTimes()
		peerStore.EXPECT().AddAddrs(peer2, []multiaddr.Multiaddr{peer2addr}, gomock.Any()).AnyTimes()

		host := mocks.NewMockHost(ctrl)
		host.EXPECT().ID().Return(hostID).AnyTimes()
		host.EXPECT().Network().Return(net).AnyTimes()
		host.EXPECT().Peerstore().Return(peerStore).AnyTimes()

		host.EXPECT().SetStreamHandler(gomock.Any(), gomock.Any()).AnyTimes()
		host.EXPECT().RemoveStreamHandler(gomock.Any()).AnyTimes()

		host.EXPECT().Connect(gomock.Any(), pstore.PeerInfo{ID: peer2})
		host.EXPECT().NewStream(gomock.Any(), peer1, protocol.PrivateCoordinatedConfigPID).Return(nil, errors.New("no stream"))
		host.EXPECT().NewStream(gomock.Any(), peer2, protocol.PrivateCoordinatedConfigPID).Return(nil, errors.New("no stream"))

		networkCfg := mockprotector.NewMockNetworkConfig(ctrl)
		networkCfg.EXPECT().NetworkState(gomock.Any()).Return(protectorpb.NetworkState_PROTECTED)
		networkCfg.EXPECT().AllowedPeers(gomock.Any()).Return([]peer.ID{hostID, peer1, peer2}).AnyTimes()
		networkCfg.EXPECT().AllowedAddrs(gomock.Any(), peer2).Return([]multiaddr.Multiaddr{peer2addr}).AnyTimes()
		networkCfg.EXPECT().Copy(gomock.Any())

		networkMode := &protector.NetworkMode{
			ProtectionMode: protector.PrivateWithCoordinatorMode,
			IsCoordinator:  true,
		}

		serv := testService(ctx, t, host, networkMode, networkCfg)
		testservice.TestRun(ctx, t, serv, time.Second)
	})
}

func TestService_SetConfig(t *testing.T) {
	errAny := errors.New("any error")

	tests := []struct {
		name string
		set  func(*Config)
		err  error
	}{{
		"invalid interval",
		func(c *Config) { c.Interval = "1" },
		errAny,
	}, {
		"invalid timeout",
		func(c *Config) { c.ConnectionTimeout = "ten" },
		errAny,
	}, {
		"invalid address",
		func(c *Config) { c.Addresses = []string{"http://example.com"} },
		errAny,
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			serv := Service{}
			config := serv.Config().(Config)
			config.Addresses = []string{testAddr}
			tt.set(&config)

			err := errors.Cause(serv.SetConfig(config))
			switch {
			case err != nil && tt.err == errAny:
			case err != tt.err:
				assert.Equal(t, tt.err, err)
			}
		})
	}
}

func TestService_Needs(t *testing.T) {
	tests := []struct {
		name  string
		set   func(*Config)
		needs []string
	}{{
		"host",
		func(c *Config) { c.Host = "myhost" },
		[]string{"myhost", "swarm", "p2p"},
	}, {
		"needs",
		func(c *Config) { c.Needs = []string{"something"} },
		[]string{"host", "swarm", "something"},
	}}

	toSet := func(keys []string) map[string]struct{} {
		set := map[string]struct{}{}
		for _, v := range keys {
			set[v] = struct{}{}
		}

		return set
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			serv := Service{}
			config := serv.Config().(Config)
			config.Addresses = []string{testAddr}
			tt.set(&config)

			require.NoError(t, serv.SetConfig(config), "serv.SetConfig(config)")
			assert.Equal(t, toSet(tt.needs), serv.Needs())
		})
	}
}

func TestService_Plug(t *testing.T) {
	errAny := errors.New("any error")

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	host := mocks.NewMockHost(ctrl)

	tests := []struct {
		name string
		set  func(*Config)
		deps map[string]interface{}
		err  error
	}{{
		"valid host and swarm",
		func(c *Config) { c.Host = "myhost" },
		map[string]interface{}{
			"myhost": host,
			"swarm":  &swarm.Swarm{},
		},
		nil,
	}, {
		"invalid host",
		func(c *Config) { c.Host = "myhost" },
		map[string]interface{}{
			"myhost": struct{}{},
			"swarm":  &swarm.Swarm{},
		},
		ErrNotHost,
	}, {
		"invalid swarm",
		func(c *Config) { c.Swarm = "myswarm" },
		map[string]interface{}{
			"host":    host,
			"myswarm": struct{}{},
		},
		ErrNotSwarm,
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			serv := Service{}
			config := serv.Config().(Config)
			config.Addresses = []string{testAddr}
			tt.set(&config)

			require.NoError(t, serv.SetConfig(config), "serv.SetConfig(config)")

			err := errors.Cause(serv.Plug(tt.deps))
			switch {
			case err != nil && tt.err == errAny:
			case err != tt.err:
				assert.Equal(t, tt.err, err)
			}
		})
	}
}
