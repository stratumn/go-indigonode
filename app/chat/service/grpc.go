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
	pb "github.com/stratumn/go-indigonode/app/chat/grpc"

	peer "gx/ipfs/QmQsErDt8Qgw1XrsXf2BpEzDgGWtB1YLsTAARBup5b6B9W/go-libp2p-peer"
	pstore "gx/ipfs/Qmda4cPRvSRyox3SqgJN6DfSZGU5TtHufPTp9uXjFj71X6/go-libp2p-peerstore"
)

// grpcServer is a gRPC server for the chat service.
type grpcServer struct {
	Connect        func(context.Context, pstore.PeerInfo) error
	Send           func(context.Context, peer.ID, string) error
	GetPeerHistory func(peer.ID) (PeerHistory, error)
}

// Message sends a message to the specified peer.
func (s grpcServer) Message(ctx context.Context, req *pb.ChatMessage) (response *pb.Ack, err error) {
	response = &pb.Ack{}
	pid, err := peer.IDFromBytes(req.PeerId)
	if err != nil {
		err = errors.WithStack(err)
		return
	}

	pi := pstore.PeerInfo{ID: pid}

	// Make sure there is a connection to the peer.
	if err = s.Connect(ctx, pi); err != nil {
		return
	}

	if err = s.Send(ctx, pid, req.Message); err != nil {
		return
	}

	return
}

func (s grpcServer) GetHistory(req *pb.HistoryReq, ss pb.Chat_GetHistoryServer) error {
	id, err := peer.IDFromBytes(req.PeerId)
	if err != nil {
		return errors.WithStack(err)
	}

	msgs, err := s.GetPeerHistory(id)
	if err != nil {
		return errors.WithStack(err)
	}

	for _, msg := range msgs {
		err := ss.Send(&msg)
		if err != nil {
			return errors.WithStack(err)
		}
	}
	return nil
}
