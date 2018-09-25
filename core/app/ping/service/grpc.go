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
	pb "github.com/stratumn/go-indigonode/core/app/ping/grpc"

	peer "gx/ipfs/QmQsErDt8Qgw1XrsXf2BpEzDgGWtB1YLsTAARBup5b6B9W/go-libp2p-peer"
	pstore "gx/ipfs/Qmda4cPRvSRyox3SqgJN6DfSZGU5TtHufPTp9uXjFj71X6/go-libp2p-peerstore"
)

// DefaultPingTimes is the default number of pings.
const DefaultPingTimes = 5

// grpcServer is a gRPC server for the ping service.
type grpcServer struct {
	PingPeer func(context.Context, peer.ID) (<-chan time.Duration, error)
	Connect  func(context.Context, pstore.PeerInfo) error
}

// Ping does a ping request to the specified peer.
func (s grpcServer) Ping(req *pb.PingReq, ss pb.Ping_PingServer) error {
	pid, err := peer.IDFromBytes(req.PeerId)
	if err != nil {
		return errors.WithStack(err)
	}

	pi := pstore.PeerInfo{ID: pid}

	// Make sure there is a connection to the peer.
	if err := s.Connect(ss.Context(), pi); err != nil {
		return err
	}

	pingCtx, cancelPing := context.WithCancel(ss.Context())
	defer cancelPing()

	pong, err := s.PingPeer(pingCtx, pi.ID)
	if err != nil {
		return errors.WithStack(err)
	}

	count := uint32(0)
	times := req.Times
	if times == 0 {
		times = DefaultPingTimes
	}

	for latency := range pong {
		err := ss.Send(&pb.Response{
			Latency: int64(latency),
		})
		if err != nil {
			return errors.WithStack(err)
		}

		count++
		if count >= times {
			cancelPing()
		}
	}

	return nil
}
