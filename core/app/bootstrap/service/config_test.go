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

package service_test

import (
	"context"
	"testing"

	"github.com/stratumn/go-node/core/app/bootstrap/protocol/proposal"
	"github.com/stratumn/go-node/core/app/bootstrap/service"
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
