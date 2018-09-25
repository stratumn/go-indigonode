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
	"time"

	"github.com/pkg/errors"
	"github.com/stratumn/go-indigonode/core/httputil"

	"contrib.go.opencensus.io/exporter/stackdriver"
	"go.opencensus.io/exporter/jaeger"
	"go.opencensus.io/exporter/prometheus"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/trace"

	manet "gx/ipfs/QmV6FjemM1K8oXjrvuq3wuVWWoU2TLDPmNnKrxHzY3v6Ai/go-multiaddr-net"
	ma "gx/ipfs/QmYmsdtJ3HsodkePE3eU3TsCaP2YvPZJ4LoXnNkDE5Tpt7/go-multiaddr"
)

// Available exporters.
const (
	PrometheusExporter  = "prometheus"
	JaegerExporter      = "jaeger"
	StackdriverExporter = "stackdriver"
)

// Errors used by the configuration component.
var (
	ErrInvalidRatio           = errors.New("invalid ratio (must be in [0;1])")
	ErrInvalidMetricsExporter = errors.New("metrics exporter should be 'prometheus' or 'stackdriver'")
	ErrInvalidTraceExporter   = errors.New("trace exporter should be 'jaeger' or 'stackdriver'")
	ErrMissingExporterConfig  = errors.New("missing exporter configuration section")
	ErrMissingProjectID       = errors.New("missing stackdriver project id")
)

// Config contains configuration options for the Monitoring service.
type Config struct {
	// Interval is the interval between updates of periodic stats.
	Interval string `toml:"interval" comment:"Interval between updates of periodic stats."`

	// TraceSamplingRatio is the fraction of traces to record.
	TraceSamplingRatio float64 `toml:"trace_sampling_ratio" comment:"Fraction of traces to record."`

	// MetricsExporter is the name of the metrics exporter.
	MetricsExporter string `toml:"metrics_exporter" comment:"Name of the metrics exporter (prometheus or stackdriver). Leave empty to disable metrics."`

	// TraceExporter is the name of the trace exporter.
	TraceExporter string `toml:"trace_exporter" comment:"Name of the trace exporter (jaeger or stackdriver). Leave empty to disable tracing."`

	// JaegerConfig options (if enabled).
	JaegerConfig *JaegerConfig `toml:"jaeger" comment:"Jaeger configuration options (if enabled)."`

	// PrometheusConfig options (if enabled).
	PrometheusConfig *PrometheusConfig `toml:"prometheus" comment:"Prometheus configuration options (if enabled)."`

	// StackdriverConfig options (if enabled).
	StackdriverConfig *StackdriverConfig `toml:"stackdriver" comment:"Stackdriver configuration options (if enabled)."`
}

// StackdriverConfig contains configuration options for Stackdriver.
type StackdriverConfig struct {
	// ProjectID is the identifier of the Stackdriver project
	ProjectID string `toml:"project_id" comment:"Identifier of the Stackdriver project."`
}

// JaegerConfig contains configuration options for Jaeger (tracing).
type JaegerConfig struct {
	// Endpoint is the address of the Jaeger agent to collect traces.
	Endpoint string `toml:"endpoint" comment:"Address of the Jaeger agent to collect traces."`
}

// PrometheusConfig contains configuration options for Prometheus.
type PrometheusConfig struct {
	// Endpoint is the address of the endpoint to expose prometheus metrics.
	Endpoint string `toml:"endpoint" comment:"Address of the endpoint to expose Prometheus metrics."`
}

// Config returns the current service configuration or creates one with
// good default values.
func (s *Service) Config() interface{} {
	if s.config != nil {
		return *s.config
	}

	return Config{
		Interval:           "10s",
		TraceSamplingRatio: 1.0,
		MetricsExporter:    PrometheusExporter,
		TraceExporter:      "",
		JaegerConfig:       &JaegerConfig{Endpoint: "/ip4/127.0.0.1/tcp/14268"},
		PrometheusConfig:   &PrometheusConfig{Endpoint: "/ip4/127.0.0.1/tcp/8905"},
		StackdriverConfig:  &StackdriverConfig{ProjectID: "your-stackdriver-project-id"},
	}
}

// SetConfig configures the service.
func (s *Service) SetConfig(config interface{}) error {
	conf := config.(Config)

	if err := conf.ValidateSamplingRatio(); err != nil {
		return err
	}

	interval, err := time.ParseDuration(conf.Interval)
	if err != nil {
		return errors.WithStack(err)
	}

	if err := conf.ValidateMetricsConfig(); err != nil {
		return err
	}

	if err := conf.ValidateTraceConfig(); err != nil {
		return err
	}

	s.config = &conf
	s.interval = interval

	return nil
}

// ValidateSamplingRatio validates the tracing sampling ratio.
func (c *Config) ValidateSamplingRatio() error {
	if c.TraceSamplingRatio < 0 || c.TraceSamplingRatio > 1.0 {
		return ErrInvalidRatio
	}

	return nil
}

// ValidateMetricsConfig validates the metrics configuration.
func (c *Config) ValidateMetricsConfig() error {
	switch c.MetricsExporter {
	case "":
		return nil
	case PrometheusExporter:
		return c.PrometheusConfig.Validate()
	case StackdriverExporter:
		return c.StackdriverConfig.Validate()
	default:
		return ErrInvalidMetricsExporter
	}
}

// ValidateTraceConfig validates the tracing configuration.
func (c *Config) ValidateTraceConfig() error {
	switch c.TraceExporter {
	case "":
		return nil
	case JaegerExporter:
		return c.JaegerConfig.Validate()
	case StackdriverExporter:
		return c.StackdriverConfig.Validate()
	default:
		return ErrInvalidTraceExporter
	}
}

// Validate the prometheus configuration section.
func (c *PrometheusConfig) Validate() error {
	if c == nil {
		return ErrMissingExporterConfig
	}

	return validateEndpoint(c.Endpoint)
}

// Validate the jaeger configuration section.
func (c *JaegerConfig) Validate() error {
	if c == nil {
		return ErrMissingExporterConfig
	}

	return validateEndpoint(c.Endpoint)
}

// Validate the stackdriver configuration section.
func (c *StackdriverConfig) Validate() error {
	if c == nil {
		return ErrMissingExporterConfig
	}

	if c.ProjectID == "" {
		return ErrMissingProjectID
	}

	return nil
}

func validateEndpoint(endpoint string) error {
	_, err := ma.NewMultiaddr(endpoint)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

// CreateMetricsExporter configures the metrics exporter.
// It returns the exporter itself and a cancel function to stop the exporter.
func (c *Config) CreateMetricsExporter(ctx context.Context, errChan chan<- error) (view.Exporter, context.CancelFunc, error) {
	var err error
	var exporter view.Exporter

	metricsCtx, cancel := context.WithCancel(ctx)

	switch c.MetricsExporter {
	case PrometheusExporter:
		exporter, err = prometheus.NewExporter(prometheus.Options{})
		if err != nil {
			break
		}

		go func() {
			errChan <- httputil.StartServer(
				metricsCtx,
				c.PrometheusConfig.Endpoint,
				exporter.(*prometheus.Exporter))
		}()
	case StackdriverExporter:
		exporter, err = stackdriver.NewExporter(stackdriver.Options{
			ProjectID: c.StackdriverConfig.ProjectID,
		})
	}

	return exporter, cancel, err
}

// CreateTraceExporter configures the trace exporter.
func (c *Config) CreateTraceExporter() (trace.Exporter, error) {
	var err error
	var exporter trace.Exporter

	switch c.TraceExporter {
	case JaegerExporter:
		jaegerMultiaddr, _ := ma.NewMultiaddr(c.JaegerConfig.Endpoint)
		jaegerEndpoint, addrErr := manet.ToNetAddr(jaegerMultiaddr)
		if addrErr != nil {
			return nil, addrErr
		}

		exporter, err = jaeger.NewExporter(jaeger.Options{
			Endpoint:    "http://" + jaegerEndpoint.String(),
			ServiceName: "indigo-node",
		})
	case StackdriverExporter:
		exporter, err = stackdriver.NewExporter(stackdriver.Options{
			ProjectID: c.StackdriverConfig.ProjectID,
		})
	}

	return exporter, err
}
