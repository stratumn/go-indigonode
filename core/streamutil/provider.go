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

package streamutil

import (
	"context"

	"github.com/pkg/errors"
	"github.com/stratumn/go-node/core/monitoring"

	ihost "github.com/libp2p/go-libp2p-host"
	inet "github.com/libp2p/go-libp2p-net"
	peer "github.com/libp2p/go-libp2p-peer"
	protocol "github.com/libp2p/go-libp2p-protocol"
)

// Errors used by the stream provider.
var (
	ErrMissingPeerID      = errors.New("missing peer ID")
	ErrMissingProtocolIDs = errors.New("missing protocol IDs")
)

// Stream is a simplified stream abstraction.
type Stream interface {
	Conn() inet.Conn
	Codec() Codec
	Close()
}

type wrappedStream struct {
	codec  Codec
	stream inet.Stream
	span   *monitoring.Span
}

// WrapStream wraps a stream to our simplified abstraction.
func wrapStream(stream inet.Stream, span *monitoring.Span) Stream {
	codec := NewProtobufCodec(stream)

	return &wrappedStream{
		stream: stream,
		codec:  codec,
		span:   span,
	}
}

func (s *wrappedStream) Conn() inet.Conn {
	return s.stream.Conn()
}

func (s *wrappedStream) Codec() Codec {
	return s.codec
}

func (s *wrappedStream) Close() {
	err := inet.FullClose(s.stream)
	if err != nil {
		s.span.Annotate(context.TODO(), "close_err", err.Error())
	}

	s.span.End()
}

// Provider lets you configure streams with added features.
type Provider interface {
	NewStream(context.Context, ihost.Host, ...StreamOption) (Stream, error)
}

// StreamProvider implements the Provider interface.
type StreamProvider struct{}

// NewStreamProvider returns a new provider.
func NewStreamProvider() Provider {
	return &StreamProvider{}
}

// StreamOptions are used to configure a stream.
type StreamOptions struct {
	PeerID peer.ID
	PIDs   []protocol.ID
}

// StreamOption configures a single stream option.
type StreamOption func(opts *StreamOptions)

// OptPeerID configures the remote peer.
var OptPeerID = func(peerID peer.ID) StreamOption {
	return func(opts *StreamOptions) {
		opts.PeerID = peerID
	}
}

// OptProtocolIDs configures the protocol IDs.
var OptProtocolIDs = func(pids ...protocol.ID) StreamOption {
	return func(opts *StreamOptions) {
		opts.PIDs = pids
	}
}

// NewStream creates a new stream.
func (p *StreamProvider) NewStream(ctx context.Context, host ihost.Host, opts ...StreamOption) (Stream, error) {
	ctx, span := monitoring.StartSpan(ctx, "streamutil", "NewStream")
	// We don't end the span here, it will end when Close() is called on the
	// stream we create.

	streamOpts := &StreamOptions{}
	for _, opt := range opts {
		opt(streamOpts)
	}

	if streamOpts.PeerID == "" {
		return nil, ErrMissingPeerID
	}

	if len(streamOpts.PIDs) == 0 {
		return nil, ErrMissingProtocolIDs
	}

	span.SetPeerID(streamOpts.PeerID)
	span.SetProtocolID(streamOpts.PIDs[0])

	s, err := host.NewStream(ctx, streamOpts.PeerID, streamOpts.PIDs...)
	if err != nil {
		span.SetUnknownError(err)
		span.End()

		return nil, err
	}

	return wrapStream(s, span), nil
}
