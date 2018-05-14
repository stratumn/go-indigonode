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

	"github.com/stratumn/alice/core/protocol/indigo/store/audit/dummyauditstore"
	"github.com/stratumn/alice/core/protocol/indigo/store/audit/postgresauditstore"
	"github.com/stratumn/alice/core/protocol/indigo/store/constants"
	"github.com/stratumn/alice/core/service/indigo/store"
	"github.com/stratumn/alice/test/containers"
	"github.com/stratumn/go-indigocore/cs/cstesting"
	"github.com/stratumn/go-indigocore/dummystore"
	"github.com/stratumn/go-indigocore/postgresstore"
	"github.com/stratumn/go-indigocore/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	peer "gx/ipfs/QmcJukH2sAFjY3HdBKq35WDzWoL3UUu2gt9wdfqZTUyM74/go-libp2p-peer"
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
