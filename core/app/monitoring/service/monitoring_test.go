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
	"fmt"
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/stratumn/go-node/core/manager/testservice"
	"github.com/stratumn/go-node/core/netutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testService(ctx context.Context, t *testing.T) *Service {
	serv := &Service{}
	config := serv.Config().(Config)
	config.MetricsExporter = PrometheusExporter
	config.PrometheusConfig = &PrometheusConfig{
		Endpoint: fmt.Sprintf("/ip4/127.0.0.1/tcp/%d", netutil.RandomPort()),
	}

	require.NoError(t, serv.SetConfig(config), "serv.SetConfig(config)")

	return serv
}

func TestService_strings(t *testing.T) {
	testservice.CheckStrings(t, &Service{})
}

func TestService_Run(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	serv := testService(ctx, t)
	testservice.TestRun(ctx, t, serv, time.Second)
}

func TestService_Expose(t *testing.T) {
	serv := Service{}

	err := serv.SetConfig(Config{Interval: "42s"})
	require.NoError(t, err)

	exposed := serv.Expose()
	interval, ok := exposed.(time.Duration)
	require.True(t, ok)
	assert.Equal(t, float64(42), interval.Seconds())
}

func TestService_SetConfig(t *testing.T) {
	errAny := errors.New("any error")

	tests := []struct {
		name string
		set  func(*Config)
		err  error
	}{{
		"invalid interval",
		func(c *Config) { c.Interval = "1" },
		errAny,
	}, {
		"invalid sampling ratio",
		func(c *Config) { c.TraceSamplingRatio = 1.5 },
		errAny,
	}, {
		"invalid Prometheus endpoint",
		func(c *Config) {
			c.MetricsExporter = PrometheusExporter
			c.PrometheusConfig = &PrometheusConfig{Endpoint: "http://example.com"}
		},
		errAny,
	}, {
		"invalid jaeger endpoint",
		func(c *Config) {
			c.TraceExporter = JaegerExporter
			c.JaegerConfig = &JaegerConfig{Endpoint: "http://example.com"}
		},
		errAny,
	}, {
		"missing stackdriver project id",
		func(c *Config) {
			c.MetricsExporter = StackdriverExporter
			c.StackdriverConfig = &StackdriverConfig{}
		},
		errAny,
	}, {
		"invalid metrics exporter",
		func(c *Config) { c.MetricsExporter = "zipkin" },
		errAny,
	}, {
		"invalid trace exporter",
		func(c *Config) { c.TraceExporter = "zipkin" },
		errAny,
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			serv := Service{}
			config := serv.Config().(Config)
			// Disable default prometheus exporter - no need to setup
			// an http server for each test
			config.MetricsExporter = ""
			config.PrometheusConfig = nil

			tt.set(&config)

			err := errors.Cause(serv.SetConfig(config))
			switch {
			case err != nil && tt.err == errAny:
			case err != tt.err:
				assert.Equal(t, tt.err, err)
			}
		})
	}
}
