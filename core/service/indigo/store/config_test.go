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
	"testing"

	"github.com/stratumn/alice/core/protocol/indigo/store/audit/dummyauditstore"
	"github.com/stratumn/alice/core/service/indigo/store"
	"github.com/stratumn/go-indigocore/dummystore"
	"github.com/stretchr/testify/assert"
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
	config := &store.Config{StorageType: "in-memory"}

	auditStore, err := config.CreateAuditStore()
	assert.NoError(t, err)
	assert.NotNil(t, auditStore)
	_, ok := auditStore.(*dummyauditstore.DummyAuditStore)
	assert.True(t, ok)

	indigoStore, err := config.CreateIndigoStore()
	assert.NoError(t, err)
	assert.NotNil(t, indigoStore)
	_, ok = indigoStore.(*dummystore.DummyStore)
	assert.True(t, ok)
}
