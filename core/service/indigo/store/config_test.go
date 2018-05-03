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

	peer "gx/ipfs/QmZoWKhxUmZ2seW4BzX6fJkNR8hh9PsGModr7q171yq2SS/go-libp2p-peer"

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
)

func TestConfig_UnmarshalPrivateKey(t *testing.T) {
	tests := []struct {
		name    string
		config  *store.Config
		success bool
	}{{
		"missing-key",
		&store.Config{},
		false,
	}, {
		"invalid-key",
		&store.Config{PrivateKey: "sp0ng3b0b"},
		false,
	}, {
		"valid-key",
		&store.Config{PrivateKey: "CAESYKecc4tj7XAXruOYfd4m61d3mvxJUUdUVwIuFbB/PYFAtAoPM/Pbft/aS3mc5jFkb2dScZS61XOl9PnU3uDWuPq0Cg8z89t+39pLeZzmMWRvZ1JxlLrVc6X0+dTe4Na4+g=="},
		true,
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sk, err := tt.config.UnmarshalPrivateKey()
			if tt.success {
				assert.NoError(t, err)
				assert.NotNil(t, sk)
			} else {
				assert.Error(t, err)
				assert.Nil(t, sk)
			}
		})
	}
}

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

		t.Run("CreateIndigoStore ", func(t *testing.T) {
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
			auditStore, err := badConf.CreateIndigoStore(ctx)
			assert.EqualError(t, err, store.ErrMissingConfig.Error())
			assert.Nil(t, auditStore)
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
