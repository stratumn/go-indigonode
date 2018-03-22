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

package coin

import (
	"context"
	"testing"

	"github.com/stratumn/alice/core/protocol/coin/chain"
	"github.com/stratumn/alice/core/protocol/coin/coinutil"
	"github.com/stratumn/alice/core/protocol/coin/state"
	tassert "github.com/stratumn/alice/core/protocol/coin/testutil/assert"
	"github.com/stratumn/alice/core/protocol/coin/testutil/blocktest"
	txtest "github.com/stratumn/alice/core/protocol/coin/testutil/transaction"
	"github.com/stratumn/alice/core/protocol/coin/validator"
	"github.com/stratumn/alice/lib/db"
	pb "github.com/stratumn/alice/pb/coin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	peer "gx/ipfs/QmZoWKhxUmZ2seW4BzX6fJkNR8hh9PsGModr7q171yq2SS/go-libp2p-peer"
)

func TestCoinMining_SingleNode(t *testing.T) {
	t.Run("start-stop-mining", func(t *testing.T) {
		c := NewCoinBuilder(t).WithMinerID("alice").Build(t)
		verifyAccount(t, c, "alice", &pb.Account{})

		ctx, cancel := context.WithCancel(context.Background())
		errChan := make(chan error)
		go func() {
			errChan <- c.StartMining(ctx)
		}()

		cancel()
		err := <-errChan
		assert.EqualError(t, err, context.Canceled.Error(), "<-errChan")

		verifyAccount(t, c, "alice", &pb.Account{})
	})

	t.Run("reject-invalid-txs", func(t *testing.T) {
		c := NewCoinBuilder(t).Build(t)
		err := c.AddTransaction(txtest.NewTransaction(t, 0, 1, 1))
		assert.EqualError(t, err, validator.ErrInvalidTxValue.Error(), "c.AddTransaction()")

		err = c.AddTransaction(txtest.NewTransaction(t, 42, 1, 3))
		assert.EqualError(t, err, state.ErrInsufficientBalance.Error(), "c.AddTransaction()")
	})

	t.Run("produce-blocks-from-old-txs", func(t *testing.T) {
		c := mineAllTransactions(
			t,
			&MiningTestConfig{
				initialBalance: 80,
				maxTxPerBlock:  2,
				minerID:        "bob",
				reward:         3,
				pendingTxs: []*pb.Transaction{
					txtest.NewTransaction(t, 20, 5, 3),
					txtest.NewTransaction(t, 10, 4, 4),
					txtest.NewTransaction(t, 5, 3, 5),
				},
				validTxsOnly: true,
				stopCond: func(c *Coin) bool {
					h, err := c.chain.CurrentHeader()
					assert.NoError(t, err, "chain.CurrentHeader()")
					return h.BlockNumber == 2
				},
			},
		)

		t.Run("adds-blocks-to-chain", func(t *testing.T) {
			h, err := c.chain.CurrentHeader()
			assert.NoError(t, err, "chain.CurrentHeader()")
			assert.Equal(t, uint64(2), h.BlockNumber, "h.BlockNumber")

			hs, err := c.chain.GetHeadersByNumber(2)
			assert.NoError(t, err, "chain.GetHeadersByNumber()")
			assert.Len(t, hs, 1, "chain.GetHeadersByNumber()")
			assert.Equal(t, h, hs[0], "chain.GetHeadersByNumber()")
		})

		t.Run("updates-miner-account", func(t *testing.T) {
			verifyAccount(t, c, "bob", &pb.Account{Balance: 2*3 + 5 + 4 + 3})
		})

		t.Run("updates-sender-account", func(t *testing.T) {
			verifyAccount(t, c, txtest.TxSenderPID,
				&pb.Account{Balance: 80 - 20 - 5 - 10 - 4 - 5 - 3, Nonce: 5})
		})

		t.Run("updates-receiver-account", func(t *testing.T) {
			verifyAccount(t, c, txtest.TxRecipientPID,
				&pb.Account{Balance: 20 + 10 + 5})
		})
	})

	t.Run("produce-blocks-live-txs", func(t *testing.T) {
		c := mineAllTransactions(
			t,
			&MiningTestConfig{
				initialBalance: 200,
				maxTxPerBlock:  3,
				minerID:        "carol",
				reward:         5,
				pendingTxs: []*pb.Transaction{
					// These three txs should be included in the same block,
					// so their order (nonce) should not matter.
					txtest.NewTransaction(t, 5, 2, 1),
					txtest.NewTransaction(t, 10, 3, 3),
					txtest.NewTransaction(t, 7, 10, 4),
				},
				liveTxs: []*pb.Transaction{
					// This tx should be rejected because of its nonce and low fee.
					txtest.NewTransaction(t, 5, 1, 3),
					txtest.NewTransaction(t, 5, 2, 6),
					txtest.NewTransaction(t, 5, 1, 7),
				},
				validTxsOnly: false,
				stopCond: func(c *Coin) bool {
					h, err := c.chain.CurrentHeader()
					assert.NoError(t, err, "chain.CurrentHeader()")
					return h.BlockNumber == 2
				},
			},
		)

		t.Run("adds-blocks-to-chain", func(t *testing.T) {
			h, err := c.chain.CurrentHeader()
			assert.NoError(t, err, "chain.CurrentHeader()")
			assert.Equal(t, uint64(2), h.BlockNumber, "h.BlockNumber")

			mh, err := coinutil.HashHeader(h)
			assert.NoError(t, err, "coinutil.HashHeader()")

			block, err := c.chain.GetBlock(mh, 2)
			assert.NoError(t, err, "chain.GetBlock()")
			verifyBlockTxs(t, block, []uint64{6, 7})

			block, err = c.chain.GetBlock(block.PreviousHash(), 1)
			assert.NoError(t, err, "chain.GetBlock()")
			verifyBlockTxs(t, block, []uint64{1, 3, 4})
		})

		t.Run("updates-miner-account", func(t *testing.T) {
			verifyAccount(t, c, "carol", &pb.Account{Balance: 2*5 + 2 + 3 + 10 + 2 + 1})
		})

		t.Run("updates-sender-account", func(t *testing.T) {
			verifyAccount(t, c, txtest.TxSenderPID,
				&pb.Account{Balance: 200 - 5 - 2 - 10 - 3 - 7 - 10 - 5 - 2 - 5 - 1, Nonce: 7})
		})

		t.Run("updates-receiver-account", func(t *testing.T) {
			verifyAccount(t, c, txtest.TxRecipientPID,
				&pb.Account{Balance: 5 + 10 + 7 + 5 + 5})
		})
	})
}

// MiningTestConfig contains the configuration of a coin miner run.
// For simplicity we only have one sender and one recipient (since
// the miner doesn't care about the transaction addresses, we don't
// lose test coverage with that assumpion).
type MiningTestConfig struct {
	initialBalance uint64  // Initial balance of the sender's account.
	maxTxPerBlock  uint32  // Max number of txs to include per block.
	minerID        peer.ID // ID of the miner.
	reward         uint64  // Miner reward for producing blocks.

	validTxsOnly bool              // Should the test fail if transactions are rejected
	pendingTxs   []*pb.Transaction // Transactions in the pool before mining starts.
	liveTxs      []*pb.Transaction // Transactions that will be received during mining.

	stopCond func(c *Coin) bool // Should return true when all valid txs have been included.
}

// mineAllTransactions creates a new chain and starts mining with the
// given configuration.
// It stops when all transactions have been included in blocks.
// You can then assert on the resulting state of different components.
func mineAllTransactions(t *testing.T, config *MiningTestConfig) *Coin {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	db, err := db.NewMemDB(nil)
	require.NoError(t, err, "db.NewMemDB()")

	s := state.NewState(db, state.OptPrefix([]byte("test")))
	err = s.UpdateAccount(
		[]byte(txtest.TxSenderPID),
		&pb.Account{Balance: config.initialBalance},
	)
	assert.NoError(t, err, "s.UpdateAccount()")

	genesisBlock := blocktest.NewBlock(
		t,
		[]*pb.Transaction{&pb.Transaction{
			Value: config.initialBalance,
			To:    []byte(txtest.TxSenderPID),
		},
		})
	genesisBlock.Header.BlockNumber = 0
	chain := chain.NewChainDB(db, chain.OptGenesisBlock(genesisBlock))

	c := NewCoinBuilder(t).
		WithChain(chain).
		WithMaxTxPerBlock(config.maxTxPerBlock).
		WithMinerID(config.minerID).
		WithReward(config.reward).
		WithState(s).
		Build(t)

	for _, tx := range config.pendingTxs {
		err := c.AddTransaction(tx)
		if config.validTxsOnly {
			assert.NoError(t, err, "c.AddTransaction()")
		}
	}

	go c.StartMining(ctx)
	tassert.WaitUntil(t, c.miner.IsRunning, "c.miner.IsRunning")

	for _, tx := range config.liveTxs {
		err := c.AddTransaction(tx)
		if config.validTxsOnly {
			assert.NoError(t, err, "c.AddTransaction()")
		}
	}

	// Wait until all required transactions are processed.
	tassert.WaitUntil(
		t,
		func() bool { return config.stopCond(c) },
		"all transactions processed",
	)

	return c
}

func verifyAccount(t *testing.T, c *Coin, id peer.ID, expected *pb.Account) {
	account, err := c.GetAccount([]byte(id))
	assert.NoError(t, err, "c.GetAccount()")
	assert.Equal(t, expected, account, "account")
}

func verifyBlockTxs(t *testing.T, block *pb.Block, expectedTxsNonces []uint64) {
	var blockTxNonces []uint64
	for _, tx := range block.Transactions {
		if tx.From != nil {
			blockTxNonces = append(blockTxNonces, tx.Nonce)
		}
	}

	assert.Equalf(t, len(expectedTxsNonces), len(blockTxNonces),
		"len(block.Transactions): %v", blockTxNonces)

	assert.ElementsMatch(t, expectedTxsNonces, blockTxNonces)
}
