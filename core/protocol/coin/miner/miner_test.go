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

package miner

import (
	"context"
	"testing"

	"github.com/stratumn/alice/core/protocol/coin/testutil"
	"github.com/stretchr/testify/assert"
)

func TestMiner_StartStop(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	m := NewMiner(ctx, nil, nil, nil, nil, nil, nil)

	testutil.WaitUntil(t, m.IsRunning, "m.IsRunning()")

	cancel()

	testutil.WaitUntil(t, func() bool { return !m.IsRunning() }, "!m.IsRunning()")
}

func TestMiner_Mempool(t *testing.T) {
	t.Run("Pops transactions from mempool", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		mempool := &testutil.InMemoryMempool{}
		m := NewMiner(ctx, mempool, nil, nil, nil, nil, nil)

		assert.Equal(t, 0, mempool.TxCount(), "mempool.TxCount()")

		t1 := testutil.NewTransaction(t, 1, 1)
		mempool.AddTransaction(t1)
		txs := <-m.txsChan
		assert.True(t, testutil.NewTxMatcher(t1).Matches(txs[0]), "TxMatcher.Matches(tx)")

		t2 := testutil.NewTransaction(t, 1, 2)
		mempool.AddTransaction(t2)
		txs = <-m.txsChan
		assert.True(t, testutil.NewTxMatcher(t2).Matches(txs[0]), "TxMatcher.Matches(tx)")

		assert.Equal(t, 0, mempool.TxCount(), "mempool.TxCount()")
		assert.True(t, m.IsRunning(), "m.IsRunning()")
	})
}
