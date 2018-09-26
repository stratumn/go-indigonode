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

package miner

import (
	"context"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stratumn/go-node/app/coin/pb"
	"github.com/stratumn/go-node/app/coin/protocol/chain"
	"github.com/stratumn/go-node/app/coin/protocol/engine"
	"github.com/stratumn/go-node/app/coin/protocol/engine/mockengine"
	"github.com/stratumn/go-node/app/coin/protocol/processor"
	"github.com/stratumn/go-node/app/coin/protocol/processor/mockprocessor"
	"github.com/stratumn/go-node/app/coin/protocol/state"
	"github.com/stratumn/go-node/app/coin/protocol/testutil"
	tassert "github.com/stratumn/go-node/app/coin/protocol/testutil/assert"
	txtest "github.com/stratumn/go-node/app/coin/protocol/testutil/transaction"
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

		txpool.AddTransaction(txtest.NewTransaction(t, 1, 1, 1))
		txpool.AddTransaction(txtest.NewTransaction(t, 1, 1, 2))

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

		txpool.AddTransaction(txtest.NewTransaction(t, 1, 1, 1))
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

		txpool.AddTransaction(txtest.NewTransaction(t, 1, 1, 1))
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

	testChain := &testutil.SimpleChain{}

	// Start a miner with a txpool containing a valid transaction.
	startMiner := func(p processor.Processor, e engine.Engine) *Miner {
		txpool := &testutil.InMemoryTxPool{}
		txpool.AddTransaction(txtest.NewTransaction(t, 3, 1, 5))

		m := NewMinerBuilder(t).
			WithChain(testChain).
			WithEngine(e).
			WithTxPool(txpool).
			WithProcessor(p).
			Build()
		go m.Start(ctx)

		return m
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
		m := startMiner(processor, &testutil.DummyEngine{})

		tassert.WaitUntil(
			t,
			func() bool { return processor.ProcessedCount() > 0 },
			"processor.ProcessedCount() > 0",
		)

		// If the transaction goes back to the txpool it means the block
		// was correctly aborted.
		tassert.WaitUntil(
			t,
			func() bool { return m.txpool.(*testutil.InMemoryTxPool).TxCount() > 0 },
			"txpool.TxCount() > 0",
		)
	})

	t.Run("Aborts block if new head is received", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		e := mockengine.NewMockPoW(ctrl)
		e.EXPECT().Prepare(gomock.Any(), gomock.Any()).Return(nil).Times(2)

		// Simulate a long block production that will be cancelled.
		e.EXPECT().
			Finalize(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Return(&pb.Block{}, nil).
			Do(func(context.Context, chain.Reader, *pb.Header, state.Reader, []*pb.Transaction) {
				<-time.After(5 * time.Second)
			})

		// We'll receive a new block from an external source (gossip)
		// and should start mining on it.
		gossipedNewHead := &pb.Header{BlockNumber: 42}

		// The block we produce should take into account the received new head.
		p := mockprocessor.NewMockProcessor(ctrl)
		p.EXPECT().
			Process(
				gomock.Any(),
				&pb.Block{Header: &pb.Header{BlockNumber: gossipedNewHead.BlockNumber + 1}},
				gomock.Any(),
				gomock.Any()).
			Return(nil)
		processor := testutil.NewInstrumentedProcessor(p)

		m := startMiner(processor, e)

		// If the first Finalize is correctly cancelled, this second call to Finalize will be made.
		e.EXPECT().
			Finalize(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Return(&pb.Block{Header: &pb.Header{BlockNumber: gossipedNewHead.BlockNumber + 1}}, nil)

		assert.NoError(t, testChain.SetHead(&pb.Block{Header: gossipedNewHead}))
		m.newBlockChan <- gossipedNewHead

		tassert.WaitUntil(t, func() bool { return processor.ProcessedCount() == 1 }, "processor.ProcessedCount() == 1")
	})

	t.Run("Ignores new block if it's not the head", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		unblockProduceChan := make(chan struct{})
		e := mockengine.NewMockPoW(ctrl)
		e.EXPECT().Prepare(gomock.Any(), gomock.Any()).Return(nil).Times(1)
		e.EXPECT().
			Finalize(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Return(&pb.Block{Header: &pb.Header{BlockNumber: 42}}, nil).
			Do(func(context.Context, chain.Reader, *pb.Header, state.Reader, []*pb.Transaction) {
				<-unblockProduceChan
			})

		p := mockprocessor.NewMockProcessor(ctrl)
		p.EXPECT().
			Process(
				gomock.Any(),
				&pb.Block{Header: &pb.Header{BlockNumber: 42}},
				gomock.Any(),
				gomock.Any()).
			Return(nil)
		processor := testutil.NewInstrumentedProcessor(p)

		// We keep the chain head at a value different from the incoming blocks.
		assert.NoError(t, testChain.SetHead(&pb.Block{Header: &pb.Header{BlockNumber: 41}}))

		m := startMiner(processor, e)

		m.newBlockChan <- &pb.Header{BlockNumber: 13}
		m.newBlockChan <- &pb.Header{BlockNumber: 22}

		unblockProduceChan <- struct{}{}

		tassert.WaitUntil(t, func() bool { return processor.ProcessedCount() == 1 }, "processor.ProcessedCount() == 1")
	})
}
