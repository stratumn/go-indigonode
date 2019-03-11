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

	"github.com/pkg/errors"

	pb "github.com/stratumn/go-node/core/app/swarm/grpc"

	peer "github.com/libp2p/go-libp2p-peer"
	inet "github.com/libp2p/go-libp2p-net"
	swarm "github.com/libp2p/go-libp2p-swarm"
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

	var conns []inet.Conn

	if len(req.PeerId) < 1 {
		conns = swm.Conns()
	} else {
		pid, err := peer.IDFromBytes(req.PeerId)
		if err != nil {
			return errors.WithStack(err)
		}

		conns = swm.ConnsToPeer(pid)
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
