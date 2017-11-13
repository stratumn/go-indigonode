// Copyright © 2017 Stratumn SAS
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

package grpcapi

import (
	"context"

	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/peer"

	logging "gx/ipfs/QmSpJByNKFX1sCsHBEp3R73FL4NF6FnQTEGyNAXHm2GS52/go-log"
	manet "gx/ipfs/QmX3U3YXCQ6UYBxq2LVWF8dARS1hPUTEYLrSx654Qyxyw6/go-multiaddr-net"
)

// logRequest is used to log requests.
func logRequest(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (interface{}, error) {
	event := log.EventBegin(ctx, "request", logging.Metadata{
		"request": req,
		"method":  info.FullMethod,
	})
	defer event.Done()

	// Get the gRPC peer from the context.
	pr, ok := peer.FromContext(ctx)
	if !ok {
		event.SetError(ErrPeerNotFound)
		return nil, errors.WithStack(ErrPeerNotFound)
	}

	// Convert the peer address to a multiaddr.
	addr := pr.Addr.String()
	event.Append(logging.Metadata{"netaddr": addr})

	maddr, err := manet.FromNetAddr(pr.Addr)
	if err != nil {
		event.SetError(err)
		return nil, errors.WithStack(err)
	}
	event.Append(logging.Metadata{"multiaddr": maddr})

	// Pass the request to the handler.
	res, err := handler(ctx, req)
	if err != nil {
		event.SetError(err)
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
	event := log.EventBegin(ss.Context(), "request", logging.Metadata{
		"request": srv,
		"method":  info.FullMethod,
	})
	defer event.Done()

	// Get the gRPC peer from the stream's context.
	pr, ok := peer.FromContext(ss.Context())
	if !ok {
		event.SetError(ErrPeerNotFound)
		return errors.WithStack(ErrPeerNotFound)
	}

	// Convert the peer address to a multiaddr.
	addr := pr.Addr.String()
	event.Append(logging.Metadata{"netaddr": addr})

	maddr, err := manet.FromNetAddr(pr.Addr)
	if err != nil {
		event.SetError(err)
		return errors.WithStack(err)
	}
	event.Append(logging.Metadata{"multiaddr": maddr})

	// Pass the stream to the handler.
	err = handler(srv, ss)
	if err != nil {
		event.SetError(err)
		return errors.WithStack(err)
	}

	return nil
}