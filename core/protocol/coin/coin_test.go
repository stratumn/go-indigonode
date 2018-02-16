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

	"github.com/golang/mock/gomock"
	"github.com/stratumn/alice/core/p2p"
	"github.com/stratumn/alice/core/protocol/coin/chain"
	"github.com/stratumn/alice/core/protocol/coin/chain/mockchain"
	"github.com/stratumn/alice/core/protocol/coin/coinutil"
	"github.com/stratumn/alice/core/protocol/coin/db"
	"github.com/stratumn/alice/core/protocol/coin/state"
	ctestutil "github.com/stratumn/alice/core/protocol/coin/testutil"
	tassert "github.com/stratumn/alice/core/protocol/coin/testutil/assert"
	"github.com/stratumn/alice/core/protocol/coin/testutil/blocktest"
	"github.com/stratumn/alice/core/protocol/coin/validator"
	pb "github.com/stratumn/alice/pb/coin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	inet "gx/ipfs/QmQm7WmgYCa4RSz76tKEYpRjApjnRw8ZTUVQC15b8JM4a2/go-libp2p-net"
	protobuf "gx/ipfs/QmRDePEiL4Yupq5EkcK3L3ko3iMgYaqUdLu7xc1kqs7dnV/go-multicodec/protobuf"
	testutil "gx/ipfs/QmV1axkk86DDkYwS269AvPy9eV5h7mUyHveJkSVHPjrQtY/go-libp2p-netutil"
	ihost "gx/ipfs/QmfCtHMCd9xFvehvHeVxtKVXJTMVTuHhyPRVHEXetn87vL/go-libp2p-host"
)

func TestCoinProtocolHandler(t *testing.T) {
	ctx := context.Background()
	hosts := make([]ihost.Host, 2)
	coins := make([]*Coin, 2)

	for i := 0; i < 2; i++ {
		hosts[i] = p2p.NewHost(ctx, testutil.GenSwarmNetwork(t, ctx))
		defer func(h ihost.Host) {
			h.Close()
		}(hosts[i])

		// We configure validation to fail.
		// We only want to test that the coin protocol
		// correctly called the validator.
		coins[i] = &Coin{validator: ctestutil.NewInstrumentedValidator(&ctestutil.Rejector{})}

		ii := i
		hosts[i].SetStreamHandler(ProtocolID, func(stream inet.Stream) {
			coins[ii].StreamHandler(ctx, stream)
		})
	}

	require.NoError(t, hosts[0].Connect(ctx, hosts[1].Peerstore().PeerInfo(hosts[1].ID())), "Connect()")
	require.NoError(t, hosts[1].Connect(ctx, hosts[0].Peerstore().PeerInfo(hosts[0].ID())), "Connect()")

	t.Run("Send transactions and blocks on the same stream", func(t *testing.T) {
		s0_1, err := hosts[0].NewStream(ctx, hosts[1].ID(), ProtocolID)
		require.NoError(t, err, "NewStream()")
		defer s0_1.Close()

		s1_0, err := hosts[1].NewStream(ctx, hosts[0].ID(), ProtocolID)
		require.NoError(t, err, "NewStream()")
		defer s1_0.Close()

		enc0_1 := protobuf.Multicodec(nil).Encoder(s0_1)
		enc1_0 := protobuf.Multicodec(nil).Encoder(s1_0)

		block1 := ctestutil.RandomGossipBlock()
		err = enc0_1.Encode(block1)
		assert.NoError(t, err, "Encode()")

		block2 := ctestutil.RandomGossipBlock()
		err = enc1_0.Encode(block2)
		assert.NoError(t, err, "Encode()")

		tx := ctestutil.RandomGossipTx()
		err = enc0_1.Encode(tx)
		assert.NoError(t, err, "Encode()")

		v0 := coins[0].validator.(*ctestutil.InstrumentedValidator)
		v1 := coins[1].validator.(*ctestutil.InstrumentedValidator)

		tassert.WaitUntil(t, func() bool {
			return v0.ValidatedBlock(block2.GetBlock()) &&
				v1.ValidatedBlock(block1.GetBlock()) &&
				v1.ValidatedTx(tx.GetTx())
		}, "validator.ValidatedBlock()")
	})

	t.Run("Ignore invalid messages", func(t *testing.T) {
		s1_0, err := hosts[1].NewStream(ctx, hosts[0].ID(), ProtocolID)
		require.NoError(t, err, "NewStream()")
		defer s1_0.Close()

		enc1_0 := protobuf.Multicodec(nil).Encoder(s1_0)
		err = enc1_0.Encode(&pb.Transaction{Value: 42})
		assert.NoError(t, err, "Encode()")
	})

	t.Run("Open and close streams", func(t *testing.T) {
		s1_0, err := hosts[1].NewStream(ctx, hosts[0].ID(), ProtocolID)
		require.NoError(t, err, "NewStream()")
		defer s1_0.Close()

		s0_1, err := hosts[0].NewStream(ctx, hosts[1].ID(), ProtocolID)
		require.NoError(t, err, "NewStream()")

		tx1 := ctestutil.RandomGossipTx()
		enc0_1 := protobuf.Multicodec(nil).Encoder(s0_1)
		err = enc0_1.Encode(tx1)
		assert.NoError(t, err, "Encode()")

		err = s0_1.Close()
		assert.NoError(t, err, "Close()")

		tx2 := ctestutil.RandomGossipTx()
		enc1_0 := protobuf.Multicodec(nil).Encoder(s1_0)
		err = enc1_0.Encode(tx2)
		assert.NoError(t, err, "Encode()")

		v0 := coins[0].validator.(*ctestutil.InstrumentedValidator)
		v1 := coins[1].validator.(*ctestutil.InstrumentedValidator)

		tassert.WaitUntil(t, func() bool {
			return v1.ValidatedTx(tx1.GetTx()) && v0.ValidatedTx(tx2.GetTx())
		}, "validator.ValidatedTx")
	})
}

func TestGetAccountTransactions(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	chain := mockchain.NewMockChain(ctrl)

	alice := []byte("alice")

	c := NewCoinBuilder(t).WithChain(chain).Build(t)
	tx := ctestutil.NewTransaction(t, 1, 1, 1)
	expected := []*pb.Transaction{tx}
	blk := blocktest.NewBlock(t, expected)

	err := c.state.ProcessBlock(blk)
	require.NoError(t, err, "c.state.ProcessBlock(blk)")

	chain.EXPECT().GetBlockByHash(gomock.Any()).Return(blk, nil).Times(2)

	actual, err := c.GetAccountTransactions(tx.From)
	require.NoError(t, err, "c.GetAccountTransactions(tx.From)")
	assert.Equal(t, expected, actual, "c.GetAccountTransactions(tx.From)")

	actual, err = c.GetAccountTransactions(tx.To)
	require.NoError(t, err, "c.GetAccountTransactions(tx.To)")
	assert.Equal(t, expected, actual, "c.GetAccountTransactions(tx.To)")

	actual, err = c.GetAccountTransactions(alice)
	require.NoError(t, err, "c.GetAccountTransactions(alice)")
	assert.Empty(t, actual, "c.GetAccountTransactions(alice)")
}

func TestCoinMining_SingleNode(t *testing.T) {
	minerPubKey := NewCoinBuilder(t).PublicKey()
	minerAddress, err := minerPubKey.Bytes()
	require.NoError(t, err, "minerPubKey.Bytes()")

	t.Run("start-stop-mining", func(t *testing.T) {
		c := NewCoinBuilder(t).WithPublicKey(minerPubKey).Build(t)
		verifyAccount(t, c, minerAddress, &pb.Account{})

		ctx, cancel := context.WithCancel(context.Background())
		errChan := make(chan error)
		go func() {
			errChan <- c.StartMining(ctx)
		}()

		cancel()
		err = <-errChan
		assert.EqualError(t, err, context.Canceled.Error(), "<-errChan")

		verifyAccount(t, c, minerAddress, &pb.Account{})
	})

	t.Run("reject-invalid-txs", func(t *testing.T) {
		c := NewCoinBuilder(t).Build(t)
		err := c.AddTransaction(ctestutil.NewTransaction(t, 0, 1, 1))
		assert.EqualError(t, err, validator.ErrInvalidTxValue.Error(), "c.AddTransaction()")

		err = c.AddTransaction(ctestutil.NewTransaction(t, 42, 1, 3))
		assert.EqualError(t, err, validator.ErrInsufficientBalance.Error(), "c.AddTransaction()")
	})

	t.Run("produce-blocks-from-old-txs", func(t *testing.T) {
		c := mineAllTransactions(
			t,
			&MiningTestConfig{
				initialBalance: 80,
				maxTxPerBlock:  2,
				minerPubKey:    minerPubKey,
				reward:         3,
				pendingTxs: []*pb.Transaction{
					ctestutil.NewTransaction(t, 20, 5, 3),
					ctestutil.NewTransaction(t, 10, 4, 4),
					ctestutil.NewTransaction(t, 5, 3, 5),
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
			verifyAccount(t, c, minerAddress, &pb.Account{Balance: 2*3 + 5 + 4 + 3})
		})

		t.Run("updates-sender-account", func(t *testing.T) {
			verifyAccount(t, c, []byte(ctestutil.TxSenderPID),
				&pb.Account{Balance: 80 - 20 - 5 - 10 - 4 - 5 - 3, Nonce: 5})
		})

		t.Run("updates-receiver-account", func(t *testing.T) {
			verifyAccount(t, c, []byte(ctestutil.TxRecipientPID),
				&pb.Account{Balance: 20 + 10 + 5})
		})
	})

	t.Run("produce-blocks-live-txs", func(t *testing.T) {
		c := mineAllTransactions(
			t,
			&MiningTestConfig{
				initialBalance: 200,
				maxTxPerBlock:  3,
				minerPubKey:    minerPubKey,
				reward:         5,
				pendingTxs: []*pb.Transaction{
					// These three txs should be included in the same block,
					// so their order (nonce) should not matter.
					ctestutil.NewTransaction(t, 5, 2, 1),
					ctestutil.NewTransaction(t, 10, 3, 3),
					ctestutil.NewTransaction(t, 7, 10, 4),
				},
				liveTxs: []*pb.Transaction{
					// This tx should be rejected because of its nonce and low fee.
					ctestutil.NewTransaction(t, 5, 1, 3),
					ctestutil.NewTransaction(t, 5, 2, 6),
					ctestutil.NewTransaction(t, 5, 1, 7),
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
			verifyAccount(t, c, minerAddress, &pb.Account{Balance: 2*5 + 2 + 3 + 10 + 2 + 1})
		})

		t.Run("updates-sender-account", func(t *testing.T) {
			verifyAccount(t, c, []byte(ctestutil.TxSenderPID),
				&pb.Account{Balance: 200 - 5 - 2 - 10 - 3 - 7 - 10 - 5 - 2 - 5 - 1, Nonce: 7})
		})

		t.Run("updates-receiver-account", func(t *testing.T) {
			verifyAccount(t, c, []byte(ctestutil.TxRecipientPID),
				&pb.Account{Balance: 5 + 10 + 7 + 5 + 5})
		})
	})
}

// MiningTestConfig contains the configuration of a coin miner run.
// For simplicity we only have one sender and one recipient (since
// the miner doesn't care about the transaction addresses, we don't
// lose test coverage with that assumpion).
type MiningTestConfig struct {
	initialBalance uint64              // Initial balance of the sender's account.
	maxTxPerBlock  uint32              // Max number of txs to include per block.
	minerPubKey    *coinutil.PublicKey // Public key of the miner.
	reward         uint64              // Miner reward for producing blocks.

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
		[]byte(ctestutil.TxSenderPID),
		&pb.Account{Balance: config.initialBalance},
	)
	assert.NoError(t, err, "s.UpdateAccount()")

	genesisBlock := blocktest.NewBlock(
		t,
		[]*pb.Transaction{&pb.Transaction{
			Value: config.initialBalance,
			To:    []byte(ctestutil.TxSenderPID),
		},
		})
	genesisBlock.Header.BlockNumber = 0
	chain := chain.NewChainDB(db, chain.OptGenesisBlock(genesisBlock))

	c := NewCoinBuilder(t).
		WithChain(chain).
		WithMaxTxPerBlock(config.maxTxPerBlock).
		WithPublicKey(config.minerPubKey).
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

func verifyAccount(t *testing.T, c *Coin, address []byte, expected *pb.Account) {
	account, err := c.GetAccount(address)
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
