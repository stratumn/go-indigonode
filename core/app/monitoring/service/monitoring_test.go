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
	"fmt"
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/stratumn/go-indigonode/core/manager/testservice"
	"github.com/stratumn/go-indigonode/core/netutil"
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
