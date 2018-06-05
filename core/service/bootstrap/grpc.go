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

	pb "github.com/stratumn/alice/grpc/bootstrap"
)

// grpcServer is a gRPC server for the bootstrap service.
type grpcServer struct {
}

// AddNode proposes adding a node to the network.
func (s grpcServer) AddNode(ctx context.Context, req *pb.NodeIdentity) (*pb.Ack, error) {
	return nil, nil
}

// Accept a proposal to add or remove a network node.
func (s grpcServer) Accept(ctx context.Context, req *pb.PeerID) (*pb.Ack, error) {
	return nil, nil
}
