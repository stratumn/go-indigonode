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

	"github.com/pkg/errors"
	pb "github.com/stratumn/alice/grpc/chat"
	pbevent "github.com/stratumn/alice/grpc/event"

	protobuf "gx/ipfs/QmRDePEiL4Yupq5EkcK3L3ko3iMgYaqUdLu7xc1kqs7dnV/go-multicodec/protobuf"
	logging "gx/ipfs/QmSpJByNKFX1sCsHBEp3R73FL4NF6FnQTEGyNAXHm2GS52/go-log"
	inet "gx/ipfs/QmU4vCDZTPLDqSDKguWbHCiUe46mZUtmM2g2suBZ9NE8ko/go-libp2p-net"
	peer "gx/ipfs/QmWNY7dV54ZDYmTA1ykVdwNCqC11mpU4zSUp6XDpLTH9eG/go-libp2p-peer"
	protocol "gx/ipfs/QmZNkThpqfVXs9GNbexPrfBbXSLNYeKrE7jwFM2oqHbyqN/go-libp2p-protocol"
)

// ProtocolID is the protocol ID of the service.
var ProtocolID = protocol.ID("/alice/chat/v1.0.0")

// Chat implements the chat protocol.
type Chat struct {
	host      Host
	listeners []chan *pbevent.Event
}

// NewChat creates a new chat server.
func NewChat(host Host) *Chat {
	return &Chat{host: host}
}

// StreamHandler handles incoming messages from peers.
func (c *Chat) StreamHandler(ctx context.Context, stream inet.Stream) {
	log.Event(ctx, "beginStream", logging.Metadata{
		"stream": stream,
	})
	defer log.Event(ctx, "endStream", logging.Metadata{
		"stream": stream,
	})

	c.receive(ctx, stream)
	if err := stream.Close(); err != nil {
		log.Event(ctx, "streamCloseError", logging.Metadata{
			"error":  err.Error(),
			"stream": stream,
		})
	}
}

// receive reads a message from an incoming stream.
func (c *Chat) receive(ctx context.Context, stream inet.Stream) {
	from := stream.Conn().RemotePeer().Pretty()
	event := log.EventBegin(ctx, "Receive", logging.Metadata{
		"peerID": from,
	})
	defer event.Done()

	dec := protobuf.Multicodec(nil).Decoder(stream)
	var message pb.ChatMessage
	err := dec.Decode(&message)
	if err != nil {
		event.SetError(err)
		return
	}

	event.Append(logging.Metadata{
		"message": message.Message,
	})

	chatEvent := &pbevent.Event{
		Message: message.Message,
		Level:   pbevent.Event_INFO,
	}

	for _, listener := range c.listeners {
		listener <- chatEvent
	}
}

// Send sends a message to a peer.
func (c *Chat) Send(ctx context.Context, pid peer.ID, message string) error {
	event := log.EventBegin(ctx, "Send", logging.Metadata{
		"peerID": pid.Pretty(),
	})
	defer event.Done()

	successCh := make(chan struct{}, 1)
	errCh := make(chan error, 1)

	go func() {
		stream, err := c.host.NewStream(ctx, pid, ProtocolID)
		if err != nil {
			event.SetError(err)
			errCh <- errors.WithStack(err)
			return
		}

		enc := protobuf.Multicodec(nil).Encoder(stream)
		err = enc.Encode(&pb.ChatMessage{Message: message})
		if err != nil {
			event.SetError(err)
			errCh <- errors.WithStack(err)
			return
		}

		successCh <- struct{}{}
	}()

	select {
	case <-ctx.Done():
		return errors.WithStack(ctx.Err())
	case <-successCh:
		return nil
	case err := <-errCh:
		return err
	}
}

// Close closes the chat server.
func (c *Chat) Close() {
	for _, listener := range c.listeners {
		close(listener)
	}

	c.listeners = nil
}
