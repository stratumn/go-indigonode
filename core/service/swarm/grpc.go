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

package swarm

import (
	"context"

	"github.com/pkg/errors"

	pb "github.com/stratumn/alice/grpc/swarm"

	swarm "gx/ipfs/QmSwZMWwFZSUpe5muU2xgTUwppH24KfMwdPXiwbEp2c6G5/go-libp2p-swarm"
	peer "gx/ipfs/QmZoWKhxUmZ2seW4BzX6fJkNR8hh9PsGModr7q171yq2SS/go-libp2p-peer"
)

// grpcServer is a gRPC server for the swarm service.
type grpcServer struct {
	GetSwarm func() *swarm.Swarm
}

// LocalPeer returns the local peer.
func (s grpcServer) LocalPeer(ctx context.Context, req *pb.LocalPeerReq) (*pb.Peer, error) {
	swm := s.GetSwarm()
	if swm == nil {
		return nil, errors.WithStack(ErrUnavailable)
	}

	return &pb.Peer{Id: []byte(swm.LocalPeer())}, nil
}

// Peers lists the peers connected to the local peer.
func (s grpcServer) Peers(req *pb.PeersReq, ss pb.Swarm_PeersServer) error {
	swm := s.GetSwarm()
	if swm == nil {
		return errors.WithStack(ErrUnavailable)
	}

	for _, pid := range swm.Peers() {
		if err := ss.Send(&pb.Peer{Id: []byte(pid)}); err != nil {
			return errors.WithStack(err)
		}
	}

	return nil
}

// Peers lists the connections of the swarm.
func (s grpcServer) Connections(req *pb.ConnectionsReq, ss pb.Swarm_ConnectionsServer) error {
	swm := s.GetSwarm()
	if swm == nil {
		return errors.WithStack(ErrUnavailable)
	}

	var conns []*swarm.Conn

	if len(req.PeerId) < 1 {
		conns = swm.Connections()
	} else {
		pid, err := peer.IDFromBytes(req.PeerId)
		if err != nil {
			return errors.WithStack(err)
		}

		conns = swm.ConnectionsToPeer(pid)
	}

	for _, conn := range conns {
		err := ss.Send(&pb.Connection{
			PeerId:        []byte(conn.RemotePeer()),
			LocalAddress:  conn.LocalMultiaddr().Bytes(),
			RemoteAddress: conn.RemoteMultiaddr().Bytes(),
		})
		if err != nil {
			return errors.WithStack(err)
		}
	}

	return nil
}
