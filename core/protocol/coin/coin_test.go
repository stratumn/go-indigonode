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
	"github.com/stratumn/alice/core/protocol/coin/engine/mockengine"
	"github.com/stratumn/alice/core/protocol/coin/state/mockstate"
	ctestutil "github.com/stratumn/alice/core/protocol/coin/testutil"
	pb "github.com/stratumn/alice/pb/coin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	ihost "gx/ipfs/QmP46LGWhzVZTMmt5akNNLfoV8qL4h5wTwmzQxLyDafggd/go-libp2p-host"
	protobuf "gx/ipfs/QmRDePEiL4Yupq5EkcK3L3ko3iMgYaqUdLu7xc1kqs7dnV/go-multicodec/protobuf"
	inet "gx/ipfs/QmU4vCDZTPLDqSDKguWbHCiUe46mZUtmM2g2suBZ9NE8ko/go-libp2p-net"
	testutil "gx/ipfs/QmZTcPxK6VqrwY94JpKZPvEqAZ6tEr1rLrpcqJbbRZbg2V/go-libp2p-netutil"
)

func TestCoinProtocolHandler(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	hosts := make([]ihost.Host, 2)
	coins := make([]*Coin, 2)

	for i := 0; i < 2; i++ {
		hosts[i] = p2p.NewHost(ctx, testutil.GenSwarmNetwork(t, ctx))
		defer func(h ihost.Host) {
			h.Close()
		}(hosts[i])

		coins[i] = &Coin{
			mempool: mockstate.NewMockMempool(ctrl),
			engine:  mockengine.NewMockEngine(ctrl),
		}

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
		coins[1].engine.(*mockengine.MockEngine).
			EXPECT().
			VerifyHeader(nil, ctestutil.NewHeaderMatcher(block1.GetBlock().Header.BlockNumber)).
			Times(1)

		err = enc0_1.Encode(block1)
		assert.NoError(t, err, "Encode()")

		block2 := ctestutil.RandomGossipBlock()
		coins[0].engine.(*mockengine.MockEngine).
			EXPECT().
			VerifyHeader(nil, ctestutil.NewHeaderMatcher(block2.GetBlock().Header.BlockNumber)).
			Times(1)

		err = enc1_0.Encode(block2)
		assert.NoError(t, err, "Encode()")

		tx := ctestutil.RandomGossipTx()
		coins[1].mempool.(*mockstate.MockMempool).
			EXPECT().
			AddTransaction(ctestutil.NewTxMatcher(tx.GetTx().Value)).
			Times(1)

		err = enc0_1.Encode(tx)
		assert.NoError(t, err, "Encode()")
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
		coins[1].mempool.(*mockstate.MockMempool).
			EXPECT().
			AddTransaction(ctestutil.NewTxMatcher(tx1.GetTx().Value)).
			Times(1)

		enc0_1 := protobuf.Multicodec(nil).Encoder(s0_1)
		err = enc0_1.Encode(tx1)
		assert.NoError(t, err, "Encode()")

		err = s0_1.Close()
		assert.NoError(t, err, "Close()")

		tx2 := ctestutil.RandomGossipTx()
		coins[0].mempool.(*mockstate.MockMempool).
			EXPECT().
			AddTransaction(ctestutil.NewTxMatcher(tx2.GetTx().Value)).
			Times(1)

		enc1_0 := protobuf.Multicodec(nil).Encoder(s1_0)
		err = enc1_0.Encode(tx2)
		assert.NoError(t, err, "Encode()")
	})
}
