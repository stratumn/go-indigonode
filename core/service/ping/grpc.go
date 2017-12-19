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
	"time"

	"github.com/pkg/errors"
	pb "github.com/stratumn/alice/grpc/ping"

	peer "gx/ipfs/QmWNY7dV54ZDYmTA1ykVdwNCqC11mpU4zSUp6XDpLTH9eG/go-libp2p-peer"
	pstore "gx/ipfs/QmYijbtjCxFEjSXaudaQAUz3LN5VKLssm8WCUsRoqzXmQR/go-libp2p-peerstore"
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
