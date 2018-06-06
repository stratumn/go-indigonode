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
	pb "github.com/stratumn/alice/app/chat/grpc"

	peer "gx/ipfs/QmcJukH2sAFjY3HdBKq35WDzWoL3UUu2gt9wdfqZTUyM74/go-libp2p-peer"
	pstore "gx/ipfs/QmdeiKhUy1TVGBaKxt7y1QmBDLBdisSrLJ1x58Eoj4PXUh/go-libp2p-peerstore"
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
