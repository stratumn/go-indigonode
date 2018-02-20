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

package coinutil_test

import (
	"testing"

	"github.com/stratumn/alice/core/protocol/coin/coinutil"
	txtest "github.com/stratumn/alice/core/protocol/coin/testutil/transaction"
	pb "github.com/stratumn/alice/pb/coin"
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
