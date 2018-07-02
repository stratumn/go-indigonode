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

// Package service defines a service that collects metrics and can expose them
// to a Prometheus server.
//
// It exposes the type Metrics which can be used by other services to send
// metrics.
package service

import (
	"context"
	"time"

	"github.com/pkg/errors"
	pb "github.com/stratumn/go-indigonode/core/app/metrics/grpc"
	"github.com/stratumn/go-indigonode/core/httputil"
	"github.com/stratumn/go-indigonode/core/p2p"

	"go.opencensus.io/exporter/prometheus"
	"go.opencensus.io/stats/view"
	"google.golang.org/grpc"

	ma "gx/ipfs/QmWWQ2Txc2c6tqjsBpzg5Ar652cHPGNsQQp2SejkNmkUMb/go-multiaddr"
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
	p2p.Latency,
}

// Service is the Metrics service.
type Service struct {
	config   *Config
	interval time.Duration
}

// Config contains configuration options for the Metrics service.
type Config struct {
	// Interval is the interval between updates of periodic stats.
	Interval string `toml:"interval" comment:"Interval between updates of periodic stats."`

	// PrometheusEndpoint is the address of the endpoint to expose
	// Prometheus metrics.
	PrometheusEndpoint string `toml:"prometheus_endpoint" comment:"Address of the endpoint to expose Prometheus metrics (blank = disabled)."`
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
	}
}

// SetConfig configures the service.
func (s *Service) SetConfig(config interface{}) error {
	conf := config.(Config)

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
	exporter, err := prometheus.NewExporter(prometheus.Options{})
	if err != nil {
		return err
	}

	view.RegisterExporter(exporter)

	promCtx, cancelProm := context.WithCancel(ctx)
	defer cancelProm()

	promDone := make(chan error, 1)
	if s.config.PrometheusEndpoint != "" {
		go func() {
			promDone <- httputil.StartServer(
				promCtx,
				s.config.PrometheusEndpoint,
				exporter)
		}()
	}

	if err := view.Register(views...); err != nil {
		return err
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

	view.Unregister(views...)
	view.UnregisterExporter(exporter)

	if err == nil {
		return errors.WithStack(ctx.Err())
	}

	return err
}

// AddToGRPCServer adds the service to a gRPC server.
func (s *Service) AddToGRPCServer(gs *grpc.Server) {
	pb.RegisterMetricsServer(gs, grpcServer{})
}
