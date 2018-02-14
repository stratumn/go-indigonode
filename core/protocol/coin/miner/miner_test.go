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

	"github.com/stratumn/alice/core/protocol/coin/engine"
	"github.com/stratumn/alice/core/protocol/coin/processor"

	"github.com/stratumn/alice/core/protocol/coin/testutil"
	tassert "github.com/stratumn/alice/core/protocol/coin/testutil/assert"
	"github.com/stretchr/testify/assert"
)

func TestMiner_StartStop(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	m := NewMinerBuilder(t).Build()

	assert.False(t, m.IsRunning(), "m.IsRunning()")

	errChan := make(chan error)
	go func() {
		errChan <- m.Start(ctx)
	}()

	tassert.WaitUntil(t, m.IsRunning, "m.IsRunning()")

	cancel()
	assert.EqualError(t, <-errChan, context.Canceled.Error(), "m.Start(ctx)")
	assert.False(t, m.IsRunning(), "m.IsRunning()")
}

func TestMiner_TxPool(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	t.Run("Pops transactions from txpool", func(t *testing.T) {
		txpool := &testutil.InMemoryTxPool{}
		m := NewMinerBuilder(t).WithTxPool(txpool).Build()
		go m.Start(ctx)

		assert.Equal(t, 0, txpool.TxCount(), "txpool.TxCount()")

		txpool.AddTransaction(testutil.NewTransaction(t, 1, 1, 1))
		txpool.AddTransaction(testutil.NewTransaction(t, 1, 1, 2))

		tassert.WaitUntil(
			t,
			func() bool { return txpool.TxCount() == 0 },
			"txpool.TxCount() == 0",
		)

		assert.True(t, m.IsRunning(), "m.IsRunning()")
	})

	t.Run("Puts transactions back into txpool if block production failed", func(t *testing.T) {
		txpool := &testutil.InMemoryTxPool{}
		m := NewMinerBuilder(t).
			WithTxPool(txpool).
			WithEngine(&testutil.FaultyEngine{}).
			Build()
		go m.Start(ctx)

		txpool.AddTransaction(testutil.NewTransaction(t, 1, 1, 1))
		tassert.WaitUntil(
			t,
			func() bool { return txpool.PopCount() >= 1 },
			"txpool.PopCount() >= 1",
		)

		// Transaction should be put back in the txpool after the engine error.
		tassert.WaitUntil(
			t,
			func() bool { return txpool.TxCount() == 1 },
			"txpool.TxCount() == 1",
		)

		assert.True(t, m.IsRunning(), "m.IsRunning()")
	})

	t.Run("Removes invalid transactions from txpool definitively", func(t *testing.T) {
		txpool := &testutil.InMemoryTxPool{}
		m := NewMinerBuilder(t).
			WithTxPool(txpool).
			WithValidator(&testutil.Rejector{}).
			Build()
		go m.Start(ctx)

		txpool.AddTransaction(testutil.NewTransaction(t, 1, 1, 1))
		tassert.WaitUntil(
			t,
			func() bool { return txpool.PopCount() >= 1 },
			"txpool.PopCount() == 1",
		)

		// Wait a bit before verifying that the transaction
		// was not put back in the txpool.
		<-time.After(10 * time.Millisecond)

		assert.Equal(t, 0, txpool.TxCount(), "txpool.TxCount()")
		assert.True(t, m.IsRunning(), "m.IsRunning()")
	})
}

func TestMiner_Produce(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start a miner with a txpool containing a valid transaction.
	startMiner := func(p processor.Processor, e engine.Engine) *testutil.InMemoryTxPool {
		txpool := &testutil.InMemoryTxPool{}
		txpool.AddTransaction(testutil.NewTransaction(t, 3, 1, 5))

		m := NewMinerBuilder(t).
			WithEngine(e).
			WithTxPool(txpool).
			WithProcessor(p).
			Build()
		go m.Start(ctx)

		return txpool
	}

	t.Run("Sends new valid block to processor", func(t *testing.T) {
		processor := testutil.NewInstrumentedProcessor(&testutil.DummyProcessor{})
		startMiner(processor, &testutil.DummyEngine{})

		tassert.WaitUntil(
			t,
			func() bool { return processor.ProcessedCount() > 0 },
			"p.ProcessedCount() > 0",
		)
	})

	t.Run("Aborts block if engine returns an error", func(t *testing.T) {
		processor := testutil.NewInstrumentedProcessor(&testutil.DummyProcessor{})
		engine := testutil.NewInstrumentedEngine(&testutil.FaultyEngine{})
		startMiner(processor, engine)

		tassert.WaitUntil(
			t,
			func() bool { return engine.PrepareCount() > 0 },
			"engine.PrepareCount() > 0",
		)

		assert.Equal(t, uint32(0), processor.ProcessedCount(), "processor.ProcessedCount()")
	})

	t.Run("Aborts block if processor returns an error", func(t *testing.T) {
		processor := testutil.NewInstrumentedProcessor(&testutil.FaultyProcessor{})
		txpool := startMiner(processor, &testutil.DummyEngine{})

		tassert.WaitUntil(
			t,
			func() bool { return processor.ProcessedCount() > 0 },
			"processor.ProcessedCount() > 0",
		)

		// If the transaction goes back to the txpool it means the block
		// was correctly aborted.
		tassert.WaitUntil(
			t,
			func() bool { return txpool.TxCount() > 0 },
			"txpool.TxCount() > 0",
		)
	})
}
