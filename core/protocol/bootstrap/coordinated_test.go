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
	"github.com/stratumn/alice/core/protocol/bootstrap/bootstraptesting"
	protectorpb "github.com/stratumn/alice/pb/protector"
	"github.com/stratumn/alice/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	netutil "gx/ipfs/Qmb6BsZf6Y3kxffXMNTubGPF1w1bkHtpvhfYbmnwP3NQyw/go-libp2p-netutil"
	bhost "gx/ipfs/Qmc64U41EEB4nPG7wxjEqFwKJajS2f8kk5q2TvUrQf78Xu/go-libp2p-blankhost"
	"gx/ipfs/QmcJukH2sAFjY3HdBKq35WDzWoL3UUu2gt9wdfqZTUyM74/go-libp2p-peer"
)

func generateCoordinatedNetworkMode(t *testing.T, peerID peer.ID) *protector.NetworkMode {
	peerAddr := test.GeneratePeerMultiaddr(t, peerID)

	mode, err := protector.NewCoordinatedNetworkMode(
		peerID.Pretty(),
		[]string{peerAddr.String()},
	)

	require.NoError(t, err)
	return mode
}

func TestCoordinated_Handshake_Error(t *testing.T) {
	testCases := []struct {
		name                 string
		configureCoordinator func(*testing.T, *gomock.Controller, *bootstraptesting.CoordinatorBuilder, peer.ID)
		err                  error
	}{{
		"fails-if-coordinator-unavailable",
		func(_ *testing.T, _ *gomock.Controller, b *bootstraptesting.CoordinatorBuilder, _ peer.ID) {
			b.Unavailable().Build()
		},
		protector.ErrConnectionRefused,
	}, {
		"fails-if-coordinated-not-allowed",
		func(t *testing.T, ctrl *gomock.Controller, b *bootstraptesting.CoordinatorBuilder, _ peer.ID) {
			networkConfig := bootstraptesting.NewNetworkConfigBuilder(t, ctrl).
				WithNetworkState(protectorpb.NetworkState_PROTECTED).
				Build()

			b.WithNetworkConfig(networkConfig).Build()
		},
		protector.ErrConnectionRefused,
	}, {
		"fails-if-coordinator-sends-empty-config",
		func(t *testing.T, ctrl *gomock.Controller, b *bootstraptesting.CoordinatorBuilder, _ peer.ID) {
			networkConfig := bootstraptesting.NewNetworkConfigBuilder(t, ctrl).
				WithNetworkState(protectorpb.NetworkState_BOOTSTRAP).
				Build()

			b.WithNetworkConfig(networkConfig).Build()
		},
		protector.ErrConnectionRefused,
	}, {
		"fails-if-coordinator-sends-invalid-signature",
		func(t *testing.T, ctrl *gomock.Controller, b *bootstraptesting.CoordinatorBuilder, peerID peer.ID) {
			networkConfig := bootstraptesting.NewNetworkConfigBuilder(t, ctrl).
				WithNetworkState(protectorpb.NetworkState_BOOTSTRAP).
				WithAllowedPeer(peerID).
				WithInvalidSignature().
				Build()

			b.WithNetworkConfig(networkConfig).Build()
		},
		protectorpb.ErrInvalidSignature,
	}}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			h := bhost.NewBlankHost(netutil.GenSwarmNetwork(t, ctx))
			defer h.Close()

			coordinatorBuilder := bootstraptesting.NewCoordinatorBuilder(ctx, t)
			defer coordinatorBuilder.Close(ctx)

			require.NoError(t, h.Connect(ctx, coordinatorBuilder.Host().Peerstore().PeerInfo(coordinatorBuilder.PeerID())))

			tt.configureCoordinator(t, ctrl, coordinatorBuilder, h.ID())

			networkMode := generateCoordinatedNetworkMode(t, coordinatorBuilder.PeerID())
			handler, err := bootstrap.NewCoordinatedHandler(ctx, h, networkMode, nil)
			assert.EqualError(t, err, tt.err.Error())
			assert.Nil(t, handler)
		})
	}
}
