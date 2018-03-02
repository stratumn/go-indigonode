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

package gossip

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stratumn/alice/core/protocol/coin/chain"
	"github.com/stratumn/alice/core/protocol/coin/coinutil"
	"github.com/stratumn/alice/core/protocol/coin/db"
	"github.com/stratumn/alice/core/protocol/coin/engine/mockengine"
	"github.com/stratumn/alice/core/protocol/coin/p2p/mockp2p"
	"github.com/stratumn/alice/core/protocol/coin/processor"
	"github.com/stratumn/alice/core/protocol/coin/state"
	"github.com/stratumn/alice/core/protocol/coin/synchronizer"
	"github.com/stratumn/alice/core/protocol/coin/synchronizer/mocksynchronizer"
	tassert "github.com/stratumn/alice/core/protocol/coin/testutil/assert"
	"github.com/stratumn/alice/core/protocol/coin/testutil/blocktest"
	txtest "github.com/stratumn/alice/core/protocol/coin/testutil/transaction"
	"github.com/stratumn/alice/core/protocol/coin/validator"
	"github.com/stratumn/alice/core/protocol/coin/validator/mockvalidator"
	pb "github.com/stratumn/alice/pb/coin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	floodsub "gx/ipfs/QmSjoxpBJV71bpSojnUY1K382Ly3Up55EspnDx6EKAmQX4/go-libp2p-floodsub"
	netutil "gx/ipfs/QmWUugnJBbcuin8qdfiCYKAsNkG8NeDLhzoBqRaqXhAHd4/go-libp2p-netutil"
	bhost "gx/ipfs/QmZ15dDSCo4DKn4o4GnqqLExKATBeeo3oNyQ5FBKtNjEQT/go-libp2p-blankhost"
	cid "gx/ipfs/QmcZfnkapfECQGcLZaf9B79NRg7cRa9EnZh4LSbkCzwNvY/go-cid"
	pstore "gx/ipfs/QmeZVQzUrXqaszo24DAoHfGzcmCptN9JyngLkGAiEfk2x7/go-libp2p-peerstore"
	host "gx/ipfs/QmfCtHMCd9xFvehvHeVxtKVXJTMVTuHhyPRVHEXetn87vL/go-libp2p-host"
)

type msg struct {
	topic string
	data  interface{}
}

func newHost(ctx context.Context, t *testing.T) host.Host {
	ntw := netutil.GenSwarmNetwork(t, ctx)
	return bhost.NewBlankHost(ntw)
}

type GossipBuilder struct {
	h  host.Host
	db db.DB

	chain     chain.Chain
	pubsub    *floodsub.PubSub
	state     state.State
	validator validator.Validator
}

func NewGossipBuilder(h host.Host) *GossipBuilder {
	return &GossipBuilder{h: h}
}

func (g *GossipBuilder) WithChain(c chain.Chain) *GossipBuilder {
	g.chain = c
	return g
}

func (g *GossipBuilder) WithState(s state.State) *GossipBuilder {
	g.state = s
	return g
}

func (g *GossipBuilder) WithValidator(v validator.Validator) *GossipBuilder {
	g.validator = v
	return g
}

func (g *GossipBuilder) Build(ctx context.Context, t *testing.T) Gossip {
	if g.pubsub == nil {
		p, err := floodsub.NewFloodSub(ctx, g.h)
		require.NoError(t, err, "floodsub.NewFloodSub()")
		g.pubsub = p
	}
	if g.db == nil {
		db, err := db.NewMemDB(nil)
		require.NoError(t, err, "db.NewMemDB()")
		g.db = db
	}
	if g.chain == nil {
		g.chain = chain.NewChainDB(g.db)
	}
	if g.state == nil {
		g.state = state.NewState(g.db)
	}
	if g.validator == nil {
		e := mockengine.NewMockPoW(gomock.NewController(t))
		e.EXPECT().VerifyHeader(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
		g.validator = validator.NewGossipValidator(2, e, g.chain)
	}

	return NewGossip(g.h, g.pubsub, g.state, g.chain, g.validator)
}

func TestGossip(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	v := mockvalidator.NewMockValidator(mockCtrl)

	h1 := newHost(ctx, t)
	h2 := newHost(ctx, t)

	g1 := NewGossipBuilder(h1).WithValidator(v).Build(ctx, t)
	g2 := NewGossipBuilder(h2).WithValidator(v).Build(ctx, t)

	h2pi := h2.Peerstore().PeerInfo(h2.ID())
	require.NoError(t, h1.Connect(ctx, h2pi), "Connect()")

	c1 := make(chan msg)
	c2 := make(chan msg)

	hashCh1 := make(chan []byte)
	hashCh2 := make(chan []byte)

	t.Run("SubscribeListen", func(t *testing.T) {
		topic := "coin.test"

		gg1 := g1.(*gossip)
		gg2 := g2.(*gossip)

		ctx := context.Background()

		// listen before subscribe
		err := gg1.listen(ctx, topic, func(msg *floodsub.Message) error { return nil })
		assert.Error(t, err)

		err = gg1.subscribe(topic, func(ctx context.Context, m *floodsub.Message) bool { return true })
		require.NoError(t, err)
		assert.Len(t, gg1.pubsub.GetTopics(), 1)
		assert.True(t, gg1.isSubscribed(topic))

		err = gg1.listen(ctx, topic, func(msg *floodsub.Message) error { return nil })
		require.NoError(t, err)

		err = gg2.subscribe(topic, func(ctx context.Context, m *floodsub.Message) bool { return true })
		require.NoError(t, err)

		err = gg2.listen(ctx, topic, func(msg *floodsub.Message) error { return nil })
		require.NoError(t, err)

		tassert.WaitUntil(t, func() bool {
			return len(gg2.pubsub.ListPeers(topic)) == 1
		}, "ListPeers(topic) should have length of 1")
		assert.Contains(t, gg2.pubsub.ListPeers(topic), gg1.host.ID())
		assert.Len(t, gg1.pubsub.ListPeers(topic), 1)
		assert.Contains(t, gg1.pubsub.ListPeers(topic), gg2.host.ID())
	})

	t.Run("ListenTx", func(t *testing.T) {
		err := g1.ListenTx(context.Background(), func(tx *pb.Transaction) error {
			c1 <- msg{
				topic: TxTopicName,
				data:  tx,
			}
			return nil
		})
		assert.NoError(t, err)

		err = g2.ListenTx(context.Background(), func(tx *pb.Transaction) error {
			c2 <- msg{
				topic: TxTopicName,
				data:  tx,
			}
			return nil
		})
		assert.NoError(t, err)

		tassert.WaitUntil(t, func() bool {
			return len(g2.(*gossip).pubsub.ListPeers(TxTopicName)) == 1
		}, "ListPeers(topic) should have length of 1")
	})

	t.Run("ListenBlock", func(t *testing.T) {
		err := g1.ListenBlock(context.Background(), func(b *pb.Block) error {
			c1 <- msg{
				topic: BlockTopicName,
				data:  b,
			}
			return nil
		}, func(h []byte) error {
			go func() { hashCh1 <- h }()
			return nil
		})
		assert.NoError(t, err)

		err = g2.ListenBlock(context.Background(), func(b *pb.Block) error {
			c2 <- msg{
				topic: BlockTopicName,
				data:  b,
			}
			return nil
		}, func(h []byte) error {
			go func() { hashCh2 <- h }()
			return nil
		})
		assert.NoError(t, err)

		tassert.WaitUntil(t, func() bool {
			return len(g2.(*gossip).pubsub.ListPeers(BlockTopicName)) == 1
		}, "ListPeers(topic) should have length of 1")
	})

	t.Run("PublishTx", func(t *testing.T) {
		tx := &pb.Transaction{}
		v.EXPECT().ValidateTx(tx, gomock.Any()).Return(nil).Times(2)

		err := g1.PublishTx(tx)
		assert.NoError(t, err)

		assertNotReceive(t, c1)
		assertReceive(t, c2, tx)
	})

	t.Run("PublishInvalidTx", func(t *testing.T) {
		tx := &pb.Transaction{}
		v.EXPECT().ValidateTx(tx, gomock.Any()).Return(validator.ErrEmptyTx)

		err := g1.PublishTx(tx)
		assert.NoError(t, err)

		assertNotReceive(t, c1)
		assertNotReceive(t, c2)
	})

	t.Run("PublishMalformedTx", func(t *testing.T) {
		tx := struct {
			From []byte
		}{
			From: []byte("from"),
		}

		txBytes, _ := json.Marshal(tx)

		g := g1.(*gossip)
		err := g.pubsub.Publish(TxTopicName, txBytes)
		assert.NoError(t, err)

		assertNotReceive(t, c1)
		assertNotReceive(t, c2)
	})

	t.Run("PublishBlock", func(t *testing.T) {
		b := &pb.Block{Header: &pb.Header{PreviousHash: []byte("prev")}}
		v.EXPECT().ValidateBlock(b, gomock.Any()).Return(nil).Times(2)

		err := g2.PublishBlock(b)
		assert.NoError(t, err)

		assertReceive(t, c1, b)
		assertNotReceive(t, c2)

		// Check that we asked for the sync.
		tassert.WaitUntil(t, func() bool {
			h := <-hashCh1
			return bytes.Equal(h, []byte("prev"))
		}, "zou")
	})

	t.Run("PublishInvalidBlock", func(t *testing.T) {
		b := &pb.Block{Header: &pb.Header{PreviousHash: []byte("prev")}}
		v.EXPECT().ValidateBlock(b, gomock.Any()).Return(validator.ErrTooManyTxs)

		err := g1.PublishBlock(b)
		assert.NoError(t, err)

		assertNotReceive(t, c1)
		assertNotReceive(t, c2)
	})

	t.Run("PublishMalformedBlock", func(t *testing.T) {
		b := struct {
			Header []byte
		}{
			Header: []byte("header"),
		}

		bBytes, _ := json.Marshal(b)

		g := g1.(*gossip)
		err := g.pubsub.Publish(BlockTopicName, bBytes)
		assert.NoError(t, err)

		assertNotReceive(t, c1)
		assertNotReceive(t, c2)
	})

	t.Run("NotifyBlockListeners", func(t *testing.T) {
		b := &pb.Block{Header: &pb.Header{BlockNumber: 42}}
		v.EXPECT().ValidateBlock(b, gomock.Any()).Return(nil).Times(2)

		l1 := g2.AddBlockListener()
		l2 := g2.AddBlockListener()

		err := g1.PublishBlock(b)
		assert.NoError(t, err)

		assertReceive(t, c2, b)
		assertNotReceive(t, c1)

		select {
		case h1 := <-l1:
			assert.Equal(t, b.Header, h1, "<-l1")
		case <-time.After(10 * time.Millisecond):
			assert.Fail(t, "<-l1 timed out")
		}

		select {
		case h2 := <-l2:
			assert.Equal(t, b.Header, h2, "<-l2")
		case <-time.After(10 * time.Millisecond):
			assert.Fail(t, "<-l2 timed out")
		}
	})

	t.Run("Close", func(t *testing.T) {
		g := NewGossipBuilder(h1).Build(ctx, t)

		gg := g.(*gossip)
		err := gg.subscribe("topic1", func(ctx context.Context, m *floodsub.Message) bool { return true })
		require.NoError(t, err)

		err = gg.subscribe("topic2", func(ctx context.Context, m *floodsub.Message) bool { return true })
		require.NoError(t, err)

		assert.Len(t, gg.pubsub.GetTopics(), 2)

		g.AddBlockListener()
		assert.Len(t, gg.blockListeners, 1)

		assert.NoError(t, gg.Close(), "Close()")
		assert.Len(t, gg.pubsub.GetTopics(), 0)
		assert.Len(t, gg.blockListeners, 0)
	})
}

func TestGossipWithSync(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	/***** Setup two nodes with diverging chains *****/

	// block0 --- block1
	//        \-- block1bis --- block2bis

	block0 := blocktest.NewBlock(t, []*pb.Transaction{
		&pb.Transaction{To: []byte(txtest.TxSenderPID), Value: 42},
	})
	block0.Header.BlockNumber = 0
	block0Hash, err := coinutil.HashHeader(block0.Header)
	require.NoError(t, err, "coinutil.HashHeader()")

	block1 := blocktest.NewBlock(t, []*pb.Transaction{txtest.NewTransaction(t, 7, 2, 2)})
	block1.Header.BlockNumber = 1
	block1.Header.PreviousHash = block0Hash

	block1bis := blocktest.NewBlock(t, []*pb.Transaction{txtest.NewTransaction(t, 5, 2, 2)})
	block1bis.Header.BlockNumber = 1
	block1bis.Header.PreviousHash = block0Hash
	block1bisHash, err := coinutil.HashHeader(block1bis.Header)
	require.NoError(t, err, "coinutil.HashHeader()")

	block2bis := blocktest.NewBlock(t, []*pb.Transaction{txtest.NewTransaction(t, 3, 2, 3)})
	block2bis.Header.BlockNumber = 2
	block2bis.Header.PreviousHash = block1bisHash

	p := processor.NewProcessor(nil)

	// Host1's chain contains block0 -> block1
	h1 := newHost(ctx, t)
	db1, err := db.NewMemDB(nil)
	require.NoError(t, err, "db.NewMemDB()")

	chain1 := chain.NewChainDB(db1)
	state1 := state.NewState(db1)
	require.NoError(t, p.Process(ctx, block0, state1, chain1), "p.Process()")
	require.NoError(t, p.Process(ctx, block1, state1, chain1), "p.Process()")

	g1 := NewGossipBuilder(h1).WithChain(chain1).WithState(state1).Build(ctx, t)

	// Host2's chain contains block0 -> block1bis -> block2bis
	h2 := newHost(ctx, t)
	db2, err := db.NewMemDB(nil)
	require.NoError(t, err, "db.NewMemDB()")

	chain2 := chain.NewChainDB(db2)
	state2 := state.NewState(db2)
	require.NoError(t, p.Process(ctx, block0, state2, chain2), "p.Process()")
	require.NoError(t, p.Process(ctx, block1bis, state2, chain2), "p.Process()")
	require.NoError(t, p.Process(ctx, block2bis, state2, chain2), "p.Process()")

	g2 := NewGossipBuilder(h2).WithChain(chain2).WithState(state2).Build(ctx, t)

	/***** Connect the two nodes *****/

	h2pi := h2.Peerstore().PeerInfo(h2.ID())
	require.NoError(t, h1.Connect(ctx, h2pi), "Connect()")

	/***** Set synchronize expectations *****/

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	p2p := mockp2p.NewMockP2P(ctrl)
	dht := mocksynchronizer.NewMockContentProviderFinder(ctrl)
	sync := synchronizer.NewSynchronizer(p2p, dht, synchronizer.OptMaxBatchSizes(2))

	// The sync should request the missing blocks from h2.
	cid1, _ := cid.Cast(block1bisHash)
	dht.EXPECT().
		FindProviders(gomock.Any(), cid1).
		Return([]pstore.PeerInfo{h2pi}, nil).
		Times(1)

	p2p.EXPECT().
		RequestBlockByHash(gomock.Any(), h2.ID(), []byte(block1bisHash)).
		Return(block1bis, nil).
		Times(1)

	p2p.EXPECT().
		RequestHeadersByNumber(gomock.Any(), h2.ID(), uint64(0), uint64(2)).
		Return([]*pb.Header{block0.Header, block1bis.Header}, nil).
		Times(1)

	p2p.EXPECT().
		RequestBlocksByNumber(gomock.Any(), h2.ID(), uint64(1), uint64(2)).
		Return([]*pb.Block{block1bis, block2bis}, nil).
		Times(1)

	p2p.EXPECT().
		RequestBlocksByNumber(gomock.Any(), h2.ID(), uint64(3), uint64(2)).
		Return(nil, nil).
		Times(1)

	/***** Listen to blocks and synchronize *****/

	err = g1.ListenBlock(ctx, func(b *pb.Block) error {
		return p.Process(ctx, b, state1, chain1)
	}, func(h []byte) error {
		resCh, errCh := sync.Synchronize(ctx, h, chain1)

		for {
			select {
			case b, ok := <-resCh:
				if !ok {
					// No more blocks to process.
					return nil
				}
				if err := p.Process(ctx, b, state1, chain1); err != nil {
					return err
				}
			case err := <-errCh:
				return err
			}
		}
	})
	assert.NoError(t, err, "g1.ListenBlock()")

	err = g2.ListenBlock(ctx, func(*pb.Block) error { return nil }, func([]byte) error { return nil })
	assert.NoError(t, err, "g2.ListenBlock()")

	tassert.WaitUntil(t, func() bool {
		return len(g1.(*gossip).pubsub.ListPeers(BlockTopicName)) == 1
	}, "ListPeers(topic) should have length of 1")

	err = g2.PublishBlock(block2bis)
	assert.NoError(t, err, "g2.PublishBlock()")

	tassert.WaitUntil(t, func() bool {
		b, err := chain1.CurrentHeader()
		assert.NoError(t, err, "chain.CurrentHeader()")
		return b.BlockNumber == 2
	}, "chain not synchronized")

	a1, err := state1.GetAccount([]byte(txtest.TxSenderPID))
	assert.NoError(t, err, "state1.GetAccount()")
	a2, err := state2.GetAccount([]byte(txtest.TxSenderPID))
	assert.NoError(t, err, "state2.GetAccount()")

	assert.Equal(t, uint64(30), a1.Balance, "a1.Balance")
	assert.Equal(t, uint64(30), a1.Balance, "a2.Balance")
	assert.Equal(t, uint64(3), a1.Nonce, "a1.Nonce")
	assert.Equal(t, uint64(3), a2.Nonce, "a2.Nonce")
}

func assertReceive(t *testing.T, c chan msg, want interface{}) {
	select {
	case m := <-c:
		switch m.topic {
		case TxTopicName:
			assert.Equal(t, want, m.data.(*pb.Transaction))
		case BlockTopicName:
			assert.Equal(t, want, m.data.(*pb.Block))
		}
	case <-time.After(time.Millisecond * 50):
		t.Fatalf("timed out waiting for transaction: %+v", want)
	}
}

func assertNotReceive(t *testing.T, c chan msg) {
	select {
	case m := <-c:
		t.Fatalf("received unexpected message: %+v", m)
	case <-time.After(time.Millisecond * 50):
		return
	}
}
