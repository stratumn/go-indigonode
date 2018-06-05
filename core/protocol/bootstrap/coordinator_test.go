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

package bootstrap_test

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stratumn/alice/core/protector"
	"github.com/stratumn/alice/core/protocol/bootstrap"
	pb "github.com/stratumn/alice/pb/bootstrap"
	protectorpb "github.com/stratumn/alice/pb/protector"
	"github.com/stratumn/alice/test"
	"github.com/stratumn/alice/test/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	protobuf "gx/ipfs/QmRDePEiL4Yupq5EkcK3L3ko3iMgYaqUdLu7xc1kqs7dnV/go-multicodec/protobuf"
	"gx/ipfs/QmVxf27kucSvCLiCq6dAXjDU2WG3xZN9ae7Ny6osroP28u/yamux"
	inet "gx/ipfs/QmXoz9o2PT3tEzf7hicegwex5UgVP54n3k82K7jrWFyN86/go-libp2p-net"
	netutil "gx/ipfs/Qmb6BsZf6Y3kxffXMNTubGPF1w1bkHtpvhfYbmnwP3NQyw/go-libp2p-netutil"
	bhost "gx/ipfs/Qmc64U41EEB4nPG7wxjEqFwKJajS2f8kk5q2TvUrQf78Xu/go-libp2p-blankhost"
)

func TestCoordinator_Close(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	host := mocks.NewMockHost(ctrl)
	host.EXPECT().SetStreamHandler(bootstrap.PrivateCoordinatorProtocolID, gomock.Any()).Times(1)

	handler, err := bootstrap.NewCoordinatorHandler(host, nil)
	require.NoError(t, err)
	require.NotNil(t, handler)

	host.EXPECT().RemoveStreamHandler(bootstrap.PrivateCoordinatorProtocolID).Times(1)
	handler.Close(context.Background())
}

func TestCoordinator_Handle_Hello(t *testing.T) {
	sendHello := func(t *testing.T, stream inet.Stream) {
		enc := protobuf.Multicodec(nil).Encoder(stream)
		require.NoError(t, enc.Encode(&pb.Hello{}), "enc.Encode()")
	}

	testCases := []struct {
		name          string
		networkConfig func(context.Context, *bhost.BlankHost, *bhost.BlankHost) protector.NetworkConfig
		send          func(*testing.T, inet.Stream)
		validate      func(*testing.T, *protectorpb.NetworkConfig, *bhost.BlankHost, *bhost.BlankHost)
		receiveErr    error
	}{{
		"during-bootstrap-send-participants-to-white-listed-peer",
		func(ctx context.Context, coordinator, sender *bhost.BlankHost) protector.NetworkConfig {
			networkConfig, _ := protector.NewInMemoryConfig(
				ctx,
				protectorpb.NewNetworkConfig(protectorpb.NetworkState_BOOTSTRAP),
			)

			networkConfig.AddPeer(
				ctx,
				coordinator.ID(),
				test.GeneratePeerMultiaddrs(t, coordinator.ID()),
			)
			networkConfig.AddPeer(
				ctx,
				sender.ID(),
				test.GeneratePeerMultiaddrs(t, sender.ID()),
			)

			return networkConfig
		},
		sendHello,
		func(t *testing.T, networkConfig *protectorpb.NetworkConfig, coordinator, sender *bhost.BlankHost) {
			assert.Equal(t, protectorpb.NetworkState_BOOTSTRAP, networkConfig.NetworkState)
			assert.Len(t, networkConfig.Participants, 2)
			assert.Contains(t, networkConfig.Participants, sender.ID().Pretty())
			assert.Contains(t, networkConfig.Participants, coordinator.ID().Pretty())
		},
		nil,
	}, {
		"during-bootstrap-do-not-send-participants-to-non-white-listed-peer",
		func(ctx context.Context, coordinator, sender *bhost.BlankHost) protector.NetworkConfig {
			networkConfig, _ := protector.NewInMemoryConfig(
				ctx,
				protectorpb.NewNetworkConfig(protectorpb.NetworkState_BOOTSTRAP),
			)

			networkConfig.AddPeer(
				ctx,
				coordinator.ID(),
				test.GeneratePeerMultiaddrs(t, coordinator.ID()),
			)

			return networkConfig
		},
		sendHello,
		func(t *testing.T, networkConfig *protectorpb.NetworkConfig, coordinator, sender *bhost.BlankHost) {
			assert.Nil(t, networkConfig.Signature, "networkConfig.Signature")
			assert.Zero(t, networkConfig.NetworkState, "networkConfig.NetworkState")
			assert.Nil(t, networkConfig.Participants, "networkConfig.Participants")
		},
		nil,
	}, {
		"after-bootstrap-send-participants-to-white-listed-peer",
		func(ctx context.Context, coordinator, sender *bhost.BlankHost) protector.NetworkConfig {
			networkConfig, _ := protector.NewInMemoryConfig(
				ctx,
				protectorpb.NewNetworkConfig(protectorpb.NetworkState_PROTECTED),
			)

			networkConfig.AddPeer(
				ctx,
				sender.ID(),
				test.GeneratePeerMultiaddrs(t, sender.ID()),
			)

			return networkConfig
		},
		sendHello,
		func(t *testing.T, networkConfig *protectorpb.NetworkConfig, coordinator, sender *bhost.BlankHost) {
			assert.Equal(t, protectorpb.NetworkState_PROTECTED, networkConfig.NetworkState)
			assert.Len(t, networkConfig.Participants, 1)
			assert.Contains(t, networkConfig.Participants, sender.ID().Pretty())
		},
		nil,
	}, {
		"after-bootstrap-reject-non-white-listed-peer",
		func(ctx context.Context, coordinator, sender *bhost.BlankHost) protector.NetworkConfig {
			networkConfig, _ := protector.NewInMemoryConfig(
				ctx,
				protectorpb.NewNetworkConfig(protectorpb.NetworkState_PROTECTED),
			)

			networkConfig.AddPeer(
				ctx,
				coordinator.ID(),
				test.GeneratePeerMultiaddrs(t, coordinator.ID()),
			)

			return networkConfig
		},
		sendHello,
		nil,
		yamux.ErrConnectionReset,
	}}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			coordinator := bhost.NewBlankHost(netutil.GenSwarmNetwork(t, ctx))
			defer coordinator.Close()

			sender := bhost.NewBlankHost(netutil.GenSwarmNetwork(t, ctx))
			defer sender.Close()

			require.NoError(t, sender.Connect(ctx, coordinator.Peerstore().PeerInfo(coordinator.ID())))

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			handler, err := bootstrap.NewCoordinatorHandler(
				coordinator,
				tt.networkConfig(ctx, coordinator, sender),
			)
			require.NoError(t, err)
			defer handler.Close(ctx)

			stream, err := sender.NewStream(ctx, coordinator.ID(), bootstrap.PrivateCoordinatorProtocolID)
			// If the coordinator is expected to reject streams, it can happen either
			// when initiating the stream (sender.NewStream()) or when writing to it (below).
			// It depends on the underlying yamux implementation but both are ok for our usecase.
			if err != nil {
				require.Error(t, tt.receiveErr)
				require.EqualError(t, err, tt.receiveErr.Error())
				return
			}

			tt.send(t, stream)

			dec := protobuf.Multicodec(nil).Decoder(stream)
			var response protectorpb.NetworkConfig
			err = dec.Decode(&response)

			if tt.receiveErr != nil {
				assert.EqualError(t, err, tt.receiveErr.Error())
			} else {
				require.NoError(t, err)
				tt.validate(t, &response, coordinator, sender)
			}
		})
	}
}
