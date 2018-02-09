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
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	floodsub "gx/ipfs/QmSjoxpBJV71bpSojnUY1K382Ly3Up55EspnDx6EKAmQX4/go-libp2p-floodsub"
	netutil "gx/ipfs/QmWUugnJBbcuin8qdfiCYKAsNkG8NeDLhzoBqRaqXhAHd4/go-libp2p-netutil"
	bhost "gx/ipfs/QmZ15dDSCo4DKn4o4GnqqLExKATBeeo3oNyQ5FBKtNjEQT/go-libp2p-blankhost"
	host "gx/ipfs/QmfCtHMCd9xFvehvHeVxtKVXJTMVTuHhyPRVHEXetn87vL/go-libp2p-host"

	"github.com/stratumn/alice/core/protocol/coin/db"
	"github.com/stratumn/alice/core/protocol/coin/state"
	"github.com/stratumn/alice/core/protocol/coin/validator"
	pb "github.com/stratumn/alice/pb/coin"
)

var (
	g1, g2 *Gossip
)

type val struct{}

func (v *val) ValidateTx(tx *pb.Transaction, state state.Reader) error {
	if bytes.Equal(tx.GetFrom(), []byte("invalid")) {
		return validator.ErrInvalidTxSender
	}

	return nil
}

func (v *val) ValidateBlock(block *pb.Block, state state.Reader) error {
	return nil
}

func newHost(ctx context.Context, t *testing.T) host.Host {
	ntw := netutil.GenSwarmNetwork(t, ctx)
	return bhost.NewBlankHost(ntw)
}

func newGossip(ctx context.Context, t *testing.T, h host.Host) (*Gossip, error) {
	p, err := floodsub.NewFloodSub(ctx, h)
	if err != nil {
		return nil, err
	}

	db, err := db.NewMemDB(nil)
	if err != nil {
		return nil, err
	}

	s := state.NewState(db)

	return NewGossip(*p, s, &val{}), nil
}

func TestGossip(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	h1 := newHost(ctx, t)
	h2 := newHost(ctx, t)

	g1, err := newGossip(ctx, t, h1)
	require.NoError(t, err)

	g2, err := newGossip(ctx, t, h2)
	require.NoError(t, err)

	h2pi := h2.Peerstore().PeerInfo(h2.ID())
	require.NoError(t, h1.Connect(ctx, h2pi), "Connect()")

	c1 := make(chan *pb.Transaction)
	c2 := make(chan *pb.Transaction)

	t.Run("Subscribe", func(t *testing.T) {
		err = g1.SubscribeTx()
		assert.NoError(t, err)
		require.Len(t, g1.pubsub.GetTopics(), 1)
		assert.Contains(t, g1.pubsub.GetTopics(), TxTopicName)

		err = g2.SubscribeTx()
		assert.NoError(t, err)
		require.Len(t, g2.pubsub.GetTopics(), 1)
		assert.Contains(t, g2.pubsub.GetTopics(), TxTopicName)
	})

	t.Run("ListenTx", func(t *testing.T) {
		g1.ListenTx(context.Background(), func(tx *pb.Transaction) error {
			c1 <- tx
			return nil
		}, make(chan error))

		g2.ListenTx(context.Background(), func(tx *pb.Transaction) error {
			c2 <- tx
			return nil
		}, make(chan error))

		time.Sleep(1 * time.Second)
	})

	t.Run("PublishTx", func(t *testing.T) {
		tx := &pb.Transaction{
			From: []byte("valid"),
		}
		err = g1.PublishTx(tx)

		assert.NoError(t, err)

		assertReceive(t, c1, tx)
		assertReceive(t, c2, tx)
	})

	t.Run("PublishInvalidTx", func(t *testing.T) {
		tx := &pb.Transaction{
			From: []byte("invalid"),
		}
		err = g1.PublishTx(tx)

		assert.NoError(t, err)

		assertNotReceive(t, c1)
		assertNotReceive(t, c2)
	})
}

func assertReceive(t *testing.T, c chan *pb.Transaction, want *pb.Transaction) {
	select {
	case got := <-c:
		assert.Equal(t, want, got)
	case <-time.After(time.Second * 2):
		t.Fatalf("timed out waiting for transaction: %+v", want)
	}
}

func assertNotReceive(t *testing.T, c chan *pb.Transaction) {
	select {
	case got := <-c:
		t.Fatalf("received unexpected transaction: %+v", got)
	case <-time.After(time.Second * 2):
		return
	}
}
