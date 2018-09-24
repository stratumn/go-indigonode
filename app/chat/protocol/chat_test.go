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
	"fmt"
	"testing"
	"time"

	pbchat "github.com/stratumn/go-indigonode/app/chat/grpc"
	pbevent "github.com/stratumn/go-indigonode/core/app/event/grpc"
	event "github.com/stratumn/go-indigonode/core/app/event/service"
	"github.com/stratumn/go-indigonode/core/p2p"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	inet "gx/ipfs/QmZNJyx9GGCX4GeuHnLB8fxaxMLs4MjTjHokxfQcCd6Nve/go-libp2p-net"
	swarmtesting "gx/ipfs/QmeDpqUwwdye8ABKVMPXKuWwPVURFdqTqssbTUB39E2Nwd/go-libp2p-swarm/testing"
)

func TestChat(t *testing.T) {
	ctx := context.Background()
	h1 := p2p.NewHost(ctx, swarmtesting.GenSwarm(t, ctx))
	h2 := p2p.NewHost(ctx, swarmtesting.GenSwarm(t, ctx))
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
