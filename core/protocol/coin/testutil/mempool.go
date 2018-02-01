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

	pb "github.com/stratumn/alice/pb/coin"
)

// InMemoryMempool is a basic mempool implementation that stores
// transactions in RAM.
type InMemoryMempool struct {
	mu  sync.RWMutex
	txs []*pb.Transaction

	popCount uint32
}

// AddTransaction adds transaction to the mempool.
func (m *InMemoryMempool) AddTransaction(tx *pb.Transaction) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.txs = append(m.txs, tx)
	return nil
}

// Contains returns true if the mempool contains the given transaction.
func (m *InMemoryMempool) Contains(tx *pb.Transaction) bool {
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

// TxCount returns the number of transactions in the mempool.
func (m *InMemoryMempool) TxCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return len(m.txs)
}

// PopTransaction pops the oldest transaction from the mempool.
func (m *InMemoryMempool) PopTransaction() *pb.Transaction {
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
func (m *InMemoryMempool) PopCount() uint32 {
	return m.popCount
}
