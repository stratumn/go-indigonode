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

	"github.com/stratumn/alice/app/coin/protocol/state"
	txtest "github.com/stratumn/alice/app/coin/protocol/testutil/transaction"
	"github.com/stretchr/testify/assert"
)

func TestGreedyInMemoryTxPool(t *testing.T) {
	tests := []struct {
		name string
		test func(*testing.T, state.TxPool)
	}{{
		"empty-pool",
		func(t *testing.T, txp state.TxPool) {
			assert.Nil(t, txp.PopTransaction(), "txp.PopTransaction()")
			assert.Equal(t, uint64(0), txp.Pending(), "txp.Pending()")
			assert.Len(t, txp.Peek(3), 0, "txp.Peek()")
		},
	}, {
		"non-empty-pool",
		func(t *testing.T, txp state.TxPool) {
			txp.AddTransaction(txtest.NewTransaction(t, 5, 1, 1))
			txp.AddTransaction(txtest.NewTransaction(t, 5, 3, 1))

			assert.Equal(t, uint64(2), txp.Pending(), "txp.Pending()")
			assert.Len(t, txp.Peek(3), 2, "txp.Peek()")
		},
	}, {
		"highest-fee",
		func(t *testing.T, txp state.TxPool) {
			txp.AddTransaction(txtest.NewTransaction(t, 5, 1, 1))
			txp.AddTransaction(txtest.NewTransaction(t, 5, 3, 1))
			txp.AddTransaction(txtest.NewTransaction(t, 5, 2, 1))

			tx := txp.PopTransaction()
			assert.NotNil(t, tx, "txp.PopTransaction()")
			assert.Equal(t, uint64(3), tx.Fee, "tx.Fee")
		},
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			txp := &state.GreedyInMemoryTxPool{}
			tt.test(t, txp)
		})
	}
}
