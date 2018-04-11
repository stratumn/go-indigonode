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

package store

import (
	"context"
	"encoding/json"

	"github.com/pkg/errors"
	rpcpb "github.com/stratumn/alice/grpc/indigo/store"
	"github.com/stratumn/go-indigocore/cs"
	"github.com/stratumn/go-indigocore/types"
)

var (
	// ErrNotFound is returned when no segment matched the request.
	ErrNotFound = errors.New("segment not found")

	// ErrInvalidArgument is returned when the input is invalid.
	ErrInvalidArgument = errors.New("invalid argument")
)

// grpcServer is a gRPC server for the indigo service.
type grpcServer struct {
	DoGetInfo    func() (interface{}, error)
	DoCreateLink func(ctx context.Context, link *cs.Link) (*types.Bytes32, error)
	DoGetSegment func(ctx context.Context, linkHash *types.Bytes32) (*cs.Segment, error)
}

// GetInfo returns information about the indigo service.
func (s grpcServer) GetInfo(ctx context.Context, req *rpcpb.InfoReq) (*rpcpb.InfoResp, error) {
	info, err := s.DoGetInfo()
	if err != nil {
		return nil, err
	}

	infoBytes, err := json.Marshal(info)
	if err != nil {
		return nil, err
	}

	return &rpcpb.InfoResp{Data: infoBytes}, nil
}

// CreateLink creates a link in the Indigo Store.
func (s grpcServer) CreateLink(ctx context.Context, link *rpcpb.Link) (*rpcpb.LinkHash, error) {
	if link == nil {
		return nil, ErrInvalidArgument
	}

	var l cs.Link
	err := json.Unmarshal(link.Data, &l)
	if err != nil {
		return nil, ErrInvalidArgument
	}

	lh, err := s.DoCreateLink(ctx, &l)
	if err != nil {
		return nil, err
	}

	return &rpcpb.LinkHash{Data: lh[:]}, nil
}

// GetSegment looks up a segment in the Indigo Store.
func (s grpcServer) GetSegment(ctx context.Context, req *rpcpb.LinkHash) (*rpcpb.Segment, error) {
	if req == nil {
		return nil, ErrInvalidArgument
	}

	lh := types.NewBytes32FromBytes(req.Data)
	seg, err := s.DoGetSegment(ctx, lh)
	if err != nil {
		return nil, err
	}

	if seg == nil {
		return nil, ErrNotFound
	}

	segmentBytes, err := json.Marshal(seg)
	if err != nil {
		return nil, err
	}

	return &rpcpb.Segment{Data: segmentBytes}, nil
}
