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

package metrics

import (
	"context"

	"github.com/pkg/errors"
	pb "github.com/stratumn/alice/grpc/metrics"

	metrics "gx/ipfs/QmQbh3Rb7KM37As3vkHYnEFnzkVXNCP8EYGtHz6g2fXk14/go-libp2p-metrics"
	peer "gx/ipfs/QmXYjuNuxVzXKJCfWasQk1RqkhVLDM9jtUKhqc2WPQmFSB/go-libp2p-peer"
	protocol "gx/ipfs/QmZNkThpqfVXs9GNbexPrfBbXSLNYeKrE7jwFM2oqHbyqN/go-libp2p-protocol"
)

// grpcServer is a gRPC server for the metrics service.
type grpcServer struct {
	service *Service
}

// Bandwidth reports bandwith usage.
func (s grpcServer) Bandwidth(ctx context.Context, req *pb.BandwidthReq) (*pb.BandwidthStats, error) {
	bwc := s.service.bwc
	if bwc == nil {
		return nil, errors.WithStack(ErrUnavailable)
	}

	var stats metrics.Stats

	if req.PeerId != nil {
		pid, err := peer.IDFromBytes(req.PeerId)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		stats = bwc.GetBandwidthForPeer(pid)
	} else if req.ProtocolId != "" {
		stats = bwc.GetBandwidthForProtocol(protocol.ID(req.ProtocolId))
	} else {
		stats = bwc.GetBandwidthTotals()
	}

	return &pb.BandwidthStats{
		TotalIn:  uint64(stats.TotalIn),
		TotalOut: uint64(stats.TotalOut),
		RateIn:   uint64(stats.RateIn),
		RateOut:  uint64(stats.RateOut),
	}, nil
}
