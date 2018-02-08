// Copyright © 2017-2018  Stratumn SAS
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

package state_test

import (
	"testing"

	"github.com/stratumn/alice/core/protocol/coin/state"
	"github.com/stratumn/alice/core/protocol/coin/testutil"
	"github.com/stretchr/testify/assert"
)

func TestInMemoryMempool(t *testing.T) {
	tests := []struct {
		name string
		test func(*testing.T, state.Mempool)
	}{{
		"empty-pool",
		func(t *testing.T, m state.Mempool) {
			assert.Nil(t, m.PopTransaction(), "m.PopTransaction()")
		},
	}, {
		"highest-fee",
		func(t *testing.T, m state.Mempool) {
			m.AddTransaction(testutil.NewTransaction(t, 5, 1, 1))
			m.AddTransaction(testutil.NewTransaction(t, 5, 3, 1))
			m.AddTransaction(testutil.NewTransaction(t, 5, 2, 1))

			tx := m.PopTransaction()
			assert.NotNil(t, tx, "m.PopTransaction()")
			assert.Equal(t, uint64(3), tx.Fee, "tx.Fee")
		},
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &state.GreedyInMemoryMempool{}
			tt.test(t, m)
		})
	}
}
