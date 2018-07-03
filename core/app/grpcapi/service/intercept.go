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

package service

import (
	"context"

	"github.com/pkg/errors"
	"github.com/stratumn/go-indigonode/core/monitoring"

	"google.golang.org/grpc"
	"google.golang.org/grpc/peer"

	manet "gx/ipfs/QmRK2LxanhK2gZq6k6R7vk5ZoYZk8ULSSTB7FzDsMUX6CB/go-multiaddr-net"
	logging "gx/ipfs/QmSpJByNKFX1sCsHBEp3R73FL4NF6FnQTEGyNAXHm2GS52/go-log"
)

// logRequest is used to log requests.
func logRequest(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (interface{}, error) {
	ctx, _ = monitoring.NewTaggedContext(ctx).Tag(methodTag, info.FullMethod).Build()
	requestReceived.Record(ctx, 1)

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
	ctx, _ := monitoring.NewTaggedContext(ss.Context()).Tag(methodTag, info.FullMethod).Build()
	requestReceived.Record(ctx, 1)

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
