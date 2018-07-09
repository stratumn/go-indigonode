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

	"go.opencensus.io/trace"

	logging "gx/ipfs/QmSpJByNKFX1sCsHBEp3R73FL4NF6FnQTEGyNAXHm2GS52/go-log"
	protocol "gx/ipfs/QmZNkThpqfVXs9GNbexPrfBbXSLNYeKrE7jwFM2oqHbyqN/go-libp2p-protocol"
	peer "gx/ipfs/QmcJukH2sAFjY3HdBKq35WDzWoL3UUu2gt9wdfqZTUyM74/go-libp2p-peer"
)

// Span represents a span of a trace. It wraps an OpenCensus span.
// It will also log to the configured libp2p logger.
type Span struct {
	event *logging.EventInProgress
	span  *trace.Span
}

// StartSpan starts a new span.
func StartSpan(ctx context.Context, service string, method string) (context.Context, *Span) {
	log := logging.Logger(service)
	event := log.EventBegin(ctx, method)
	ctx, span := trace.StartSpan(ctx, fmt.Sprintf("indigo-node/%s/%s", service, method))
	return ctx, &Span{event: event, span: span}
}

// SetStatus sets the status of the span.
func (s *Span) SetStatus(status Status) {
	if status.Code != StatusCodeOK {
		s.event.SetError(errors.New(status.Message))
	}

	s.span.SetStatus(trace.Status{Code: status.Code, Message: status.Message})
}

// SetPeerID sets the span's peer ID.
func (s *Span) SetPeerID(peerID peer.ID) {
	s.event.Append(peerID)
	s.span.AddAttributes(trace.StringAttribute("peer_id", peerID.Pretty()))
}

// SetProtocolID sets the span's protocol ID.
func (s *Span) SetProtocolID(pid protocol.ID) {
	s.event.Append(logging.Metadata{"protocol": string(pid)})
	s.span.AddAttributes(trace.StringAttribute("protocol_id", string(pid)))
}

// End ends the span.
func (s *Span) End() {
	s.event.Done()
	s.span.End()
}
