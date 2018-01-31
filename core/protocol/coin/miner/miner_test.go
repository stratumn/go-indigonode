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
	"time"

	"github.com/stratumn/alice/core/protocol/coin/testutil"
	"github.com/stretchr/testify/assert"
)

func TestMiner_StartStop(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	m := NewMinerBuilder(ctx).Build()

	testutil.WaitUntil(t, m.IsRunning, "m.IsRunning()")

	cancel()

	testutil.WaitUntil(t, func() bool { return !m.IsRunning() }, "!m.IsRunning()")
}

func TestMiner_Mempool(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	t.Run("Pops transactions from mempool", func(t *testing.T) {
		mempool := &testutil.InMemoryMempool{}
		m := NewMinerBuilder(ctx).WithMempool(mempool).Build()

		assert.Equal(t, 0, mempool.TxCount(), "mempool.TxCount()")

		mempool.AddTransaction(testutil.NewTransaction(t, 1, 1))
		mempool.AddTransaction(testutil.NewTransaction(t, 1, 2))

		testutil.WaitUntil(
			t,
			func() bool { return mempool.TxCount() == 0 },
			"mempool.TxCount() == 0",
		)

		assert.True(t, m.IsRunning(), "m.IsRunning()")
	})

	t.Run("Puts transactions back into mempool if block production failed", func(t *testing.T) {
		mempool := &testutil.InMemoryMempool{}
		m := NewMinerBuilder(ctx).
			WithMempool(mempool).
			WithEngine(&testutil.FaultyEngine{}).
			Build()

		mempool.AddTransaction(testutil.NewTransaction(t, 1, 1))
		testutil.WaitUntil(
			t,
			func() bool { return mempool.PopCount() >= 1 },
			"mempool.PopCount() >= 1",
		)

		// Transaction should be put back in the mempool after the engine error.
		testutil.WaitUntil(
			t,
			func() bool { return mempool.TxCount() == 1 },
			"mempool.TxCount() == 1",
		)

		assert.True(t, m.IsRunning(), "m.IsRunning()")
	})

	t.Run("Removes invalid transactions from mempool definitively", func(t *testing.T) {
		mempool := &testutil.InMemoryMempool{}
		m := NewMinerBuilder(ctx).
			WithMempool(mempool).
			WithValidator(&testutil.Rejector{}).
			Build()

		mempool.AddTransaction(testutil.NewTransaction(t, 1, 1))
		testutil.WaitUntil(
			t,
			func() bool { return mempool.PopCount() >= 1 },
			"mempool.PopCount() == 1",
		)

		// Wait a bit before verifying that the transaction
		// was not put back in the mempool.
		<-time.After(10 * time.Millisecond)

		assert.Equal(t, 0, mempool.TxCount(), "mempool.TxCount()")
		assert.True(t, m.IsRunning(), "m.IsRunning()")
	})
}
