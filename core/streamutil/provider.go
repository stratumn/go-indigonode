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

package streamutil

import (
	"context"

	"github.com/pkg/errors"

	logging "gx/ipfs/QmSpJByNKFX1sCsHBEp3R73FL4NF6FnQTEGyNAXHm2GS52/go-log"
	inet "gx/ipfs/QmXoz9o2PT3tEzf7hicegwex5UgVP54n3k82K7jrWFyN86/go-libp2p-net"
	"gx/ipfs/QmZNkThpqfVXs9GNbexPrfBbXSLNYeKrE7jwFM2oqHbyqN/go-libp2p-protocol"
	"gx/ipfs/QmcJukH2sAFjY3HdBKq35WDzWoL3UUu2gt9wdfqZTUyM74/go-libp2p-peer"
	ihost "gx/ipfs/QmfZTdmunzKzAGJrSvXXQbQ5kLLUiEMX5vdwux7iXkdk7D/go-libp2p-host"
)

// Errors used by the stream provider.
var (
	ErrMissingPeerID      = errors.New("missing peer ID")
	ErrMissingProtocolIDs = errors.New("missing protocol IDs")
)

// Provider lets you configure streams with added features.
type Provider interface {
	NewStream(context.Context, ihost.Host, ...StreamOption) (*WrappedStream, error)
}

// StreamProvider implements the Provider interface.
type StreamProvider struct{}

// NewStreamProvider returns a new provider.
func NewStreamProvider() Provider {
	return &StreamProvider{}
}

// WrappedStream wraps a stream with a Codec and a simplified
// close method.
type WrappedStream struct {
	Conn  inet.Conn
	Codec Codec
	Close func()
}

// StreamOptions are used to configure a stream.
type StreamOptions struct {
	PeerID peer.ID
	PIDs   []protocol.ID
	Event  *logging.EventInProgress
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

// OptLog configures logging.
var OptLog = func(event *logging.EventInProgress) StreamOption {
	return func(opts *StreamOptions) {
		opts.Event = event
	}
}

// NewStream creates a new stream.
func (p *StreamProvider) NewStream(ctx context.Context, host ihost.Host, opts ...StreamOption) (*WrappedStream, error) {
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

	s, err := host.NewStream(ctx, streamOpts.PeerID, streamOpts.PIDs...)
	if err != nil {
		if streamOpts.Event != nil {
			streamOpts.Event.SetError(err)
		}

		return nil, err
	}

	codec := NewProtobufCodec(s)
	close := func() {
		err := s.Close()
		if err != nil && streamOpts.Event != nil {
			streamOpts.Event.Append(logging.Metadata{
				"close_err": err.Error(),
			})
		}
	}

	return &WrappedStream{
		Conn:  s.Conn(),
		Codec: codec,
		Close: close,
	}, nil
}
