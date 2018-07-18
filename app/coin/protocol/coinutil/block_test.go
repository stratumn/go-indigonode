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

package coinutil_test

import (
	"testing"

	"github.com/stratumn/go-indigonode/app/coin/pb"
	"github.com/stratumn/go-indigonode/app/coin/protocol/coinutil"
	txtest "github.com/stratumn/go-indigonode/app/coin/protocol/testutil/transaction"
	"github.com/stretchr/testify/assert"
)

func TestGetMinerReward(t *testing.T) {
	t.Run("Returns single reward", func(t *testing.T) {
		tx, err := coinutil.GetMinerReward(&pb.Block{
			Transactions: []*pb.Transaction{
				txtest.NewTransaction(t, 1, 1, 1),
				txtest.NewTransaction(t, 1, 1, 1),
				txtest.NewRewardTransaction(t, 5),
				txtest.NewTransaction(t, 1, 1, 1),
				txtest.NewTransaction(t, 1, 1, 1),
			},
		})

		assert.NoError(t, err, "GetMinerReward()")
		assert.Equal(t, uint64(5), tx.Value, "tx.Value")
		assert.Nil(t, tx.From, "tx.From")
	})

	t.Run("Returns error if multiple rewards", func(t *testing.T) {
		_, err := coinutil.GetMinerReward(&pb.Block{
			Transactions: []*pb.Transaction{
				txtest.NewTransaction(t, 1, 1, 1),
				txtest.NewRewardTransaction(t, 5),
				txtest.NewRewardTransaction(t, 10),
			},
		})

		assert.EqualError(t, err, coinutil.ErrMultipleMinerRewards.Error(), "GetMinerReward()")
	})
}

func TestGetBlockFees(t *testing.T) {
	totalFees := coinutil.GetBlockFees(&pb.Block{
		Transactions: []*pb.Transaction{
			txtest.NewTransaction(t, 1, 1, 1),
			txtest.NewTransaction(t, 1, 2, 1),
			// Miner reward because empty sender
			&pb.Transaction{
				Value: 5,
				// The fee should be silently ignored
				Fee: 15,
			},
			txtest.NewTransaction(t, 1, 3, 1),
			txtest.NewTransaction(t, 1, 4, 1),
		},
	})

	assert.Equal(t, uint64(1+2+3+4), totalFees, "totalFees")
}
