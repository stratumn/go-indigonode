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

package monitoring

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"

	"go.opencensus.io/trace"

	logging "gx/ipfs/QmSpJByNKFX1sCsHBEp3R73FL4NF6FnQTEGyNAXHm2GS52/go-log"
	"gx/ipfs/QmWWQ2Txc2c6tqjsBpzg5Ar652cHPGNsQQp2SejkNmkUMb/go-multiaddr"
	"gx/ipfs/QmZNkThpqfVXs9GNbexPrfBbXSLNYeKrE7jwFM2oqHbyqN/go-libp2p-protocol"
	"gx/ipfs/QmcJukH2sAFjY3HdBKq35WDzWoL3UUu2gt9wdfqZTUyM74/go-libp2p-peer"
)

// SpanOption sets some initial settings for the span and context.
type SpanOption func(context.Context, *Span) context.Context

// SpanOptionPeerID sets the span's peer ID and tags the context for metrics.
func SpanOptionPeerID(peerID peer.ID) SpanOption {
	return func(ctx context.Context, span *Span) context.Context {
		span.SetPeerID(peerID)
		return NewTaggedContext(ctx).Tag(PeerIDTag, peerID.Pretty()).Build()
	}
}

// SpanOptionProtocolID sets the span's protocol ID and tags the context for
// metrics.
func SpanOptionProtocolID(pid protocol.ID) SpanOption {
	return func(ctx context.Context, span *Span) context.Context {
		span.SetProtocolID(pid)
		return NewTaggedContext(ctx).Tag(ProtocolIDTag, string(pid)).Build()
	}
}

// loggersLock is used to synchronize access to libp2p's logging.Logger method
// which is not thread-safe.
var loggersLock = sync.Mutex{}

// Span represents a span of a trace. It wraps an OpenCensus span.
// It will also log to the configured libp2p logger.
type Span struct {
	// libp2p's event is unfortunately not thread-safe since it uses a map.
	eventLock sync.Mutex
	event     *logging.EventInProgress

	log  logging.EventLogger
	span *trace.Span
}

// StartSpan starts a new span.
func StartSpan(ctx context.Context, service string, method string, opts ...SpanOption) (context.Context, *Span) {
	loggersLock.Lock()
	log := logging.Logger(service)
	loggersLock.Unlock()

	event := log.EventBegin(ctx, method)
	ctx, s := trace.StartSpan(ctx, fmt.Sprintf("indigo-node/%s/%s", service, method))
	span := &Span{event: event, log: log, span: s}

	for _, opt := range opts {
		ctx = opt(ctx, span)
	}

	return ctx, span
}

// SetStatus sets the status of the span.
func (s *Span) SetStatus(status Status) {
	if status.Code != StatusCodeOK {
		s.eventLock.Lock()
		s.event.SetError(errors.New(status.Message))
		s.eventLock.Unlock()
	}

	s.span.SetStatus(trace.Status{Code: status.Code, Message: status.Message})
}

// SetUnknownError sets the span status with an unclassified error (unknown).
func (s *Span) SetUnknownError(err error) {
	if err != nil {
		s.SetStatus(NewStatus(StatusCodeUnknown, err.Error()))
	}
}

// SetPeerID sets the span's peer ID.
func (s *Span) SetPeerID(peerID peer.ID) {
	s.eventLock.Lock()
	s.event.Append(peerID)
	s.eventLock.Unlock()

	s.span.AddAttributes(trace.StringAttribute("peer_id", peerID.Pretty()))
}

// SetProtocolID sets the span's protocol ID.
func (s *Span) SetProtocolID(pid protocol.ID) {
	s.eventLock.Lock()
	s.event.Append(logging.Metadata{"protocol": string(pid)})
	s.eventLock.Unlock()

	s.span.AddAttributes(trace.StringAttribute("protocol_id", string(pid)))
}

// SetAddrs sets span's addresses.
func (s *Span) SetAddrs(addrs []multiaddr.Multiaddr) {
	addrsStr := make([]string, len(addrs))
	for i, addr := range addrs {
		addrsStr[i] = addr.String()
	}

	s.eventLock.Lock()
	s.event.Append(logging.Metadata{"addresses": addrsStr})
	s.eventLock.Unlock()

	s.span.AddAttributes(trace.StringAttribute("addresses", strings.Join(addrsStr, ", ")))
}

// AddIntAttribute sets an integer attribute in the span.
func (s *Span) AddIntAttribute(name string, value int64) {
	s.eventLock.Lock()
	s.event.Append(logging.Metadata{name: value})
	s.eventLock.Unlock()

	s.span.AddAttributes(trace.Int64Attribute(name, value))
}

// AddBoolAttribute sets a boolean attribute in the span.
func (s *Span) AddBoolAttribute(name string, value bool) {
	s.eventLock.Lock()
	s.event.Append(logging.Metadata{name: value})
	s.eventLock.Unlock()

	s.span.AddAttributes(trace.BoolAttribute(name, value))
}

// AddStringAttribute sets a string attribute in the span.
func (s *Span) AddStringAttribute(name string, value string) {
	s.eventLock.Lock()
	s.event.Append(logging.Metadata{name: value})
	s.eventLock.Unlock()

	s.span.AddAttributes(trace.StringAttribute(name, value))
}

// Annotate adds a log message to a span.
func (s *Span) Annotate(ctx context.Context, name, message string) {
	s.span.Annotate(nil, fmt.Sprintf("%s: %s", name, message))
	s.log.Event(ctx, name, logging.Metadata{
		"message": message,
	})
}

// End ends the span.
func (s *Span) End() {
	s.event.Done()
	s.span.End()
}
