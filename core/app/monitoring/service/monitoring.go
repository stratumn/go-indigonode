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
	"github.com/stratumn/go-indigonode/core/httputil"
	"github.com/stratumn/go-indigonode/core/p2p"

	"go.opencensus.io/exporter/jaeger"
	"go.opencensus.io/exporter/prometheus"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/trace"
	"google.golang.org/grpc"

	manet "gx/ipfs/QmRK2LxanhK2gZq6k6R7vk5ZoYZk8ULSSTB7FzDsMUX6CB/go-multiaddr-net"
	ma "gx/ipfs/QmWWQ2Txc2c6tqjsBpzg5Ar652cHPGNsQQp2SejkNmkUMb/go-multiaddr"
)

var (
	// ErrUnavailable is returned from gRPC methods when the service is not
	// available.
	ErrUnavailable = errors.New("the service is not available")

	// ErrInvalidRatio when an invalid ratio is provided.
	ErrInvalidRatio = errors.New("invalid ratio (must be in [0;1])")
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

// Config contains configuration options for the Monitoring service.
type Config struct {
	// TraceSamplingRatio is the fraction of traces to record.
	TraceSamplingRatio float64 `toml:"trace_sampling_ratio" comment:"Fraction of traces to record."`

	// JaegerEndpoint is the address of the endpoint of the Jaeger agent to collect traces.
	JaegerEndpoint string `toml:"jaeger_endpoint" comment:"Address of the endpoint of the Jaeger agent to collect traces (blank = disabled)."`

	// Interval is the interval between updates of periodic stats.
	Interval string `toml:"interval" comment:"Interval between updates of periodic stats."`

	// PrometheusEndpoint is the address of the endpoint to expose
	// Prometheus metrics.
	PrometheusEndpoint string `toml:"prometheus_endpoint" comment:"Address of the endpoint to expose Prometheus metrics (blank = disabled)."`
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

// Config returns the current service configuration or creates one with
// good default values.
func (s *Service) Config() interface{} {
	if s.config != nil {
		return *s.config
	}

	return Config{
		PrometheusEndpoint: "/ip4/127.0.0.1/tcp/8905",
		Interval:           "10s",
		JaegerEndpoint:     "/ip4/127.0.0.1/tcp/14268",
		TraceSamplingRatio: 1.0,
	}
}

// SetConfig configures the service.
func (s *Service) SetConfig(config interface{}) error {
	conf := config.(Config)

	if conf.TraceSamplingRatio < 0 || conf.TraceSamplingRatio > 1.0 {
		return ErrInvalidRatio
	}

	interval, err := time.ParseDuration(conf.Interval)
	if err != nil {
		return errors.WithStack(err)
	}

	if conf.PrometheusEndpoint != "" {
		_, err = ma.NewMultiaddr(conf.PrometheusEndpoint)
		if err != nil {
			return errors.WithStack(err)
		}
	}

	if conf.JaegerEndpoint != "" {
		_, err = ma.NewMultiaddr(conf.JaegerEndpoint)
		if err != nil {
			return errors.WithStack(err)
		}
	}

	s.config = &conf
	s.interval = interval

	return nil
}

// Expose doesn't need to expose anything.
// It's a pull model instead of a push one: every component should expose
// custom metric views, and we choose here which metrics we want to record.
func (s *Service) Expose() interface{} {
	return nil
}

// Run starts the service.
func (s *Service) Run(ctx context.Context, running, stopping func()) error {
	var err error
	var jaegerExporter trace.Exporter
	if s.config.JaegerEndpoint != "" {
		jaegerMultiaddr, _ := ma.NewMultiaddr(s.config.JaegerEndpoint)
		jaegerEndpoint, err := manet.ToNetAddr(jaegerMultiaddr)
		if err != nil {
			return err
		}

		jaegerExporter, err = jaeger.NewExporter(jaeger.Options{
			Endpoint:    jaegerEndpoint.String(),
			ServiceName: "indigo-node",
		})
		if err != nil {
			return err
		}

		trace.ApplyConfig(trace.Config{DefaultSampler: trace.ProbabilitySampler(s.config.TraceSamplingRatio)})
		trace.RegisterExporter(jaegerExporter)
	}

	var promExporter *prometheus.Exporter
	promDone := make(chan error, 1)
	var cancelProm context.CancelFunc
	if s.config.PrometheusEndpoint != "" {
		promExporter, err = prometheus.NewExporter(prometheus.Options{})
		if err != nil {
			return err
		}

		view.RegisterExporter(promExporter)
		if err := view.Register(views...); err != nil {
			return err
		}

		promCtx, cancel := context.WithCancel(ctx)
		defer cancel()
		cancelProm = cancel

		go func() {
			promDone <- httputil.StartServer(
				promCtx,
				s.config.PrometheusEndpoint,
				promExporter)
		}()
	}

	running()

	select {
	case err = <-promDone:
		stopping()

	case <-ctx.Done():
		stopping()
		if s.config.PrometheusEndpoint != "" {
			cancelProm()
			err = <-promDone
		}
	}

	if promExporter != nil {
		view.Unregister(views...)
		view.UnregisterExporter(promExporter)
	}

	if jaegerExporter != nil {
		trace.UnregisterExporter(jaegerExporter)
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
