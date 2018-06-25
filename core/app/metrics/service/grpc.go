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
	pb "github.com/stratumn/go-indigonode/core/app/metrics/grpc"

	metrics "gx/ipfs/QmVvu4bS5QLfS19ePkp5Wgzn2ZUma5oXTT9BgDFyQLxUZF/go-libp2p-metrics"
	protocol "gx/ipfs/QmZNkThpqfVXs9GNbexPrfBbXSLNYeKrE7jwFM2oqHbyqN/go-libp2p-protocol"
	peer "gx/ipfs/QmcJukH2sAFjY3HdBKq35WDzWoL3UUu2gt9wdfqZTUyM74/go-libp2p-peer"
)

// grpcServer is a gRPC server for the metrics service.
type grpcServer struct {
	GetMetrics func() *Metrics
}

// Bandwidth reports bandwith usage.
func (s grpcServer) Bandwidth(ctx context.Context, req *pb.BandwidthReq) (*pb.BandwidthStats, error) {
	mtrx := s.GetMetrics()
	if mtrx == nil {
		return nil, errors.WithStack(ErrUnavailable)
	}

	var stats metrics.Stats

	if req.PeerId != nil {
		pid, err := peer.IDFromBytes(req.PeerId)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		stats = mtrx.GetBandwidthForPeer(pid)
	} else if req.ProtocolId != "" {
		stats = mtrx.GetBandwidthForProtocol(protocol.ID(req.ProtocolId))
	} else {
		stats = mtrx.GetBandwidthTotals()
	}

	return &pb.BandwidthStats{
		TotalIn:  uint64(stats.TotalIn),
		TotalOut: uint64(stats.TotalOut),
		RateIn:   uint64(stats.RateIn),
		RateOut:  uint64(stats.RateOut),
	}, nil
}
