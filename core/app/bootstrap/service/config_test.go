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

package service_test

import (
	"context"
	"testing"

	"github.com/stratumn/alice/core/app/bootstrap/protocol/proposal"
	"github.com/stratumn/alice/core/app/bootstrap/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStoreConfig(t *testing.T) {
	t.Run("unknown-store-type", func(t *testing.T) {
		_, err := (&service.Config{
			StoreConfig: &service.StoreConfig{
				Type: "IPFS",
			},
		}).NewStore(context.Background())

		assert.EqualError(t, err, service.ErrInvalidStoreType.Error())
	})

	t.Run("default-memory-store", func(t *testing.T) {
		s, err := (&service.Config{}).NewStore(context.Background())
		require.NoError(t, err)
		assert.IsType(t, &proposal.InMemoryStore{}, s)
	})

	t.Run("configure-memory-store", func(t *testing.T) {
		s, err := (&service.Config{
			StoreConfig: &service.StoreConfig{
				Type: service.InMemoryStore,
			},
		}).NewStore(context.Background())

		require.NoError(t, err)
		assert.IsType(t, &proposal.InMemoryStore{}, s)
	})

	t.Run("configure-file-store", func(t *testing.T) {
		s, err := (&service.Config{
			StoreConfig: &service.StoreConfig{
				Type: service.FileStore,
			},
		}).NewStore(context.Background())

		require.NoError(t, err)
		assert.IsType(t, &proposal.FileSaver{}, s)
	})
}
