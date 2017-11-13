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

package host

import (
	"context"

	"github.com/pkg/errors"
	pb "github.com/stratumn/alice/grpc/host"

	pstore "gx/ipfs/QmPgDWmTmuzvP7QE5zwo1TmjbJme9pmZHNujB2453jkCTr/go-libp2p-peerstore"
	maddr "gx/ipfs/QmXY77cVe7rVRQXZZQRioukUM7aRW3BTcAgJe12MCtb3Ji/go-multiaddr"
)

// grpcServer is a gRPC server for the host service.
type grpcServer struct {
	service *Service
}

// ID returns the ID of the host.
func (s grpcServer) ID(ctx context.Context, req *pb.IdReq) (*pb.HostId, error) {
	host := s.service.host
	if host == nil {
		return nil, errors.WithStack(ErrUnavailable)
	}

	return &pb.HostId{Id: []byte(host.ID())}, nil
}

// Addresses lists all the host's addresses.
func (s grpcServer) Addresses(req *pb.AddressesReq, ss pb.Host_AddressesServer) error {
	host := s.service.host
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

// Connect ensures there is a connection to the peer's address.
func (s grpcServer) Connect(req *pb.ConnectReq, ss pb.Host_ConnectServer) error {
	host := s.service.host
	if host == nil {
		return errors.WithStack(ErrUnavailable)
	}

	addr, err := maddr.NewMultiaddrBytes(req.Address)
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