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

package testutil

import (
	"sync"
	"sync/atomic"

	"github.com/stratumn/go-indigonode/app/coin/pb"
)

// InMemoryTxPool is a basic txpool implementation that stores
// transactions in RAM.
type InMemoryTxPool struct {
	mu  sync.RWMutex
	txs []*pb.Transaction

	popCount uint32
}

// AddTransaction adds transaction to the txpool.
func (m *InMemoryTxPool) AddTransaction(tx *pb.Transaction) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.txs = append(m.txs, tx)
	return nil
}

// Contains returns true if the txpool contains the given transaction.
func (m *InMemoryTxPool) Contains(tx *pb.Transaction) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	matcher := NewTxMatcher(tx)
	for _, txx := range m.txs {
		if matcher.Matches(txx) {
			return true
		}
	}

	return false
}

// TxCount returns the number of transactions in the txpool.
func (m *InMemoryTxPool) TxCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return len(m.txs)
}

// PopTransaction pops the oldest transaction from the txpool.
func (m *InMemoryTxPool) PopTransaction() *pb.Transaction {
	m.mu.Lock()
	defer m.mu.Unlock()

	if len(m.txs) == 0 {
		return nil
	}

	tx := m.txs[len(m.txs)-1]
	m.txs = m.txs[:len(m.txs)-1]

	atomic.AddUint32(&m.popCount, 1)

	return tx
}

// PopCount returns the number of times PopTransaction was called.
func (m *InMemoryTxPool) PopCount() uint32 {
	return m.popCount
}

// Peek returns the n first transactions from the pool.
func (m *InMemoryTxPool) Peek(n uint32) []*pb.Transaction {
	m.mu.Lock()
	defer m.mu.Unlock()

	if uint32(len(m.txs)) < n {
		n = uint32(len(m.txs))
	}

	res := make([]*pb.Transaction, n)
	copy(res, m.txs)

	return res
}

// Pending returns the number of transactions.
func (m *InMemoryTxPool) Pending() uint64 {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return uint64(len(m.txs))
}
