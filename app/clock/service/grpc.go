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
	pb "github.com/stratumn/go-node/app/clock/grpc"

	peer "github.com/libp2p/go-libp2p-peer"
	pstore "github.com/libp2p/go-libp2p-peerstore"
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
