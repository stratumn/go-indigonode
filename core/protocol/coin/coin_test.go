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

	"github.com/stratumn/alice/core/p2p"
	ctestutil "github.com/stratumn/alice/core/protocol/coin/testutil"
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

		ctestutil.WaitUntil(t, func() bool {
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

		ctestutil.WaitUntil(t, func() bool {
			return v1.ValidatedTx(tx1.GetTx()) && v0.ValidatedTx(tx2.GetTx())
		}, "validator.ValidatedTx")
	})
}
