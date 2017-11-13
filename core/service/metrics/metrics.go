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
	"google.golang.org/grpc"

	metrics "gx/ipfs/QmQbh3Rb7KM37As3vkHYnEFnzkVXNCP8EYGtHz6g2fXk14/go-libp2p-metrics"
)

var (
	// ErrUnavailable is returned from gRPC methods when the service is not
	// available.
	ErrUnavailable = errors.New("the service is not available")
)

// Service is the Metrics service.
type Service struct {
	bwc metrics.Reporter
}

// ID returns the unique identifier of the service.
func (s *Service) ID() string {
	return "metrics"
}

// Name returns the human friendly name of the service.
func (s *Service) Name() string {
	return "Metrics"
}

// Desc returns a description of what the service does.
func (s *Service) Desc() string {
	return "Collects metrics."
}

// Expose exposes the bandwidth reporter to other services.
//
// It exposes the type:
//	github.com/libp2p/go-libp2p-metrics/Reporter
func (s *Service) Expose() interface{} {
	return s.bwc
}

// Run starts the service.
func (s *Service) Run(ctx context.Context, running, stopping chan struct{}) error {
	// We can reuse the counter to accumulate stats after a restart of the
	// service.
	if s.bwc == nil {
		s.bwc = metrics.NewBandwidthCounter()
	}

	running <- struct{}{}
	<-ctx.Done()
	stopping <- struct{}{}

	return errors.WithStack(ctx.Err())
}

// AddToGRPCServer adds the service to a gRPC server.
func (s *Service) AddToGRPCServer(gs *grpc.Server) {
	pb.RegisterMetricsServer(gs, grpcServer{s})
}
