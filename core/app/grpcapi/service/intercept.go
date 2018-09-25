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

package service

import (
	"context"
	"time"

	"github.com/pkg/errors"
	"github.com/stratumn/go-node/core/monitoring"

	"google.golang.org/grpc"
	"google.golang.org/grpc/peer"

	logging "gx/ipfs/QmSpJByNKFX1sCsHBEp3R73FL4NF6FnQTEGyNAXHm2GS52/go-log"
	manet "gx/ipfs/QmV6FjemM1K8oXjrvuq3wuVWWoU2TLDPmNnKrxHzY3v6Ai/go-multiaddr-net"
)

// logRequest is used to log requests.
func logRequest(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (interface{}, error) {
	ctx = monitoring.NewTaggedContext(ctx).Tag(methodTag, info.FullMethod).Build()
	requestReceived.Record(ctx, 1)

	start := time.Now()
	defer requestDuration.Record(ctx, float64(time.Since(start))/float64(time.Millisecond))

	event := log.EventBegin(ctx, "request", logging.Metadata{
		"request": req,
		"method":  info.FullMethod,
	})
	defer event.Done()

	// Get the gRPC peer from the context.
	pr, ok := peer.FromContext(ctx)
	if !ok {
		event.SetError(ErrPeerNotFound)
		requestErr.Record(ctx, 1)
		return nil, errors.WithStack(ErrPeerNotFound)
	}

	// Convert the peer address to a multiaddr.
	addr := pr.Addr.String()
	event.Append(logging.Metadata{"netaddr": addr})

	ma, err := manet.FromNetAddr(pr.Addr)
	// With grpcweb, the request does not necessarily come from a multiaddr.
	if err == nil {
		event.Append(logging.Metadata{"multiaddr": ma})
	}

	// Pass the request to the handler.
	res, err := handler(ctx, req)
	if err != nil {
		event.SetError(err)
		requestErr.Record(ctx, 1)
		return nil, errors.WithStack(err)
	}

	event.Append(logging.Metadata{"response": res})

	return res, nil
}

// logStream is used to log streams.
func logStream(
	srv interface{},
	ss grpc.ServerStream,
	info *grpc.StreamServerInfo,
	handler grpc.StreamHandler,
) error {
	ctx := monitoring.NewTaggedContext(ss.Context()).Tag(methodTag, info.FullMethod).Build()
	requestReceived.Record(ctx, 1)

	start := time.Now()
	defer requestDuration.Record(ctx, float64(time.Since(start).Nanoseconds())/1e6)

	event := log.EventBegin(ctx, "request", logging.Metadata{
		"method":       info.FullMethod,
		"serverStream": info.IsServerStream,
		"clientStream": info.IsClientStream,
	})
	defer event.Done()

	// Get the gRPC peer from the stream's context.
	pr, ok := peer.FromContext(ctx)
	if !ok {
		event.SetError(ErrPeerNotFound)
		requestErr.Record(ctx, 1)
		return errors.WithStack(ErrPeerNotFound)
	}

	// Convert the peer address to a multiaddr.
	addr := pr.Addr.String()
	event.Append(logging.Metadata{"netaddr": addr})

	ma, err := manet.FromNetAddr(pr.Addr)
	// With grpcweb, the request does not necessarily come from a multiaddr.
	if err == nil {
		event.Append(logging.Metadata{"multiaddr": ma})
	}

	// Pass the stream to the handler.
	err = handler(srv, ss)
	if err != nil {
		event.SetError(err)
		requestErr.Record(ctx, 1)
		return errors.WithStack(err)
	}

	return nil
}
