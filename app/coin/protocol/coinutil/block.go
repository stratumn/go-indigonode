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

package coinutil

import (
	"errors"

	"github.com/stratumn/alice/app/coin/pb"
)

var (
	// ErrMultipleMinerRewards is returned when a block contains multiple miner rewards.
	ErrMultipleMinerRewards = errors.New("only one miner reward transaction is allowed per block")
)

// GetMinerReward verifies that there is a single reward transaction
// in the block and returns it.
func GetMinerReward(block *pb.Block) (*pb.Transaction, error) {
	found := false
	var reward *pb.Transaction
	for _, tx := range block.Transactions {
		if tx.From == nil {
			if found {
				return nil, ErrMultipleMinerRewards
			}

			reward = tx
			found = true
		}
	}

	return reward, nil
}

// GetTxCount returns the number of user transactions in a block.
// It skips the miner rewards.
func GetTxCount(block *pb.Block) uint32 {
	txCount := uint32(0)
	for _, tx := range block.Transactions {
		if tx.From == nil {
			continue
		}

		txCount++
	}

	return txCount
}

// GetBlockFees sums the fees from all transactions in a block.
func GetBlockFees(block *pb.Block) uint64 {
	fees := uint64(0)
	for _, tx := range block.Transactions {
		if tx.From != nil {
			fees += tx.Fee
		}
	}
	return fees
}
