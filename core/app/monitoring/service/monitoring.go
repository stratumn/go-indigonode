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

// Package service defines a service to configure monitoring for your Indigo
// Node.
// Metrics are collected and can be exposed to a Prometheus server.
// Traces are collected and can be exported to a tracing agent
// (Jaeger or Stackdriver).
package service

import (
	"context"
	"time"

	"github.com/pkg/errors"
	chat "github.com/stratumn/go-indigonode/app/chat/protocol"
	indigofossilizer "github.com/stratumn/go-indigonode/app/indigo/protocol/fossilizer"
	indigostore "github.com/stratumn/go-indigonode/app/indigo/protocol/store"
	bootstrap "github.com/stratumn/go-indigonode/core/app/bootstrap/protocol"
	grpcapi "github.com/stratumn/go-indigonode/core/app/grpcapi/service"
	pb "github.com/stratumn/go-indigonode/core/app/monitoring/grpc"
	"github.com/stratumn/go-indigonode/core/p2p"

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
	p2p.Connections,
	p2p.Peers,
	p2p.StreamsIn,
	p2p.StreamsOut,
	p2p.StreamsErr,
	p2p.Latency,
	chat.MessageReceived,
	chat.MessageSent,
	chat.MessageError,
	grpcapi.RequestReceived,
	grpcapi.RequestError,
	grpcapi.RequestDuration,
	bootstrap.Participants,
	indigofossilizer.Fossils,
	indigostore.SegmentsCreated,
	indigostore.SegmentsReceived,
	indigostore.InvalidSegments,
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
