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

package protocol

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stratumn/go-indigonode/app/coin/pb"
	"github.com/stratumn/go-indigonode/app/coin/protocol/chain/mockchain"
	"github.com/stratumn/go-indigonode/app/coin/protocol/coinutil"
	"github.com/stratumn/go-indigonode/app/coin/protocol/engine/mockengine"
	mockp2p "github.com/stratumn/go-indigonode/app/coin/protocol/p2p/mockp2p"
	"github.com/stratumn/go-indigonode/app/coin/protocol/processor/mockprocessor"
	ctestutil "github.com/stratumn/go-indigonode/app/coin/protocol/testutil"
	tassert "github.com/stratumn/go-indigonode/app/coin/protocol/testutil/assert"
	"github.com/stratumn/go-indigonode/app/coin/protocol/testutil/blocktest"
	txtest "github.com/stratumn/go-indigonode/app/coin/protocol/testutil/transaction"
	"github.com/stratumn/go-indigonode/app/coin/protocol/validator/mockvalidator"
	"github.com/stratumn/go-indigonode/core/p2p"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	protobuf "gx/ipfs/QmRDePEiL4Yupq5EkcK3L3ko3iMgYaqUdLu7xc1kqs7dnV/go-multicodec/protobuf"
	inet "gx/ipfs/QmXoz9o2PT3tEzf7hicegwex5UgVP54n3k82K7jrWFyN86/go-libp2p-net"
	testutil "gx/ipfs/Qmb6BsZf6Y3kxffXMNTubGPF1w1bkHtpvhfYbmnwP3NQyw/go-libp2p-netutil"
	ihost "gx/ipfs/QmfZTdmunzKzAGJrSvXXQbQ5kLLUiEMX5vdwux7iXkdk7D/go-libp2p-host"
)

func TestCoinGenesis(t *testing.T) {
	t.Run("set-genesis-block", func(t *testing.T) {
		// Check genesis block.
		genBlock := &pb.Block{Header: &pb.Header{Nonce: 42}}

		// Run coin.
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		p := mockprocessor.NewMockProcessor(ctrl)
		coin := Coin{
			genesisBlock: genBlock,
			chain:        &ctestutil.SimpleChain{},
			processor:    p,
			gossip:       ctestutil.NewDummyGossip(t),
		}

		p.EXPECT().Process(gomock.Any(), genBlock, gomock.Any(), gomock.Any()).Return(nil).Times(1)

		err := coin.Run(context.Background())
		assert.NoError(t, err, "coin.Run()")
	})
	t.Run("second-genesis-block", func(t *testing.T) {
		// Check genesis block.
		genBlock := &pb.Block{Header: &pb.Header{Nonce: 42}}
		genBlockBis := &pb.Block{Header: &pb.Header{Nonce: 43}}

		// Run coin.
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		p := mockprocessor.NewMockProcessor(ctrl)
		c := &ctestutil.SimpleChain{}
		c.AddBlock(genBlock)
		c.SetHead(genBlock)

		coin := Coin{
			genesisBlock: genBlockBis,
			chain:        c,
			processor:    p,
			gossip:       ctestutil.NewDummyGossip(t),
		}

		err := coin.Run(context.Background())
		assert.NoError(t, err, "coin.Run()")
	})
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
	genBlock := &pb.Block{Header: &pb.Header{Nonce: 42, Version: 32}}
	ch := &ctestutil.SimpleChain{}
	ch.AddBlock(CoinBuilderDefaultGen)
	ch.SetHead(CoinBuilderDefaultGen)

	// Run coin.
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	proc := mockprocessor.NewMockProcessor(ctrl)
	val := mockvalidator.NewMockValidator(ctrl)
	eng := mockengine.NewMockPoW(ctrl)
	coin := NewCoinBuilder(t).
		WithChain(ch).
		WithProcessor(proc).
		WithValidator(val).
		WithEngine(eng).
		WithGenesisBlock(genBlock).
		Build(t)

	err := coin.Run(context.Background())
	assert.NoError(t, err, "coin.Run()")

	block := &pb.Block{Header: &pb.Header{PreviousHash: []byte("plap")}}

	val.EXPECT().ValidateBlock(block, gomock.Any()).Return(nil).Times(1)
	proc.EXPECT().Process(gomock.Any(), block, gomock.Any(), gomock.Any()).Return(nil).Times(1)
	eng.EXPECT().VerifyHeader(gomock.Any(), block.Header).Return(nil).Times(1)

	err = coin.AppendBlock(context.Background(), block)
	assert.NoError(t, err, "AppendBlock()")
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

func TestGetTransactionPool(t *testing.T) {
	pool := &ctestutil.InMemoryTxPool{}
	coin := NewCoinBuilder(t).WithTxPool(pool).Build(t)

	alice := []byte("Alice")
	bob := []byte("Bob")

	pool.AddTransaction(&pb.Transaction{Value: 42, From: alice, To: bob})
	pool.AddTransaction(&pb.Transaction{Value: 41, From: bob, To: alice})
	pool.AddTransaction(&pb.Transaction{Value: 43, From: alice, To: alice})
	pool.AddTransaction(&pb.Transaction{Value: 45, From: alice, To: bob})

	t.Run("missing-count", func(t *testing.T) {
		txCount, txs, err := coin.GetTransactionPool(0)
		assert.NoError(t, err, "coin.GetTransactionPool()")
		assert.Equal(t, uint64(4), txCount, "txCount")
		assert.Len(t, txs, 1)
	})

	t.Run("huge-count", func(t *testing.T) {
		txCount, txs, err := coin.GetTransactionPool(42000)
		assert.NoError(t, err, "coin.GetTransactionPool()")
		assert.Equal(t, uint64(4), txCount, "txCount")
		assert.Len(t, txs, 4)
	})
}
