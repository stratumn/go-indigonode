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

// Package service defines a service to configure monitoring for your Stratumn
// Node.
// Metrics are collected and can be exposed to a Prometheus server.
// Traces are collected and can be exported to a tracing agent
// (Jaeger or Stackdriver).
package service

import (
	"context"
	"time"

	"github.com/pkg/errors"
	chat "github.com/stratumn/go-node/app/chat/protocol"
	bootstrap "github.com/stratumn/go-node/core/app/bootstrap/protocol"
	grpcapi "github.com/stratumn/go-node/core/app/grpcapi/service"
	pb "github.com/stratumn/go-node/core/app/monitoring/grpc"
	swarm "github.com/stratumn/go-node/core/app/swarm/service"
	"github.com/stratumn/go-node/core/p2p"

	"go.opencensus.io/stats/view"
	"go.opencensus.io/trace"

	"google.golang.org/grpc"
)

var (
	// ErrUnavailable is returned from gRPC methods when the service is not
	// available.
	ErrUnavailable = errors.New("the service is not available")
)

// views registered for metrics collection.
var views = []*view.View{
	p2p.BandwidthIn,
	p2p.BandwidthOut,
	p2p.StreamsIn,
	p2p.StreamsOut,
	p2p.StreamsErr,
	swarm.Connections,
	swarm.Peers,
	swarm.Latency,
	chat.MessageReceived,
	chat.MessageSent,
	chat.MessageError,
	grpcapi.RequestReceived,
	grpcapi.RequestError,
	grpcapi.RequestDuration,
	bootstrap.Participants,
}

// Service is the Monitoring service.
type Service struct {
	config   *Config
	interval time.Duration
}

// ID returns the unique identifier of the service.
func (s *Service) ID() string {
	return "monitoring"
}

// Name returns the human friendly name of the service.
func (s *Service) Name() string {
	return "Monitoring"
}

// Desc returns a description of what the service does.
func (s *Service) Desc() string {
	return "Collects metrics and traces."
}

// Expose returns the interval at which periodic metrics should be retrieved.
// The monitoring service uses pull model instead of push: every component
// should expose custom metric views and collect them at the interval specified
// by the monitoring service, and we choose here which views we export.
func (s *Service) Expose() interface{} {
	return s.interval
}

// Run starts the service.
func (s *Service) Run(ctx context.Context, running, stopping func()) error {
	var err error
	errChan := make(chan error, 1)

	var traceExporter trace.Exporter
	if s.config.TraceExporter != "" {
		traceExporter, err = s.config.CreateTraceExporter()
		if err != nil {
			return err
		}

		trace.ApplyConfig(trace.Config{DefaultSampler: trace.ProbabilitySampler(s.config.TraceSamplingRatio)})
		trace.RegisterExporter(traceExporter)
	}

	var metricsExporter view.Exporter
	var cancelMetrics context.CancelFunc
	if s.config.MetricsExporter != "" {
		metricsExporter, cancelMetrics, err = s.config.CreateMetricsExporter(ctx, errChan)
		if err != nil {
			return err
		}

		view.RegisterExporter(metricsExporter)
		if err := view.Register(views...); err != nil {
			return err
		}
	}

	running()

	select {
	case err = <-errChan:
		stopping()
	case <-ctx.Done():
		stopping()
		if cancelMetrics != nil {
			cancelMetrics()
			err = <-errChan
		}
	}

	if metricsExporter != nil {
		view.Unregister(views...)
		view.UnregisterExporter(metricsExporter)
	}

	if traceExporter != nil {
		trace.UnregisterExporter(traceExporter)
	}

	if err == nil {
		return errors.WithStack(ctx.Err())
	}

	return err
}

// AddToGRPCServer adds the service to a gRPC server.
func (s *Service) AddToGRPCServer(gs *grpc.Server) {
	pb.RegisterMonitoringServer(gs, grpcServer{})
}
