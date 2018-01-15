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

package chat

import (
	"context"
	"testing"
	"time"

	"github.com/stratumn/alice/core/p2p"
	pbevent "github.com/stratumn/alice/grpc/event"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	inet "gx/ipfs/QmU4vCDZTPLDqSDKguWbHCiUe46mZUtmM2g2suBZ9NE8ko/go-libp2p-net"
	testutil "gx/ipfs/QmZTcPxK6VqrwY94JpKZPvEqAZ6tEr1rLrpcqJbbRZbg2V/go-libp2p-netutil"
)

func TestChat(t *testing.T) {
	t.Run("Send and receive message", func(t *testing.T) {
		ctx := context.Background()
		h1 := p2p.NewHost(ctx, testutil.GenSwarmNetwork(t, ctx))
		h2 := p2p.NewHost(ctx, testutil.GenSwarmNetwork(t, ctx))
		defer h1.Close()
		defer h2.Close()

		h2pi := h2.Peerstore().PeerInfo(h2.ID())
		require.NoError(t, h1.Connect(ctx, h2pi), "Connect()")

		chat1 := NewChat(h1)
		chat2 := NewChat(h2)
		receiveChan := make(chan *pbevent.Event, 1)
		chat2.listeners = []chan *pbevent.Event{receiveChan}

		h2.SetStreamHandler(ProtocolID, func(stream inet.Stream) {
			chat2.StreamHandler(ctx, stream)
		})

		err := chat1.Send(ctx, h2.ID(), "hello world!")
		require.NoError(t, err, "chat.Send()")

		select {
		case <-time.After(1 * time.Second):
			assert.Fail(t, "Message not received")
		case <-receiveChan:
			break
		}
	})

	t.Run("Closes all listener channels when stopping", func(t *testing.T) {
		ctx := context.Background()
		h1 := p2p.NewHost(ctx, testutil.GenSwarmNetwork(t, ctx))
		defer h1.Close()

		chat1 := NewChat(h1)
		listeners := []chan *pbevent.Event{
			make(chan *pbevent.Event, 1),
			make(chan *pbevent.Event, 1),
		}
		chat1.listeners = listeners

		chat1.Close()

		_, ok := <-listeners[0]
		assert.False(t, ok, "Listener1 channel should be closed")
		_, ok = <-listeners[1]
		assert.False(t, ok, "Listener2 channel should be closed")
	})
}
