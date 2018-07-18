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

package service

import (
	"context"

	pb "github.com/stratumn/go-indigonode/core/app/monitoring/grpc"

	"go.opencensus.io/trace"
)

// grpcServer is a gRPC server for the monitoring service.
type grpcServer struct{}

// SetSamplingRatio sets the sampling ratio for traces.
func (s grpcServer) SetSamplingRatio(ctx context.Context, r *pb.SamplingRatio) (*pb.Ack, error) {
	if r == nil || r.Value < 0 || r.Value > 1.0 {
		return nil, ErrInvalidRatio
	}

	trace.ApplyConfig(trace.Config{
		DefaultSampler: trace.ProbabilitySampler(float64(r.Value)),
	})

	return &pb.Ack{}, nil
}
