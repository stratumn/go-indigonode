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

package state_test

import (
	"testing"

	"github.com/stratumn/go-node/app/coin/protocol/state"
	txtest "github.com/stratumn/go-node/app/coin/protocol/testutil/transaction"
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
