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
