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
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	floodsub "gx/ipfs/QmSjoxpBJV71bpSojnUY1K382Ly3Up55EspnDx6EKAmQX4/go-libp2p-floodsub"
	netutil "gx/ipfs/QmWUugnJBbcuin8qdfiCYKAsNkG8NeDLhzoBqRaqXhAHd4/go-libp2p-netutil"
	bhost "gx/ipfs/QmZ15dDSCo4DKn4o4GnqqLExKATBeeo3oNyQ5FBKtNjEQT/go-libp2p-blankhost"
	host "gx/ipfs/QmfCtHMCd9xFvehvHeVxtKVXJTMVTuHhyPRVHEXetn87vL/go-libp2p-host"

	"github.com/stratumn/alice/core/protocol/coin/db"
	"github.com/stratumn/alice/core/protocol/coin/state"
	tassert "github.com/stratumn/alice/core/protocol/coin/testutil/assert"
	"github.com/stratumn/alice/core/protocol/coin/validator"
	"github.com/stratumn/alice/core/protocol/coin/validator/mockvalidator"
	pb "github.com/stratumn/alice/pb/coin"
)

var (
	mockValidator *mockvalidator.MockValidator
)

type msg struct {
	topic string
	data  interface{}
}

func newHost(ctx context.Context, t *testing.T) host.Host {
	ntw := netutil.GenSwarmNetwork(t, ctx)
	return bhost.NewBlankHost(ntw)
}

func newGossip(ctx context.Context, t *testing.T, h host.Host) (Gossip, error) {
	p, err := floodsub.NewFloodSub(ctx, h)
	if err != nil {
		return nil, err
	}

	db, err := db.NewMemDB(nil)
	if err != nil {
		return nil, err
	}

	s := state.NewState(db)

	return NewGossip(h, p, s, mockValidator), nil
}

func TestGossip(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockValidator = mockvalidator.NewMockValidator(mockCtrl)

	h1 := newHost(ctx, t)
	h2 := newHost(ctx, t)

	g1, err := newGossip(ctx, t, h1)
	require.NoError(t, err)

	g2, err := newGossip(ctx, t, h2)
	require.NoError(t, err)

	h2pi := h2.Peerstore().PeerInfo(h2.ID())
	require.NoError(t, h1.Connect(ctx, h2pi), "Connect()")

	c1 := make(chan msg)
	c2 := make(chan msg)

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
		assert.Equal(t, true, gg1.isSubscribed(topic))

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
		})
		assert.NoError(t, err)

		err = g2.ListenBlock(context.Background(), func(b *pb.Block) error {
			c2 <- msg{
				topic: BlockTopicName,
				data:  b,
			}
			return nil
		})
		assert.NoError(t, err)

		tassert.WaitUntil(t, func() bool {
			return len(g2.(*gossip).pubsub.ListPeers(BlockTopicName)) == 1
		}, "ListPeers(topic) should have length of 1")
	})

	t.Run("PublishTx", func(t *testing.T) {
		tx := &pb.Transaction{}
		mockValidator.EXPECT().ValidateTx(tx, gomock.Any()).Return(nil).Times(2)

		err = g1.PublishTx(tx)

		assert.NoError(t, err)

		assertNotReceive(t, c1)
		assertReceive(t, c2, tx)
	})

	t.Run("PublishInvalidTx", func(t *testing.T) {
		tx := &pb.Transaction{}
		mockValidator.EXPECT().ValidateTx(tx, gomock.Any()).Return(validator.ErrEmptyTx)

		err = g1.PublishTx(tx)

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
		b := &pb.Block{}
		mockValidator.EXPECT().ValidateBlock(b, gomock.Any()).Return(nil).Times(2)

		err = g2.PublishBlock(b)

		assert.NoError(t, err)

		assertReceive(t, c1, b)
		assertNotReceive(t, c2)
	})

	t.Run("PublishInvalidBlock", func(t *testing.T) {
		b := &pb.Block{}
		mockValidator.EXPECT().ValidateBlock(b, gomock.Any()).Return(validator.ErrTooManyTxs)

		err = g1.PublishBlock(b)

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
