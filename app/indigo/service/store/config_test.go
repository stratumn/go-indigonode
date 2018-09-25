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

	"github.com/golang/mock/gomock"
	"github.com/stratumn/go-indigocore/cs/cstesting"
	"github.com/stratumn/go-indigocore/dummystore"
	"github.com/stratumn/go-indigocore/postgresstore"
	"github.com/stratumn/go-indigocore/utils"
	storeprotocol "github.com/stratumn/go-node/app/indigo/protocol/store"
	"github.com/stratumn/go-node/app/indigo/protocol/store/audit/dummyauditstore"
	"github.com/stratumn/go-node/app/indigo/protocol/store/audit/postgresauditstore"
	"github.com/stratumn/go-node/app/indigo/protocol/store/constants"
	"github.com/stratumn/go-node/app/indigo/service/store"
	swarmSvc "github.com/stratumn/go-node/core/app/swarm/service"
	"github.com/stratumn/go-node/core/protector"
	"github.com/stratumn/go-node/test"
	"github.com/stratumn/go-node/test/containers"
	"github.com/stratumn/go-node/test/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gx/ipfs/QmQsErDt8Qgw1XrsXf2BpEzDgGWtB1YLsTAARBup5b6B9W/go-libp2p-peer"
	"gx/ipfs/QmY1L5krVk8dv8d74uESmJTXGpoigVYqBVxXXz1aS8aFSb/go-libp2p-floodsub"
	"gx/ipfs/QmZNkThpqfVXs9GNbexPrfBbXSLNYeKrE7jwFM2oqHbyqN/go-libp2p-protocol"
)

func TestConfig_CreateStores(t *testing.T) {
	ctx := context.Background()
	peerID, _ := peer.IDB58Decode("QmPEeCgxxX6YbQWqkKuF42YCUpy4GdrqGLPMAFZ8A3A35d")
	link := cstesting.NewLinkBuilder().
		WithMetadata(constants.NodeIDKey, peerID.Pretty()).
		Build()

	t.Run("with dummy store", func(t *testing.T) {
		config := &store.Config{StorageType: store.InMemoryStorage}

		t.Run("CreateAuditStore", func(t *testing.T) {
			auditStore, err := config.CreateAuditStore(ctx)
			assert.NoError(t, err)
			assert.NotNil(t, auditStore)
			assert.IsType(t, &dummyauditstore.DummyAuditStore{}, auditStore)
			err = auditStore.AddSegment(ctx, link.Segmentify())
			assert.NoError(t, err)
		})

		t.Run("CreateIndigoStore", func(t *testing.T) {
			indigoStore, err := config.CreateIndigoStore(ctx)
			assert.NoError(t, err)
			assert.NotNil(t, indigoStore)
			assert.IsType(t, &dummystore.DummyStore{}, indigoStore)
		})
	})

	t.Run("with postgres store", func(t *testing.T) {
		container, err := containers.RunPostgres()
		assert.NoError(t, err)
		defer func() {
			err := utils.KillContainer(container)
			assert.NoError(t, err)
		}()

		config := &store.Config{StorageType: store.PostgreSQLStorage, PostgresConfig: &store.PostgresConfig{
			StorageDBURL: containers.PostgresTestURL,
		}}

		t.Run("CreateIndigoStore", func(t *testing.T) {
			indigoStore, err := config.CreateIndigoStore(ctx)
			require.NoError(t, err)
			assert.NotNil(t, indigoStore)
			assert.IsType(t, &postgresstore.Store{}, indigoStore)
			_, err = indigoStore.CreateLink(context.Background(), cstesting.RandomLink())
			assert.NoError(t, err)
		})

		t.Run("CreateIndigoStore with existing tables", func(t *testing.T) {
			indigoStore, err := config.CreateIndigoStore(ctx)
			require.NoError(t, err)
			assert.NotNil(t, indigoStore)
			assert.IsType(t, &postgresstore.Store{}, indigoStore)
			_, err = indigoStore.CreateLink(context.Background(), cstesting.RandomLink())
			assert.NoError(t, err)
		})

		t.Run("CreateIndigoStore with bad config", func(t *testing.T) {
			badConf := &store.Config{StorageType: store.PostgreSQLStorage}
			indigoStore, err := badConf.CreateIndigoStore(ctx)
			assert.EqualError(t, err, store.ErrMissingConfig.Error())
			assert.Nil(t, indigoStore)
		})

		t.Run("CreateAuditStore ", func(t *testing.T) {
			auditStore, err := config.CreateAuditStore(ctx)
			require.NoError(t, err)
			assert.NotNil(t, auditStore)
			assert.IsType(t, &postgresauditstore.PostgresAuditStore{}, auditStore)
			err = auditStore.AddSegment(ctx, link.Segmentify())
			assert.NoError(t, err)
		})

		t.Run("CreateAuditStore with existing tables", func(t *testing.T) {
			auditStore, err := config.CreateAuditStore(ctx)
			require.NoError(t, err)
			assert.NotNil(t, auditStore)
			assert.IsType(t, &postgresauditstore.PostgresAuditStore{}, auditStore)
			err = auditStore.AddSegment(ctx, link.Segmentify())
			assert.NoError(t, err)
		})

		t.Run("CreateAuditStore with bad config", func(t *testing.T) {
			badConf := &store.Config{StorageType: store.PostgreSQLStorage}
			auditStore, err := badConf.CreateAuditStore(ctx)
			assert.EqualError(t, err, store.ErrMissingConfig.Error())
			assert.Nil(t, auditStore)
		})
	})

	t.Run("with unsupported storage type", func(t *testing.T) {
		config := &store.Config{StorageType: "unknown"}

		t.Run("CreateAuditStore", func(t *testing.T) {
			auditStore, err := config.CreateAuditStore(ctx)
			assert.EqualError(t, err, store.ErrStorageNotSupported.Error())
			assert.Nil(t, auditStore)
		})

		t.Run("CreateAuditStore", func(t *testing.T) {
			indigoStore, err := config.CreateIndigoStore(ctx)
			assert.EqualError(t, err, store.ErrStorageNotSupported.Error())
			assert.Nil(t, indigoStore)
		})
	})
}

func TestConfig_CreateValidator(t *testing.T) {
	ctx := context.Background()

	t.Run("returns an error if no validation configuration is provided", func(t *testing.T) {
		config := &store.Config{}
		_, err := config.CreateValidator(ctx, dummystore.New(nil))
		assert.EqualError(t, err, "validation settings not found: missing configuration settings")
	})

	t.Run("returns an error if the provided store is nil", func(t *testing.T) {
		config := &store.Config{ValidationConfig: &store.ValidationConfig{}}
		_, err := config.CreateValidator(ctx, nil)
		assert.EqualError(t, err, "an indigo store adapter is needed to initialize a validator")
	})

	t.Run("returns a governance manager", func(t *testing.T) {
		config := &store.Config{ValidationConfig: &store.ValidationConfig{}}
		govMgr, err := config.CreateValidator(ctx, dummystore.New(nil))
		assert.NoError(t, err)
		assert.NotNil(t, govMgr)
	})
}

func TestConfig_JoinIndigoNetwork(t *testing.T) {
	t.Run("public-network", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		host := mocks.NewMockHost(ctrl)
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

		swarm := &swarmSvc.Swarm{}
		config := &store.Config{NetworkID: "indigo"}

		networkMgr, err := config.JoinIndigoNetwork(context.Background(), host, swarm)
		require.NoError(t, err)
		assert.IsType(t, &storeprotocol.PubSubNetworkManager{}, networkMgr)
	})

	t.Run("private-network", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		host := mocks.NewMockHost(ctrl)

		swarm := &swarmSvc.Swarm{NetworkMode: protector.NewCoordinatorNetworkMode()}
		config := &store.Config{NetworkID: "indigo"}

		networkMgr, err := config.JoinIndigoNetwork(context.Background(), host, swarm)
		require.NoError(t, err)
		assert.IsType(t, &storeprotocol.PrivateNetworkManager{}, networkMgr)
	})
}
