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

	"github.com/stratumn/alice/core/protector"
	"github.com/stratumn/alice/core/protocol/bootstrap/bootstraptesting"
	protectorpb "github.com/stratumn/alice/pb/protector"
	"github.com/stratumn/alice/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newNetworkConfig(t *testing.T) protector.NetworkConfig {
	config, err := protector.NewInMemoryConfig(
		context.Background(),
		protectorpb.NewNetworkConfig(protectorpb.NetworkState_BOOTSTRAP),
	)
	require.NoError(t, err, "protector.NewInMemoryConfig()")
	return config
}

func TestCoordinated_Handshake(t *testing.T) {
	t.Run("coordinator-unavailable", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		testNetwork := bootstraptesting.NewTestNetwork(ctx, t)
		defer testNetwork.Close()

		unavailableCoordinatorID := test.GeneratePeerID(t)

		config := newNetworkConfig(t)
		handler, err := testNetwork.AddCoordinatedNode(
			unavailableCoordinatorID,
			config,
		)

		assert.EqualError(t, err, protector.ErrConnectionRefused.Error())
		assert.Nil(t, handler)
	})

	t.Run("coordinator-closes-conn", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		testNetwork := bootstraptesting.NewTestNetwork(ctx, t)
		defer testNetwork.Close()

		coordinatorConfig := newNetworkConfig(t)
		coordinatorConfig.SetNetworkState(ctx, protectorpb.NetworkState_PROTECTED)

		_, err := testNetwork.AddCoordinatorNode(coordinatorConfig)
		require.NoError(t, err, "testNetwork.AddCoordinatorNode()")

		coordinatorID := testNetwork.CoordinatorID()
		coordinatedConfig := newNetworkConfig(t)
		handler, err := testNetwork.AddCoordinatedNode(
			coordinatorID,
			coordinatedConfig,
		)

		assert.EqualError(t, err, protector.ErrConnectionRefused.Error())
		assert.Nil(t, handler)
	})

	t.Run("coordinator-invalid-signature", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		testNetwork := bootstraptesting.NewTestNetwork(ctx, t)
		defer testNetwork.Close()

		coordinatorConfig := newNetworkConfig(t)
		_, err := testNetwork.AddCoordinatorNode(coordinatorConfig)
		require.NoError(t, err, "testNetwork.AddCoordinatorNode()")

		coordinatorID := testNetwork.CoordinatorID()
		coordinatedConfig := newNetworkConfig(t)
		coordinatedNode, connect := testNetwork.PrepareCoordinatedNode(
			coordinatorID,
			coordinatedConfig,
		)

		err = coordinatorConfig.AddPeer(ctx, coordinatedNode.ID(), coordinatedNode.Addrs())
		require.NoError(t, err, "coordinatorConfig.AddPeer()")

		unknownKey := test.GeneratePrivateKey(t)
		err = coordinatorConfig.Sign(ctx, unknownKey)
		require.NoError(t, err, "coordinatorConfig.Sign()")

		handler, err := connect()
		assert.EqualError(t, err, protectorpb.ErrInvalidSignature.Error())
		assert.Nil(t, handler)
	})

	t.Run("coordinator-empty-config", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		testNetwork := bootstraptesting.NewTestNetwork(ctx, t)
		defer testNetwork.Close()

		coordinatorConfig := newNetworkConfig(t)
		_, err := testNetwork.AddCoordinatorNode(coordinatorConfig)
		require.NoError(t, err, "testNetwork.AddCoordinatorNode()")

		err = coordinatorConfig.Sign(ctx, testNetwork.CoordinatorKey())
		require.NoError(t, err, "coordinatorConfig.Sign()")

		coordinatedConfig := newNetworkConfig(t)
		_, connect := testNetwork.PrepareCoordinatedNode(
			testNetwork.CoordinatorID(),
			coordinatedConfig,
		)

		handler, err := connect()
		assert.NoError(t, err)
		assert.NotNil(t, handler)

		// When the coordinator returns an empty config, this is not a handshake error.
		// It means we're not whitelisted yet, but the network is still bootstrapping.
		assert.Len(t, coordinatedConfig.AllowedPeers(ctx), 0)
	})

	t.Run("coordinator-valid-config", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		testNetwork := bootstraptesting.NewTestNetwork(ctx, t)
		defer testNetwork.Close()

		coordinatorConfig := newNetworkConfig(t)
		_, err := testNetwork.AddCoordinatorNode(coordinatorConfig)
		require.NoError(t, err, "testNetwork.AddCoordinatorNode()")

		coordinatedConfig := newNetworkConfig(t)
		coordinatedNode, connect := testNetwork.PrepareCoordinatedNode(
			testNetwork.CoordinatorID(),
			coordinatedConfig,
		)

		err = coordinatorConfig.AddPeer(
			ctx,
			testNetwork.CoordinatorID(),
			test.GeneratePeerMultiaddrs(t, testNetwork.CoordinatorID()),
		)
		require.NoError(t, err, "coordinatorConfig.AddPeer")

		err = coordinatorConfig.AddPeer(
			ctx,
			coordinatedNode.ID(),
			coordinatedNode.Addrs(),
		)
		require.NoError(t, err, "coordinatorConfig.AddPeer")

		err = coordinatorConfig.Sign(ctx, testNetwork.CoordinatorKey())
		require.NoError(t, err, "coordinatorConfig.Sign()")

		handler, err := connect()
		assert.NoError(t, err)
		assert.NotNil(t, handler)

		assert.Len(t, coordinatedConfig.AllowedPeers(ctx), 2)
		assert.True(t, coordinatedConfig.IsAllowed(ctx, testNetwork.CoordinatorID()))
		assert.True(t, coordinatedConfig.IsAllowed(ctx, coordinatedNode.ID()))
	})
}
