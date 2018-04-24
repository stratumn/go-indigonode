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

package store_test

import (
	"context"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stratumn/alice/core/manager/testservice"
	storeprotocol "github.com/stratumn/alice/core/protocol/indigo/store"
	"github.com/stratumn/alice/core/service/indigo/store"
	"github.com/stratumn/alice/core/service/indigo/store/mockstore"
	"github.com/stratumn/alice/core/service/pubsub/mockpubsub"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	protocol "gx/ipfs/QmZNkThpqfVXs9GNbexPrfBbXSLNYeKrE7jwFM2oqHbyqN/go-libp2p-protocol"
	floodsub "gx/ipfs/QmctbcXMMhxTjm5ybWpjMwDmabB39ANuhB5QNn8jpD4JTv/go-libp2p-floodsub"
)

func testService(ctx context.Context, t *testing.T, host *mockstore.MockHost) *store.Service {
	serv := &store.Service{}
	config := serv.Config().(store.Config)
	config.Version = "1.0.0"
	config.NetworkID = "42"
	config.PrivateKey = "CAESYKecc4tj7XAXruOYfd4m61d3mvxJUUdUVwIuFbB/PYFAtAoPM/Pbft/aS3mc5jFkb2dScZS61XOl9PnU3uDWuPq0Cg8z89t+39pLeZzmMWRvZ1JxlLrVc6X0+dTe4Na4+g=="

	require.NoError(t, serv.SetConfig(config), "serv.SetConfig(config)")

	deps := map[string]interface{}{
		"host": host,
	}

	require.NoError(t, serv.Plug(deps), "serv.Plug(deps)")

	return serv
}

func TestService_Strings(t *testing.T) {
	testservice.CheckStrings(t, &store.Service{})
}

// expectHostNetwork verifies that the service joins a PoP network via floodsub.
func expectHostNetwork(ctrl *gomock.Controller, host *mockstore.MockHost) {
	net := mockpubsub.NewMockNetwork(ctrl)

	host.EXPECT().Network().Return(net)
	net.EXPECT().Notify(gomock.Any())
	host.EXPECT().SetStreamHandler(protocol.ID(floodsub.FloodSubID), gomock.Any())
	host.EXPECT().SetStreamHandler(protocol.ID(storeprotocol.SingleNodeSyncProtocolID), gomock.Any())
	host.EXPECT().RemoveStreamHandler(protocol.ID(floodsub.FloodSubID))
	host.EXPECT().RemoveStreamHandler(protocol.ID(storeprotocol.SingleNodeSyncProtocolID))
}

func TestService_Run(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	host := mockstore.NewMockHost(ctrl)
	expectHostNetwork(ctrl, host)

	serv := testService(ctx, t, host)
	testservice.TestRun(ctx, t, serv, time.Second)
}

func TestService_Run_Error(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	tests := []struct {
		name   string
		config store.Config
	}{{
		"missing-network-id",
		store.Config{
			Version:     "1.0.0",
			StorageType: "in-memory",
			PrivateKey:  "CAESYKecc4tj7XAXruOYfd4m61d3mvxJUUdUVwIuFbB/PYFAtAoPM/Pbft/aS3mc5jFkb2dScZS61XOl9PnU3uDWuPq0Cg8z89t+39pLeZzmMWRvZ1JxlLrVc6X0+dTe4Na4+g==",
		},
	}, {
		"missing-private-key",
		store.Config{
			Version:     "1.0.0",
			StorageType: "in-memory",
			NetworkID:   "42",
		},
	}, {
		"invalid-storage-type",
		store.Config{
			Version:     "1.0.0",
			StorageType: "on-the-moon",
			NetworkID:   "42",
			PrivateKey:  "CAESYKecc4tj7XAXruOYfd4m61d3mvxJUUdUVwIuFbB/PYFAtAoPM/Pbft/aS3mc5jFkb2dScZS61XOl9PnU3uDWuPq0Cg8z89t+39pLeZzmMWRvZ1JxlLrVc6X0+dTe4Na4+g==",
		},
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			serv := &store.Service{}
			require.NoError(t, serv.SetConfig(tt.config), "serv.SetConfig(config)")
			assert.Error(t, serv.Run(ctx, func() {}, func() {}))
		})
	}
}
