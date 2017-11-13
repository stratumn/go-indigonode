// Copyright © 2017  Stratumn SAS
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

package ping

import (
	"context"

	"github.com/pkg/errors"
	pb "github.com/stratumn/alice/grpc/ping"

	pstore "gx/ipfs/QmPgDWmTmuzvP7QE5zwo1TmjbJme9pmZHNujB2453jkCTr/go-libp2p-peerstore"
	peer "gx/ipfs/QmXYjuNuxVzXKJCfWasQk1RqkhVLDM9jtUKhqc2WPQmFSB/go-libp2p-peer"
)

// DefaultPingTimes is the default number of pings.
const DefaultPingTimes = 5

// grpcServer is a gRPC server for the ping service.
type grpcServer struct {
	service *Service
}

// Ping does a ping request to the specified address.
func (s grpcServer) Ping(req *pb.PingReq, ss pb.Ping_PingServer) error {
	ping := s.service.ping
	if ping == nil {
		return errors.WithStack(ErrUnavailable)
	}

	pid, err := peer.IDFromBytes(req.PeerId)
	if err != nil {
		return errors.WithStack(err)
	}

	pi := pstore.PeerInfo{ID: pid}

	// Make sure there is a connection to the peer.
	if err := s.service.host.Connect(ss.Context(), pi); err != nil {
		return err
	}

	pingCtx, cancelPing := context.WithCancel(ss.Context())
	defer cancelPing()

	pong, err := ping.Ping(pingCtx, pi.ID)
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
