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

package state

import (
	"sync"

	"github.com/stratumn/go-indigonode/app/coin/pb"
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
