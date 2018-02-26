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
	"errors"
	"testing"

	"github.com/stratumn/alice/core/protocol/coin/synchronizer/mocksynchronizer"

	"github.com/stratumn/alice/core/protocol/coin/engine/mockengine"
	"github.com/stratumn/alice/core/protocol/coin/validator/mockvalidator"

	"github.com/golang/mock/gomock"
	"github.com/stratumn/alice/core/p2p"
	"github.com/stratumn/alice/core/protocol/coin/chain"
	"github.com/stratumn/alice/core/protocol/coin/chain/mockchain"
	"github.com/stratumn/alice/core/protocol/coin/coinutil"
	"github.com/stratumn/alice/core/protocol/coin/db"
	mockp2p "github.com/stratumn/alice/core/protocol/coin/p2p/mockp2p"
	"github.com/stratumn/alice/core/protocol/coin/processor/mockprocessor"
	"github.com/stratumn/alice/core/protocol/coin/state"
	ctestutil "github.com/stratumn/alice/core/protocol/coin/testutil"
	tassert "github.com/stratumn/alice/core/protocol/coin/testutil/assert"
	"github.com/stratumn/alice/core/protocol/coin/testutil/blocktest"
	txtest "github.com/stratumn/alice/core/protocol/coin/testutil/transaction"
	"github.com/stratumn/alice/core/protocol/coin/validator"
	pb "github.com/stratumn/alice/pb/coin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	inet "gx/ipfs/QmQm7WmgYCa4RSz76tKEYpRjApjnRw8ZTUVQC15b8JM4a2/go-libp2p-net"
	protobuf "gx/ipfs/QmRDePEiL4Yupq5EkcK3L3ko3iMgYaqUdLu7xc1kqs7dnV/go-multicodec/protobuf"
	testutil "gx/ipfs/QmV1axkk86DDkYwS269AvPy9eV5h7mUyHveJkSVHPjrQtY/go-libp2p-netutil"
	peer "gx/ipfs/Qma7H6RW8wRrfZpNSXwxYGcd1E149s42FpWNpDNieSVrnU/go-libp2p-peer"
	ihost "gx/ipfs/QmfCtHMCd9xFvehvHeVxtKVXJTMVTuHhyPRVHEXetn87vL/go-libp2p-host"
)

func TestCoinGenesis(t *testing.T) {
	// Check genesis block.
	genBlock, genHash, err := GetGenesisBlock()
	assert.NoError(t, err, "GetGenesisBlock()")
	h, err := coinutil.HashHeader(genBlock.Header)
	assert.NoError(t, err, "HashHeader()")
	assert.Equal(t, h, genHash, "GenesisHash")

	// Run coin.
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	p := mockprocessor.NewMockProcessor(ctrl)
	coin := Coin{
		chain:     &ctestutil.SimpleChain{},
		processor: p,
		gossip:    ctestutil.NewDummyGossip(t),
	}

	p.EXPECT().Process(gomock.Any(), genBlock, gomock.Any(), gomock.Any()).Return(nil).Times(1)

	err = coin.Run(context.Background())
	assert.NoError(t, err, "coin.Run()")
}

func TestCoinProtocolHandler(t *testing.T) {
	ctx := context.Background()
	hosts := make([]ihost.Host, 2)
	coins := make([]*Coin, 2)
	p2ps := make([]*mockp2p.MockP2P, 2)

	for i := 0; i < 2; i++ {
		hosts[i] = p2p.NewHost(ctx, testutil.GenSwarmNetwork(t, ctx))
		defer func(h ihost.Host) {
			h.Close()
		}(hosts[i])

		// We configure validation to fail.
		// We only want to test that the coin protocol
		// correctly called the validator.
		coins[i] = &Coin{
			validator: ctestutil.NewInstrumentedValidator(&ctestutil.Rejector{}),
		}

		ii := i
		hosts[i].SetStreamHandler(ProtocolID, func(stream inet.Stream) {
			coins[ii].StreamHandler(ctx, stream)
		})
	}

	require.NoError(t, hosts[0].Connect(ctx, hosts[1].Peerstore().PeerInfo(hosts[1].ID())), "Connect()")
	require.NoError(t, hosts[1].Connect(ctx, hosts[0].Peerstore().PeerInfo(hosts[0].ID())), "Connect()")

	t.Run("Send different messages on the same stream", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		p2ps[0] = mockp2p.NewMockP2P(ctrl)
		coins[0].p2p = p2ps[0]
		p2ps[1] = mockp2p.NewMockP2P(ctrl)
		coins[1].p2p = p2ps[1]

		s0_1, err := hosts[0].NewStream(ctx, hosts[1].ID(), ProtocolID)
		require.NoError(t, err, "NewStream()")
		defer s0_1.Close()

		s1_0, err := hosts[1].NewStream(ctx, hosts[0].ID(), ProtocolID)
		require.NoError(t, err, "NewStream()")
		defer s1_0.Close()

		done := []bool{false, false, false}
		p2ps[0].EXPECT().RespondBlockByHash(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Do(func(_, _, _, _ interface{}) { done[0] = true }).Return(nil).Times(1)
		p2ps[1].EXPECT().RespondHeadersByNumber(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Do(func(_, _, _, _ interface{}) { done[1] = true }).Return(nil).Times(1)
		p2ps[1].EXPECT().RespondBlockByHash(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Do(func(_, _, _, _ interface{}) { done[2] = true }).Return(nil).Times(1)

		enc0_1 := protobuf.Multicodec(nil).Encoder(s0_1)
		enc1_0 := protobuf.Multicodec(nil).Encoder(s1_0)

		req1 := pb.NewBlockRequest([]byte("plap"))
		err = enc0_1.Encode(req1)
		assert.NoError(t, err, "Encode()")

		req2 := pb.NewBlockRequest([]byte("zou"))
		err = enc1_0.Encode(req2)
		assert.NoError(t, err, "Encode()")

		req3 := pb.NewHeadersRequest(0, 42)
		err = enc0_1.Encode(req3)
		assert.NoError(t, err, "Encode()")

		tassert.WaitUntil(t, func() bool { return done[0] && done[1] && done[2] }, "Calls are not done.")

	})

	t.Run("Ignore invalid messages", func(t *testing.T) {
		coins[0].p2p = ctestutil.NewDummyP2P(t)
		coins[1].p2p = ctestutil.NewDummyP2P(t)

		s1_0, err := hosts[1].NewStream(ctx, hosts[0].ID(), ProtocolID)
		require.NoError(t, err, "NewStream()")
		defer s1_0.Close()

		enc1_0 := protobuf.Multicodec(nil).Encoder(s1_0)
		err = enc1_0.Encode(&pb.Transaction{Value: 42})
		assert.NoError(t, err, "Encode()")
	})

	t.Run("Open and close streams", func(t *testing.T) {
		coins[0].p2p = ctestutil.NewDummyP2P(t)
		coins[1].p2p = ctestutil.NewDummyP2P(t)

		s1_0, err := hosts[1].NewStream(ctx, hosts[0].ID(), ProtocolID)
		require.NoError(t, err, "NewStream()")
		defer s1_0.Close()

		s0_1, err := hosts[0].NewStream(ctx, hosts[1].ID(), ProtocolID)
		require.NoError(t, err, "NewStream()")

		req1 := pb.NewBlockRequest([]byte("plap"))
		enc0_1 := protobuf.Multicodec(nil).Encoder(s0_1)
		err = enc0_1.Encode(req1)
		assert.NoError(t, err, "Encode()")

		err = s0_1.Close()
		assert.NoError(t, err, "Close()")

		req2 := pb.NewBlockRequest([]byte("zou"))
		enc1_0 := protobuf.Multicodec(nil).Encoder(s1_0)
		err = enc1_0.Encode(req2)
		assert.NoError(t, err, "Encode()")
	})
}

func TestGetAccountTransactions(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	chain := mockchain.NewMockChain(ctrl)

	alice := []byte("alice")

	c := NewCoinBuilder(t).WithChain(chain).Build(t)
	err := c.state.UpdateAccount([]byte(txtest.TxSenderPID), &pb.Account{Balance: 100})
	require.NoError(t, err)
	tx := txtest.NewTransaction(t, 1, 1, 1)
	expected := []*pb.Transaction{tx}
	blk := blocktest.NewBlock(t, expected)

	err = c.state.ProcessBlock(blk)
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

func TestAppendBlock(t *testing.T) {
	genBlock, _, err := GetGenesisBlock()
	assert.NoError(t, err, "GetGenesisBlock()")
	ch := &ctestutil.SimpleChain{}
	ch.AddBlock(genBlock)
	ch.SetHead(genBlock)

	// Run coin.
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	proc := mockprocessor.NewMockProcessor(ctrl)
	val := mockvalidator.NewMockValidator(ctrl)
	eng := mockengine.NewMockEngine(ctrl)
	coin := &Coin{
		chain:     ch,
		state:     ctestutil.NewSimpleState(t),
		gossip:    ctestutil.NewDummyGossip(t),
		processor: proc,
		validator: val,
		engine:    eng,
	}

	err = coin.Run(context.Background())
	assert.NoError(t, err, "coin.Run()")

	block := &pb.Block{Header: &pb.Header{PreviousHash: []byte("plap")}}

	val.EXPECT().ValidateBlock(block, gomock.Any()).Return(nil).Times(1)
	proc.EXPECT().Process(gomock.Any(), block, gomock.Any(), gomock.Any()).Return(nil).Times(1)
	eng.EXPECT().VerifyHeader(gomock.Any(), block.Header).Return(nil).Times(1)

	err = coin.AppendBlock(context.Background(), block)
	assert.NoError(t, err, "AppendBlock()")
}

func TestSynchronize(t *testing.T) {
	genBlock, _, err := GetGenesisBlock()
	assert.NoError(t, err, "GetGenesisBlock()")
	ch := &ctestutil.SimpleChain{}
	ch.AddBlock(genBlock)
	ch.SetHead(genBlock)

	// Run coin.
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	proc := mockprocessor.NewMockProcessor(ctrl)
	val := mockvalidator.NewMockValidator(ctrl)
	eng := mockengine.NewMockEngine(ctrl)
	sync := mocksynchronizer.NewMockSynchronizer(ctrl)

	tests := map[string]func(*testing.T, *Coin){
		"append-blocks-in-order": func(t *testing.T, coin *Coin) {
			block1 := &pb.Block{Header: &pb.Header{PreviousHash: []byte("plap")}}
			block2 := &pb.Block{Header: &pb.Header{PreviousHash: []byte("zou")}}

			hash := []byte("plap")
			resCh := make(chan *pb.Block)
			errCh := make(chan error)

			gomock.InOrder(
				sync.EXPECT().Synchronize(gomock.Any(), hash, coin.chain).Return(resCh, errCh).Times(1),
				val.EXPECT().ValidateBlock(block1, gomock.Any()).Return(nil).Times(1),
				eng.EXPECT().VerifyHeader(gomock.Any(), block1.Header).Return(nil).Times(1),
				proc.EXPECT().Process(gomock.Any(), block1, gomock.Any(), gomock.Any()).Return(nil).Times(1),
				val.EXPECT().ValidateBlock(block2, gomock.Any()).Return(nil).Times(1),
				eng.EXPECT().VerifyHeader(gomock.Any(), block2.Header).Return(nil).Times(1),
				proc.EXPECT().Process(gomock.Any(), block2, gomock.Any(), gomock.Any()).Return(nil).Times(1),
			)

			go func() {
				resCh <- block1
				resCh <- block2
				close(resCh)
			}()

			err = coin.synchronize(context.Background(), hash)
			assert.NoError(t, err, "synchronize()")
		},
		"sync-fail": func(t *testing.T, coin *Coin) {
			block := &pb.Block{Header: &pb.Header{PreviousHash: []byte("plap")}}

			hash := []byte("plap")
			resCh := make(chan *pb.Block)
			errCh := make(chan error)

			gomock.InOrder(
				sync.EXPECT().Synchronize(gomock.Any(), hash, coin.chain).Return(resCh, errCh).Times(1),
				val.EXPECT().ValidateBlock(block, gomock.Any()).Return(nil).Times(1),
				eng.EXPECT().VerifyHeader(gomock.Any(), block.Header).Return(nil).Times(1),
				proc.EXPECT().Process(gomock.Any(), block, gomock.Any(), gomock.Any()).Return(nil).Times(1),
			)

			gandalf := errors.New("you shall not pass")

			go func() {
				resCh <- block
				errCh <- gandalf
			}()

			err = coin.synchronize(context.Background(), hash)
			assert.EqualError(t, err, gandalf.Error(), "synchronize()")
		},
	}

	for n, test := range tests {
		t.Run(n, func(t *testing.T) {
			coin := &Coin{
				chain:        ch,
				state:        ctestutil.NewSimpleState(t),
				gossip:       ctestutil.NewDummyGossip(t),
				processor:    proc,
				validator:    val,
				engine:       eng,
				synchronizer: sync,
			}

			err = coin.Run(context.Background())
			assert.NoError(t, err, "coin.Run()")

			test(t, coin)
		})
	}
}

func TestGetBlockchain(t *testing.T) {
	ch := &ctestutil.SimpleChain{}

	block0 := blocktest.NewBlock(t, txtest.NewTransactions(t, 2))
	block0.Header.BlockNumber = 0
	hash0, err := coinutil.HashHeader(block0.Header)
	assert.NoError(t, err, "coinutil.HashHeader()")

	block1 := blocktest.NewBlock(t, txtest.NewTransactions(t, 3))
	block1.Header.BlockNumber = 1
	block1.Header.PreviousHash = hash0
	hash1, err := coinutil.HashHeader(block1.Header)
	assert.NoError(t, err, "coinutil.HashHeader()")

	block2 := blocktest.NewBlock(t, txtest.NewTransactions(t, 1))
	block2.Header.BlockNumber = 2
	block2.Header.PreviousHash = hash1
	hash2, err := coinutil.HashHeader(block2.Header)
	assert.NoError(t, err, "coinutil.HashHeader()")

	require.NoError(t, ch.AddBlock(block0), "ch.AddBlock()")
	require.NoError(t, ch.AddBlock(block1), "ch.AddBlock()")
	require.NoError(t, ch.AddBlock(block2), "ch.AddBlock()")
	require.NoError(t, ch.SetHead(block2), "ch.SetHead()")

	coin := NewCoinBuilder(t).WithChain(ch).Build(t)

	t.Run("block-number-single", func(t *testing.T) {
		blocks, err := coin.GetBlockchain(2, nil, 0)
		require.NoError(t, err, "coin.GetBlockchain()")
		require.Len(t, blocks, 1, "len(blocks)")
		assert.Equal(t, block2, blocks[0])
	})

	t.Run("block-number-multiple", func(t *testing.T) {
		blocks, err := coin.GetBlockchain(1, hash2, 5)
		require.NoError(t, err, "coin.GetBlockchain()")
		require.Len(t, blocks, 2, "len(blocks)")
		assert.Equal(t, block1, blocks[0])
		assert.Equal(t, block0, blocks[1])
	})

	t.Run("header-hash-multiple", func(t *testing.T) {
		blocks, err := coin.GetBlockchain(0, hash2, 2)
		require.NoError(t, err, "coin.GetBlockchain()")
		require.Len(t, blocks, 2, "len(blocks)")
		assert.Equal(t, block2, blocks[0])
		assert.Equal(t, block1, blocks[1])
	})

	t.Run("head", func(t *testing.T) {
		blocks, err := coin.GetBlockchain(0, nil, 1)
		require.NoError(t, err, "coin.GetBlockchain()")
		require.Len(t, blocks, 1, "len(blocks)")
		assert.Equal(t, block2, blocks[0])
	})
}

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
