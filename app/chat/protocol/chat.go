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

// Package protocol defines types for the chat protocol.
package protocol

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	pbchat "github.com/stratumn/go-indigonode/app/chat/grpc"
	"github.com/stratumn/go-indigonode/app/chat/pb"
	pbevent "github.com/stratumn/go-indigonode/core/app/event/grpc"
	event "github.com/stratumn/go-indigonode/core/app/event/service"
	"github.com/stratumn/go-indigonode/core/monitoring"

	protobuf "gx/ipfs/QmRDePEiL4Yupq5EkcK3L3ko3iMgYaqUdLu7xc1kqs7dnV/go-multicodec/protobuf"
	logging "gx/ipfs/QmSpJByNKFX1sCsHBEp3R73FL4NF6FnQTEGyNAXHm2GS52/go-log"
	inet "gx/ipfs/QmXoz9o2PT3tEzf7hicegwex5UgVP54n3k82K7jrWFyN86/go-libp2p-net"
	protocol "gx/ipfs/QmZNkThpqfVXs9GNbexPrfBbXSLNYeKrE7jwFM2oqHbyqN/go-libp2p-protocol"
	peer "gx/ipfs/QmcJukH2sAFjY3HdBKq35WDzWoL3UUu2gt9wdfqZTUyM74/go-libp2p-peer"
	ihost "gx/ipfs/QmfZTdmunzKzAGJrSvXXQbQ5kLLUiEMX5vdwux7iXkdk7D/go-libp2p-host"
)

// Host represents an Indigo Node host.
type Host = ihost.Host

// ProtocolID is the protocol ID of the service.
var ProtocolID = protocol.ID("/indigo/node/chat/v1.0.0")

// log is the logger for the service.
var log = logging.Logger("chat")

// Chat implements the chat protocol.
type Chat struct {
	host          Host
	eventEmitter  event.Emitter
	msgReceivedCh chan *pbchat.DatedMessage
}

// NewChat creates a new chat server.
func NewChat(host Host, eventEmitter event.Emitter, msgReceivedCh chan *pbchat.DatedMessage) *Chat {
	return &Chat{
		host:          host,
		eventEmitter:  eventEmitter,
		msgReceivedCh: msgReceivedCh,
	}
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

	ctx = monitoring.NewTaggedContext(ctx).Tag(monitoring.PeerIDTag, from).Build()
	msgReceived.Record(ctx, 1)

	dec := protobuf.Multicodec(nil).Decoder(stream)
	var message pb.Message
	err := dec.Decode(&message)
	if err != nil {
		event.SetError(err)
		msgError.Record(ctx, 1)
		return
	}

	event.Append(logging.Metadata{
		"message": message.Message,
	})

	go func() {
		c.msgReceivedCh <- pbchat.NewDatedMessageReceived(stream.Conn().RemotePeer(), message.Message)
	}()

	chatEvent := &pbevent.Event{
		Message: fmt.Sprintf("[%s] %s", from, message.Message),
		Level:   pbevent.Level_INFO,
		Topic:   fmt.Sprintf("chat.%s", from),
	}

	c.eventEmitter.Emit(chatEvent)
}

// Send sends a message to a peer.
func (c *Chat) Send(ctx context.Context, pid peer.ID, message string) error {
	event := log.EventBegin(ctx, "Send", logging.Metadata{
		"peerID": pid.Pretty(),
	})
	defer event.Done()

	ctx = monitoring.NewTaggedContext(ctx).Tag(monitoring.PeerIDTag, pid.Pretty()).Build()
	msgSent.Record(ctx, 1)

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
		err = enc.Encode(&pb.Message{Message: message})
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
		msgError.Record(ctx, 1)
		return err
	}
}
