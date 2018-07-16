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

	"github.com/pkg/errors"
	pb "github.com/stratumn/go-indigonode/core/app/host/grpc"
	"github.com/stratumn/go-indigonode/core/p2p"

	ma "gx/ipfs/QmWWQ2Txc2c6tqjsBpzg5Ar652cHPGNsQQp2SejkNmkUMb/go-multiaddr"
	"gx/ipfs/QmcJukH2sAFjY3HdBKq35WDzWoL3UUu2gt9wdfqZTUyM74/go-libp2p-peer"
	pstore "gx/ipfs/QmdeiKhUy1TVGBaKxt7y1QmBDLBdisSrLJ1x58Eoj4PXUh/go-libp2p-peerstore"
)

// grpcServer is a gRPC server for the host service.
type grpcServer struct {
	GetHost func() *p2p.Host
}

// ID returns the ID of the host.
func (s grpcServer) ID(ctx context.Context, req *pb.IdReq) (*pb.PeerId, error) {
	host := s.GetHost()
	if host == nil {
		return nil, errors.WithStack(ErrUnavailable)
	}

	return &pb.PeerId{Id: []byte(host.ID())}, nil
}

// Addresses lists all the host's addresses.
func (s grpcServer) Addresses(req *pb.AddressesReq, ss pb.Host_AddressesServer) error {
	host := s.GetHost()
	if host == nil {
		return errors.WithStack(ErrUnavailable)
	}

	for _, addr := range host.Addrs() {
		if err := ss.Send(&pb.Address{Address: addr.Bytes()}); err != nil {
			return errors.WithStack(err)
		}
	}

	return nil
}

// PeerAddresses lists a peer's known addresses.
func (s grpcServer) PeerAddresses(req *pb.PeerAddressesReq, ss pb.Host_PeerAddressesServer) error {
	host := s.GetHost()
	if host == nil {
		return errors.WithStack(ErrUnavailable)
	}

	pid, err := peer.IDFromBytes(req.PeerId)
	if err != nil {
		return errors.WithStack(err)
	}

	addrs := host.Peerstore().Addrs(pid)
	for _, addr := range addrs {
		err := ss.Send(&pb.Address{
			Address: addr.Bytes(),
		})
		if err != nil {
			return errors.WithStack(err)
		}
	}

	return nil
}

// AddPeerAddress saves a new address for the given peer.
func (s grpcServer) AddPeerAddress(ctx context.Context, req *pb.AddPeerAddressReq) (*pb.PeerId, error) {
	host := s.GetHost()
	if host == nil {
		return nil, errors.WithStack(ErrUnavailable)
	}

	pid, err := peer.IDFromBytes(req.PeerId)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	addr, err := ma.NewMultiaddrBytes(req.Address)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	host.Peerstore().AddAddr(pid, addr, pstore.PermanentAddrTTL)

	return &pb.PeerId{Id: []byte(pid)}, nil
}

// Connect ensures there is a connection to the peer's address.
func (s grpcServer) Connect(req *pb.ConnectReq, ss pb.Host_ConnectServer) error {
	host := s.GetHost()
	if host == nil {
		return errors.WithStack(ErrUnavailable)
	}

	addr, err := ma.NewMultiaddrBytes(req.Address)
	if err != nil {
		return errors.WithStack(err)
	}

	pi, err := pstore.InfoFromP2pAddr(addr)
	if err != nil {
		return errors.WithStack(err)
	}

	err = host.Connect(ss.Context(), *pi)
	if err != nil {
		return err
	}

	for _, conn := range host.Network().ConnsToPeer(pi.ID) {
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
