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

package bootstrap

import (
	"context"

	"github.com/pkg/errors"
	"github.com/stratumn/alice/core/protector"
	protocol "github.com/stratumn/alice/core/protocol/bootstrap"
	pb "github.com/stratumn/alice/grpc/bootstrap"
	protectorpb "github.com/stratumn/alice/pb/protector"

	"gx/ipfs/QmWWQ2Txc2c6tqjsBpzg5Ar652cHPGNsQQp2SejkNmkUMb/go-multiaddr"
	"gx/ipfs/QmcJukH2sAFjY3HdBKq35WDzWoL3UUu2gt9wdfqZTUyM74/go-libp2p-peer"
)

// grpcServer is a gRPC server for the bootstrap service.
type grpcServer struct {
	GetNetworkMode     func() *protector.NetworkMode
	GetProtocolHandler func() protocol.Handler
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
