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
	"github.com/stratumn/go-node/core/app/bootstrap/grpc"
	"github.com/stratumn/go-node/core/app/bootstrap/pb"
	protocol "github.com/stratumn/go-node/core/app/bootstrap/protocol"
	"github.com/stratumn/go-node/core/app/bootstrap/protocol/proposal"
	"github.com/stratumn/go-node/core/protector"
	protectorpb "github.com/stratumn/go-node/core/protector/pb"

	"github.com/libp2p/go-libp2p-peer"
	"github.com/multiformats/go-multiaddr"
)

// grpcServer is a gRPC server for the bootstrap service.
type grpcServer struct {
	GetNetworkMode     func() *protector.NetworkMode
	GetProtocolHandler func() protocol.Handler
	GetProposalStore   func() proposal.Store
}

// AddNode proposes adding a node to the network.
func (s grpcServer) AddNode(ctx context.Context, req *pb.NodeIdentity) (*pb.Ack, error) {
	networkMode := s.GetNetworkMode()
	if networkMode == nil || networkMode.ProtectionMode != protector.PrivateWithCoordinatorMode {
		return nil, ErrNotAllowed
	}

	peerID, err := peer.IDFromBytes(req.PeerId)
	if err != nil {
		return nil, protectorpb.ErrInvalidPeerID
	}

	var peerAddr multiaddr.Multiaddr
	if len(req.PeerAddr) != 0 {
		peerAddr, err = multiaddr.NewMultiaddrBytes(req.PeerAddr)
		if err != nil {
			return nil, protectorpb.ErrInvalidPeerAddr
		}
	}

	err = s.GetProtocolHandler().AddNode(ctx, peerID, peerAddr, req.IdentityProof)
	if err != nil {
		return nil, err
	}

	return &pb.Ack{}, nil
}

// RemoveNode proposes removing a node from the network.
func (s grpcServer) RemoveNode(ctx context.Context, req *pb.NodeIdentity) (*pb.Ack, error) {
	networkMode := s.GetNetworkMode()
	if networkMode == nil || networkMode.ProtectionMode != protector.PrivateWithCoordinatorMode {
		return nil, ErrNotAllowed
	}

	peerID, err := peer.IDFromBytes(req.PeerId)
	if err != nil {
		return nil, protectorpb.ErrInvalidPeerID
	}

	err = s.GetProtocolHandler().RemoveNode(ctx, peerID)
	if err != nil {
		return nil, err
	}

	return &pb.Ack{}, nil
}

// Accept a proposal to add or remove a network node.
func (s grpcServer) Accept(ctx context.Context, req *pb.PeerID) (*pb.Ack, error) {
	networkMode := s.GetNetworkMode()
	if networkMode == nil || networkMode.ProtectionMode != protector.PrivateWithCoordinatorMode {
		return nil, ErrNotAllowed
	}

	peerID, err := peer.IDFromBytes(req.PeerId)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	err = s.GetProtocolHandler().Accept(ctx, peerID)
	if err != nil {
		return nil, err
	}

	return &pb.Ack{}, nil
}

// Reject a proposal to add or remove a network node.
func (s grpcServer) Reject(ctx context.Context, req *pb.PeerID) (*pb.Ack, error) {
	networkMode := s.GetNetworkMode()
	if networkMode == nil || networkMode.ProtectionMode != protector.PrivateWithCoordinatorMode {
		return nil, ErrNotAllowed
	}

	peerID, err := peer.IDFromBytes(req.PeerId)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	err = s.GetProtocolHandler().Reject(ctx, peerID)
	if err != nil {
		return nil, err
	}

	return &pb.Ack{}, nil
}

// List pending proposals to add or remove a network node.
func (s grpcServer) List(req *pb.Filter, ss grpc.Bootstrap_ListServer) error {
	ctx := ss.Context()

	networkMode := s.GetNetworkMode()
	if networkMode == nil || networkMode.ProtectionMode != protector.PrivateWithCoordinatorMode {
		return ErrNotAllowed
	}

	pending, err := s.GetProposalStore().List(ctx)
	if err != nil {
		return err
	}

	for _, r := range pending {
		prop := r.ToUpdateProposal()
		err = ss.Send(prop)
		if err != nil {
			return err
		}
	}

	return nil
}

// Complete the network bootstrap phase.
func (s grpcServer) Complete(ctx context.Context, req *pb.CompleteReq) (*pb.Ack, error) {
	networkMode := s.GetNetworkMode()
	if networkMode == nil ||
		networkMode.ProtectionMode != protector.PrivateWithCoordinatorMode ||
		!networkMode.IsCoordinator {
		return nil, ErrNotAllowed
	}

	err := s.GetProtocolHandler().CompleteBootstrap(ctx)
	if err != nil {
		return nil, err
	}

	return &pb.Ack{}, nil
}
