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

package chat

import (
	"context"

	"github.com/pkg/errors"
	pb "github.com/stratumn/alice/grpc/chat"

	peer "gx/ipfs/QmWNY7dV54ZDYmTA1ykVdwNCqC11mpU4zSUp6XDpLTH9eG/go-libp2p-peer"
	pstore "gx/ipfs/QmYijbtjCxFEjSXaudaQAUz3LN5VKLssm8WCUsRoqzXmQR/go-libp2p-peerstore"
)

// grpcServer is a gRPC server for the chat service.
type grpcServer struct {
	Connect        func(context.Context, pstore.PeerInfo) error
	Send           func(context.Context, peer.ID, string) error
	AddListener    func() <-chan *pb.ChatMessage
	RemoveListener func(<-chan *pb.ChatMessage)
}

// Message sends a message to the specified peer.
func (s grpcServer) Message(ctx context.Context, req *pb.ChatMessage) (response *pb.Ack, err error) {
	response = &pb.Ack{}
	pid, err := peer.IDFromBytes(req.ToPeer)
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

func (s grpcServer) Listen(_ *pb.ListenReq, ss pb.Chat_ListenServer) error {
	log.Event(ss.Context(), "AddListener")

	receiveChan := s.AddListener()
	defer func() {
		// Note: it's the chat server's responsibility to close the listener channels
		s.RemoveListener(receiveChan)
		log.Event(ss.Context(), "RemoveListener")
	}()

messagePump:
	// Forward incoming messages to listeners
	for {
		select {
		case message, ok := <-receiveChan:
			if !ok {
				break messagePump
			}

			err := ss.Send(message)
			if err != nil {
				log.Errorf("Could not forward message to listeners: %v", message)
				return errors.WithStack(err)
			}
		case <-ss.Context().Done():
			break messagePump
		}
	}

	return nil
}
