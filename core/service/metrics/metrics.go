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
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	metrics "github.com/armon/go-metrics"
	"github.com/armon/go-metrics/prometheus"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/stratumn/alice/core/netutil"
	pb "github.com/stratumn/alice/grpc/metrics"
	"google.golang.org/grpc"

	p2pmetrics "gx/ipfs/QmQbh3Rb7KM37As3vkHYnEFnzkVXNCP8EYGtHz6g2fXk14/go-libp2p-metrics"
	logging "gx/ipfs/QmSpJByNKFX1sCsHBEp3R73FL4NF6FnQTEGyNAXHm2GS52/go-log"
	maddr "gx/ipfs/QmXY77cVe7rVRQXZZQRioukUM7aRW3BTcAgJe12MCtb3Ji/go-multiaddr"
)

var (
	// ErrUnavailable is returned from gRPC methods when the service is not
	// available.
	ErrUnavailable = errors.New("the service is not available")
)

// log is the logger for the service.
var log = logging.Logger("metrics")

// sink is the global Prometheus sink. If not global, it crashes when gauges
// with the same name are recreated after service restart.
var promSink metrics.MetricSink

// Creates the global sink.
func init() {
	sink, err := prometheus.NewPrometheusSink()
	if err != nil {
		panic(err)
	}

	promSink = sink
}

// Service is the Metrics service.
type Service struct {
	config   *Config
	interval time.Duration
	metrics  *Metrics
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
	return "Collects metrics."
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
		_, err = maddr.NewMultiaddr(conf.PrometheusEndpoint)
		if err != nil {
			return errors.WithStack(err)
		}
	}

	s.config = &conf
	s.interval = interval

	return nil
}

// Expose exposes the bandwidth reporter to other services.
//
// It exposes the type:
//	github.com/stratumn/alice/core/service/*metrics.Metrics
func (s *Service) Expose() interface{} {
	return s.metrics
}

// Run starts the service.
func (s *Service) Run(ctx context.Context, running, stopping chan struct{}) error {
	bwc := p2pmetrics.NewBandwidthCounter()

	s.metrics = newMetrics(bwc, promSink)

	metricsCtx, cancelMetrics := context.WithCancel(ctx)
	defer cancelMetrics()

	metricsDone := make(chan struct{})
	go func() {
		s.metrics.start(metricsCtx, s.interval)
		close(metricsDone)
	}()

	promCtx, cancelProm := context.WithCancel(ctx)
	defer cancelProm()

	promDone := make(chan error, 1)
	if s.config.PrometheusEndpoint != "" {
		go func() {
			promDone <- s.startProm(promCtx)
		}()
	}

	running <- struct{}{}

	var err error

	select {
	case err = <-promDone:
		stopping <- struct{}{}

	case <-ctx.Done():
		stopping <- struct{}{}
		if s.config.PrometheusEndpoint != "" {
			cancelProm()
			err = <-promDone
		}
	}

	cancelMetrics()
	<-metricsDone

	s.metrics = nil

	if err == nil {
		return errors.WithStack(ctx.Err())
	}

	return err
}

// AddToGRPCServer adds the service to a gRPC server.
func (s *Service) AddToGRPCServer(gs *grpc.Server) {
	pb.RegisterMetricsServer(gs, grpcServer{s})
}

// startProm starts exposing Prometheus metrics on an HTTP endpoint.
func (s *Service) startProm(ctx context.Context) error {
	lis, err := netutil.Listen(s.config.PrometheusEndpoint)
	if err != nil {
		return err
	}

	srv := http.Server{Handler: promHandler{ctx}}

	done := make(chan error, 1)
	go func() {
		err := srv.Serve(lis)
		if err != nil && err != http.ErrServerClosed {
			done <- err
			return
		}

		close(done)
	}()

	select {
	case err := <-done:
		return errors.WithStack(err)

	case <-ctx.Done():
		shutdownCtx, cancelShutdown := context.WithTimeout(
			context.Background(),
			time.Second,
		)
		defer cancelShutdown()

		if err := srv.Shutdown(shutdownCtx); err != nil {
			return errors.WithStack(err)
		}

		if err := <-done; err != nil {
			return errors.WithStack(err)
		}

		return errors.WithStack(ctx.Err())
	}
}

// Metrics embeds a libp2p reporter and a metrics sink.
//
// It can also be used to add handles for periodic metrics.
type Metrics struct {
	p2pmetrics.Reporter
	metrics.MetricSink

	handlersMu sync.Mutex
	handlers   map[uint64]func(metrics.MetricSink)
}

// newMetrics creates a new metrics struct.
func newMetrics(bwc p2pmetrics.Reporter, sink metrics.MetricSink) *Metrics {
	return &Metrics{
		Reporter:   bwc,
		MetricSink: sink,
		handlers:   map[uint64]func(metrics.MetricSink){},
	}
}

var metricsHandlerID = uint64(0)

// AddPeriodicHandler adds a periodic metrics handler.
//
// It returns a function that removes the handler.
func (m *Metrics) AddPeriodicHandler(handler func(metrics.MetricSink)) func() {
	m.handlersMu.Lock()
	defer m.handlersMu.Unlock()

	id := atomic.AddUint64(&metricsHandlerID, 1)
	m.handlers[id] = handler

	return func() {
		m.handlersMu.Lock()
		delete(m.handlers, id)
		m.handlersMu.Unlock()
	}
}

// start starts the periodic ticker.
func (m *Metrics) start(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	m.tick(ctx)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			m.tick(ctx)
		}
	}
}

// tick executes all the registered periodic handlers.
func (m *Metrics) tick(ctx context.Context) {
	defer log.EventBegin(ctx, "tick").Done()

	m.handlersMu.Lock()
	defer m.handlersMu.Unlock()

	for _, handler := range m.handlers {
		handler(m)
	}
}

// promHandler is an HTTP handler for the Prometheus endpoint.
//
// It logs Prometheus requests.
type promHandler struct {
	ctx context.Context
}

// ServeHTTP serves an HTTP request for Prometheus.
func (h promHandler) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	event := log.EventBegin(h.ctx, "prometheusServe")
	defer event.Done()

	promhttp.Handler().ServeHTTP(res, req)
}
