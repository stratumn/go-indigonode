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

package store_test

import (
	"context"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/pkg/errors"
	storeprotocol "github.com/stratumn/go-indigonode/app/indigo/protocol/store"
	"github.com/stratumn/go-indigonode/app/indigo/protocol/store/sync"
	"github.com/stratumn/go-indigonode/app/indigo/service/store"
	swarm "github.com/stratumn/go-indigonode/core/app/swarm/service"
	"github.com/stratumn/go-indigonode/core/manager/testservice"
	"github.com/stratumn/go-indigonode/test"
	"github.com/stratumn/go-indigonode/test/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gx/ipfs/QmY1L5krVk8dv8d74uESmJTXGpoigVYqBVxXXz1aS8aFSb/go-libp2p-floodsub"
	"gx/ipfs/QmZNkThpqfVXs9GNbexPrfBbXSLNYeKrE7jwFM2oqHbyqN/go-libp2p-protocol"
)

func testService(ctx context.Context, t *testing.T, host *mocks.MockHost) *store.Service {
	serv := &store.Service{}
	config := serv.Config().(store.Config)
	config.Version = "1.0.0"
	config.NetworkID = "42"

	require.NoError(t, serv.SetConfig(config), "serv.SetConfig(config)")

	deps := map[string]interface{}{
		"host":  host,
		"swarm": &swarm.Swarm{},
	}

	require.NoError(t, serv.Plug(deps), "serv.Plug(deps)")

	return serv
}

func TestService_Strings(t *testing.T) {
	testservice.CheckStrings(t, &store.Service{})
}

// expectHostNetwork verifies that the service joins a PoP network via floodsub.
func expectHostNetwork(t *testing.T, ctrl *gomock.Controller, host *mocks.MockHost) {
	privKey := test.GeneratePrivateKey(t)
	peerID := test.GetPeerIDFromKey(t, privKey)

	net := mocks.NewMockNetwork(ctrl)
	net.EXPECT().Notify(gomock.Any())

	peerStore := mocks.NewMockPeerstore(ctrl)
	peerStore.EXPECT().PrivKey(peerID).Return(privKey).AnyTimes()

	host.EXPECT().ID().Return(peerID).AnyTimes()
	host.EXPECT().Network().Return(net)
	host.EXPECT().Peerstore().Return(peerStore).AnyTimes()
	host.EXPECT().SetStreamHandler(protocol.ID(floodsub.FloodSubID), gomock.Any())
	host.EXPECT().SetStreamHandler(protocol.ID(sync.SingleNodeProtocolID), gomock.Any())
	host.EXPECT().RemoveStreamHandler(protocol.ID(floodsub.FloodSubID))
	host.EXPECT().RemoveStreamHandler(protocol.ID(sync.SingleNodeProtocolID))
}

func TestService_Run(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	host := mocks.NewMockHost(ctrl)
	expectHostNetwork(t, ctrl, host)

	serv := testService(ctx, t, host)
	testservice.TestRun(ctx, t, serv, time.Second)
}

func TestService_Run_Error(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sk := test.GeneratePrivateKey(t)
	peerID := test.GetPeerIDFromKey(t, sk)

	tests := []struct {
		name   string
		err    error
		config store.Config
	}{{
		"missing-network-id",
		storeprotocol.ErrInvalidNetworkID,
		store.Config{
			Version:          "1.0.0",
			StorageType:      "in-memory",
			ValidationConfig: &store.ValidationConfig{},
		},
	}, {
		"invalid-storage-type",
		store.ErrStorageNotSupported,
		store.Config{
			Version:          "1.0.0",
			StorageType:      "on-the-moon",
			NetworkID:        "42",
			ValidationConfig: &store.ValidationConfig{},
		},
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			peerStore := mocks.NewMockPeerstore(ctrl)
			peerStore.EXPECT().PrivKey(peerID).Return(sk).AnyTimes()

			host := mocks.NewMockHost(ctrl)
			host.EXPECT().ID().Return(peerID).AnyTimes()
			host.EXPECT().Peerstore().Return(peerStore).AnyTimes()

			serv := testService(ctx, t, host)
			require.NoError(t, serv.SetConfig(tt.config), "serv.SetConfig(config)")
			assert.EqualError(t, serv.Run(ctx, func() {}, func() {}), tt.err.Error())
		})
	}
}

func TestService_Plug(t *testing.T) {
	sk := test.GeneratePrivateKey(t)
	peerID := test.GetPeerIDFromKey(t, sk)

	tests := []struct {
		name string
		set  func(*store.Config)
		deps func(*mocks.MockHost, *mocks.MockPeerstore) map[string]interface{}
		err  error
	}{{
		"valid-private-key",
		func(c *store.Config) { c.Swarm = "myswarm" },
		func(host *mocks.MockHost, peerStore *mocks.MockPeerstore) map[string]interface{} {
			peerStore.EXPECT().PrivKey(peerID).Return(sk)

			return map[string]interface{}{
				"host":    host,
				"myswarm": &swarm.Swarm{},
			}
		},
		nil,
	}, {
		"missing-private-key",
		func(c *store.Config) { c.Swarm = "myswarm" },
		func(host *mocks.MockHost, peerStore *mocks.MockPeerstore) map[string]interface{} {
			peerStore.EXPECT().PrivKey(peerID).Return(nil)

			return map[string]interface{}{
				"host":    host,
				"myswarm": &swarm.Swarm{},
			}
		},
		store.ErrMissingPrivateKey,
	}, {
		"missing-swarm",
		func(c *store.Config) { c.Swarm = "myswarm" },
		func(host *mocks.MockHost, peerStore *mocks.MockPeerstore) map[string]interface{} {
			peerStore.EXPECT().PrivKey(peerID).Return(sk)

			return map[string]interface{}{
				"host": host,
			}
		},
		store.ErrNotSwarm,
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			peerStore := mocks.NewMockPeerstore(ctrl)

			host := mocks.NewMockHost(ctrl)
			host.EXPECT().ID().Return(peerID)
			host.EXPECT().Peerstore().Return(peerStore)

			serv := store.Service{}
			config := serv.Config().(store.Config)
			config.Host = "host"
			tt.set(&config)

			require.NoError(t, serv.SetConfig(config), "serv.SetConfig(config)")

			err := errors.Cause(serv.Plug(tt.deps(host, peerStore)))
			switch {
			case err != tt.err:
				assert.Equal(t, tt.err, err)
			}
		})
	}
}
