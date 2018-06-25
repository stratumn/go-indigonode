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
	"time"

	"github.com/pkg/errors"
	pb "github.com/stratumn/go-indigonode/app/clock/grpc"

	peer "gx/ipfs/QmcJukH2sAFjY3HdBKq35WDzWoL3UUu2gt9wdfqZTUyM74/go-libp2p-peer"
	pstore "gx/ipfs/QmdeiKhUy1TVGBaKxt7y1QmBDLBdisSrLJ1x58Eoj4PXUh/go-libp2p-peerstore"
)

// grpcServer is a gRPC server for the clock service.
type grpcServer struct {
	LocalTime  func(context.Context) (*time.Time, error)
	RemoteTime func(context.Context, peer.ID) (*time.Time, error)
	Connect    func(context.Context, pstore.PeerInfo) error
}

// Local gets the local time of the node.
func (s grpcServer) Local(ctx context.Context, req *pb.LocalReq) (*pb.Time, error) {
	t, err := s.LocalTime(ctx)
	if err != nil {
		return nil, err
	}

	return &pb.Time{
		Timestamp: t.UnixNano(),
	}, nil
}

// Remote gets the time of the specified peer.
func (s grpcServer) Remote(ctx context.Context, req *pb.RemoteReq) (*pb.Time, error) {
	pid, err := peer.IDFromBytes(req.PeerId)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	pi := pstore.PeerInfo{ID: pid}

	// Make sure there is a connection to the peer.
	if err := s.Connect(ctx, pi); err != nil {
		return nil, err
	}

	t, err := s.RemoteTime(ctx, pid)
	if err != nil {
		return nil, err
	}

	return &pb.Time{
		Timestamp: t.UnixNano(),
	}, nil
}
