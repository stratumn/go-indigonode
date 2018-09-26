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

package coinutil

import (
	"errors"

	"github.com/stratumn/go-node/app/coin/pb"
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
