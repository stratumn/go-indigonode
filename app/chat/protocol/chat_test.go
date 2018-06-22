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

package protocol

import (
	"context"
	"fmt"
	"testing"
	"time"

	pbchat "github.com/stratumn/alice/app/chat/grpc"
	pbevent "github.com/stratumn/alice/core/app/event/grpc"
	event "github.com/stratumn/alice/core/app/event/service"
	"github.com/stratumn/alice/core/p2p"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	inet "gx/ipfs/QmXoz9o2PT3tEzf7hicegwex5UgVP54n3k82K7jrWFyN86/go-libp2p-net"
	testutil "gx/ipfs/Qmb6BsZf6Y3kxffXMNTubGPF1w1bkHtpvhfYbmnwP3NQyw/go-libp2p-netutil"
)

func TestChat(t *testing.T) {
	ctx := context.Background()
	h1 := p2p.NewHost(ctx, testutil.GenSwarmNetwork(t, ctx))
	h2 := p2p.NewHost(ctx, testutil.GenSwarmNetwork(t, ctx))
	defer h1.Close()
	defer h2.Close()

	h2pi := h2.Peerstore().PeerInfo(h2.ID())
	require.NoError(t, h1.Connect(ctx, h2pi), "Connect()")

	msgReceivedCh := make(chan *pbchat.DatedMessage)

	chat1 := NewChat(h1, event.NewEmitter(event.DefaultTimeout), msgReceivedCh)

	t.Run("Send and receive message", func(t *testing.T) {
		eventEmitter := event.NewEmitter(event.DefaultTimeout)
		receiveChan, err := eventEmitter.AddListener("chat.*")
		assert.NoError(t, err, "eventEmitter.AddListener(chat.*)")

		chat2 := NewChat(h2, eventEmitter, msgReceivedCh)

		h2.SetStreamHandler(ProtocolID, func(stream inet.Stream) {
			chat2.StreamHandler(ctx, stream)
		})

		err = chat1.Send(ctx, h2.ID(), "hello world!")
		require.NoError(t, err, "chat.Send()")

		select {
		case <-time.After(1 * time.Second):
			assert.Fail(t, "chat.Send() did not send message")
		case ev := <-receiveChan:
			assert.Contains(t, ev.Message, "hello world!", "event.Message")
			assert.Contains(t, ev.Topic, fmt.Sprintf("chat.%s", h1.ID().Pretty()), "event.Topic")
			assert.Equal(t, pbevent.Level_INFO, ev.Level, "event.Level")
		}
	})

	t.Run("Send message without listeners on the receiving end doesn't block", func(t *testing.T) {
		// We don't register a listener on chat2's end.
		chat2 := NewChat(h2, event.NewEmitter(event.DefaultTimeout), msgReceivedCh)

		receiveChan := make(chan struct{})

		h2.SetStreamHandler(ProtocolID, func(stream inet.Stream) {
			chat2.StreamHandler(ctx, stream)
			receiveChan <- struct{}{} // message handled without blocking
		})

		err := chat1.Send(ctx, h2.ID(), "hello no-one!")
		require.NoError(t, err, "chat.Send()")

		select {
		case <-time.After(1 * time.Second):
			assert.Fail(t, "chat.Send() did not send message")
		case <-receiveChan:
			break
		}
	})

	t.Run("Notify received messages", func(t *testing.T) {
		chat2 := NewChat(h2, event.NewEmitter(event.DefaultTimeout), msgReceivedCh)

		h2.SetStreamHandler(ProtocolID, func(stream inet.Stream) {
			chat2.StreamHandler(ctx, stream)
		})

		err := chat1.Send(ctx, h2.ID(), "hello world!")
		require.NoError(t, err, "chat.Send()")

		select {
		case <-time.After(1 * time.Second):
			assert.Fail(t, "chat.Send() did not send message")
		case msg := <-msgReceivedCh:
			assert.Contains(t, msg.Content, "hello world!", "event.Message")
			assert.Equal(t, msg.From, []byte(h1.ID()))
		}
	})
}
