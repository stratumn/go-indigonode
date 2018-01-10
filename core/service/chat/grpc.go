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
	Connect func(context.Context, pstore.PeerInfo) error
	Send    func(context.Context, peer.ID, string) error
}

// Message sends a message to the specified peer.
func (s grpcServer) Message(ctx context.Context, req *pb.MessageReq) (response *pb.Ack, err error) {
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

	if err = s.Send(ctx, pid, req.Content); err != nil {
		return
	}

	return
}