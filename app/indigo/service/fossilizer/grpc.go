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

package fossilizer

import (
	"context"
	"encoding/json"

	"github.com/pkg/errors"
	rpcpb "github.com/stratumn/go-indigonode/app/indigo/grpc/fossilizer"
)

var (
	// ErrInvalidArgument is returned when the input is invalid.
	ErrInvalidArgument = errors.New("invalid argument")
)

// grpcServer is a gRPC server for the indigo service.
type grpcServer struct {
	DoGetInfo   func(ctx context.Context) (interface{}, error)
	DoFossilize func(ctx context.Context, data []byte, meta []byte) error
}

// GetInfo returns information about the indigo service.
func (s grpcServer) GetInfo(ctx context.Context, req *rpcpb.InfoReq) (*rpcpb.InfoResp, error) {
	info, err := s.DoGetInfo(ctx)
	if err != nil {
		return nil, err
	}

	infoBytes, err := json.Marshal(info)
	if err != nil {
		return nil, err
	}

	return &rpcpb.InfoResp{Data: infoBytes}, nil
}

// Fossilize requests data to be fossilized.
func (s grpcServer) Fossilize(ctx context.Context, req *rpcpb.FossilizeReq) (*rpcpb.FossilizeResp, error) {
	if req == nil {
		return nil, ErrInvalidArgument
	}
	err := s.DoFossilize(ctx, req.Data, req.Meta)
	if err != nil {
		return nil, err
	}

	return &rpcpb.FossilizeResp{}, nil
}
