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

//go:generate mockgen -package mockstate -destination mockstate/mocktxpool.go github.com/stratumn/alice/app/coin/protocol/state txpool

package state

import (
	"sync"

	"github.com/stratumn/alice/app/coin/pb"
)

// TxPool stores transactions that need to be processed.
type TxPool interface {
	// AddTransaction adds a transaction to the pool.
	// It assumes that the transaction has been validated.
	AddTransaction(tx *pb.Transaction) error

	// PopTransaction pops the transaction with the highest score
	// from the txpool.
	// The score can be computed from various sources: transaction
	// fees, time in the txpool, priority, etc.
	// The txpool implementations can chose to prioritize fairness,
	// miner rewards, or anything else they come up with.
	PopTransaction() *pb.Transaction

	// Peek peeks at n transactions from the txpool
	// without modifying it.
	Peek(n uint32) []*pb.Transaction

	// Pending returns the number of transactions
	// waiting to be processed.
	Pending() uint64
}

// GreedyInMemoryTxPool is a very simple txpool that stores queued transactions
// in memory.
type GreedyInMemoryTxPool struct {
	mu  sync.RWMutex
	txs []*pb.Transaction
}

// AddTransaction adds a transaction in memory.
func (m *GreedyInMemoryTxPool) AddTransaction(tx *pb.Transaction) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.txs = append(m.txs, tx)
	return nil
}

// PopTransaction pops the transaction with the highest fee from the pool.
func (m *GreedyInMemoryTxPool) PopTransaction() *pb.Transaction {
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

// Peek returns the n oldest transactions from the pool.
func (m *GreedyInMemoryTxPool) Peek(n uint32) []*pb.Transaction {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if uint32(len(m.txs)) < n {
		n = uint32(len(m.txs))
	}

	res := make([]*pb.Transaction, n)
	copy(res, m.txs)

	return res
}

// Pending returns the number of transactions.
func (m *GreedyInMemoryTxPool) Pending() uint64 {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return uint64(len(m.txs))
}
