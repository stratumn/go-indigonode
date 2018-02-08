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

//go:generate mockgen -package mockstate -destination mockstate/mockmempool.go github.com/stratumn/alice/core/protocol/coin/state Mempool

package state

import (
	"sync"

	pb "github.com/stratumn/alice/pb/coin"
)

// Mempool stores transactions that need to be processed.
type Mempool interface {
	// AddTransaction adds a transaction to the mempool.
	// It assumes that the transaction has been validated.
	AddTransaction(tx *pb.Transaction) error

	// PopTransaction pops the transaction with the highest score
	// from the mempool.
	// The score can be computed from various sources: transaction
	// fees, time in the mempool, priority, etc.
	// The mempool implementations can chose to prioritize fairness,
	// miner rewards, or anything else they come up with.
	PopTransaction() *pb.Transaction
}

// GreedyInMemoryMempool is a very simple mempool that stores queued transactions
// in memory.
type GreedyInMemoryMempool struct {
	mu  sync.RWMutex
	txs []*pb.Transaction
}

// AddTransaction adds a transaction in memory.
func (m *GreedyInMemoryMempool) AddTransaction(tx *pb.Transaction) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.txs = append(m.txs, tx)
	return nil
}

// PopTransaction pops the transaction with the highest fee from the pool.
func (m *GreedyInMemoryMempool) PopTransaction() *pb.Transaction {
	m.mu.Lock()
	defer m.mu.Unlock()

	if len(m.txs) == 0 {
		return nil
	}

	highestFee := m.txs[0]
	highestFeeIndex := 0
	for i, tx := range m.txs {
		if highestFee.Fee < tx.Fee {
			highestFee = tx
			highestFeeIndex = i
		}
	}

	m.txs[highestFeeIndex] = m.txs[len(m.txs)-1]
	m.txs = m.txs[:len(m.txs)-1]
	return highestFee
}
