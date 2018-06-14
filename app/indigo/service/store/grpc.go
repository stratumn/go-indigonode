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
	rpcpb "github.com/stratumn/alice/app/indigo/grpc/store"
	pb "github.com/stratumn/alice/app/indigo/pb/store"
	"github.com/stratumn/go-indigocore/cs"
	indigostore "github.com/stratumn/go-indigocore/store"
	"github.com/stratumn/go-indigocore/types"
)

var (
	// ErrNotFound is returned when no segment matched the request.
	ErrNotFound = errors.New("segment not found")
)

// grpcServer is a gRPC server for the indigo service.
type grpcServer struct {
	DoGetInfo      func() (interface{}, error)
	DoCreateLink   func(ctx context.Context, link *cs.Link) (*types.Bytes32, error)
	DoGetSegment   func(ctx context.Context, linkHash *types.Bytes32) (*cs.Segment, error)
	DoFindSegments func(ctx context.Context, filter *indigostore.SegmentFilter) (cs.SegmentSlice, error)
	DoGetMapIDs    func(ctx context.Context, filter *indigostore.MapFilter) ([]string, error)
	DoAddEvidence  func(ctx context.Context, linkHash *types.Bytes32, evidence *cs.Evidence) error
	DoGetEvidences func(ctx context.Context, linkHash *types.Bytes32) (*cs.Evidences, error)
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
func (s grpcServer) CreateLink(ctx context.Context, link *pb.Link) (*pb.LinkHash, error) {
	l, err := link.ToLink()
	if err != nil {
		return nil, err
	}

	lh, err := s.DoCreateLink(ctx, l)
	if err != nil {
		return nil, err
	}

	return pb.FromLinkHash(lh), nil
}

// GetSegment looks up a segment in the Indigo Store.
func (s grpcServer) GetSegment(ctx context.Context, req *pb.LinkHash) (*pb.Segment, error) {
	lh, err := req.ToLinkHash()
	if err != nil {
		return nil, err
	}

	seg, err := s.DoGetSegment(ctx, lh)
	if err != nil {
		return nil, err
	}

	if seg == nil {
		return nil, ErrNotFound
	}

	return pb.FromSegment(seg)
}

// FindSegments finds segments in the Indigo Store.
func (s grpcServer) FindSegments(ctx context.Context, req *rpcpb.SegmentFilter) (*pb.Segments, error) {
	filter, err := req.ToSegmentFilter()
	if err != nil {
		return nil, err
	}

	segments, err := s.DoFindSegments(ctx, filter)
	if err != nil {
		return nil, err
	}

	return pb.FromSegments(segments)
}

// GetMapIDs finds map IDs in the Indigo Store.
func (s grpcServer) GetMapIDs(ctx context.Context, req *rpcpb.MapFilter) (*rpcpb.MapIDs, error) {
	filter, err := req.ToMapFilter()
	if err != nil {
		return nil, err
	}

	mapIDs, err := s.DoGetMapIDs(ctx, filter)
	if err != nil {
		return nil, err
	}

	return rpcpb.FromMapIDs(mapIDs)
}

// AddEvidence adds evidence to the Indigo Store.
func (s grpcServer) AddEvidence(ctx context.Context, req *rpcpb.AddEvidenceReq) (*rpcpb.AddEvidenceResp, error) {
	lh, e, err := req.ToAddEvidenceParams()
	if err != nil {
		return nil, err
	}

	err = s.DoAddEvidence(ctx, lh, e)
	if err != nil {
		return nil, err
	}

	return &rpcpb.AddEvidenceResp{}, nil
}

// GetEvidences finds evidences in the Indigo Store.
func (s grpcServer) GetEvidences(ctx context.Context, req *pb.LinkHash) (*pb.Evidences, error) {
	lh, err := req.ToLinkHash()
	if err != nil {
		return nil, err
	}

	evidences, err := s.DoGetEvidences(ctx, lh)
	if err != nil {
		return nil, err
	}

	return pb.FromEvidences(*evidences)
}
