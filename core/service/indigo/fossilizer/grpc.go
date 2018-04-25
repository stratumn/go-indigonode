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

package fossilizer

import (
	"context"

	rpcpb "github.com/stratumn/alice/grpc/indigo/fossilizer"
)

// grpcServer is a gRPC server for the indigo service.
type grpcServer struct {
	DoGetInfo   func() (interface{}, error)
	DoFossilize func(ctx context.Context, data []byte, meta []byte) error
}

// GetInfo returns information about the indigo service.
func (s grpcServer) GetInfo(ctx context.Context, req *rpcpb.InfoReq) (*rpcpb.InfoResp, error) {
	return nil, nil
}

// GetInfo returns information about the indigo service.
func (s grpcServer) Fossilize(ctx context.Context, req *rpcpb.FossilizeReq) (*rpcpb.FossilizeResp, error) {
	return nil, nil
}
