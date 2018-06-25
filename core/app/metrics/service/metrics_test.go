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

	metrics "github.com/armon/go-metrics"
	"github.com/pkg/errors"
	"github.com/stratumn/go-indigonode/core/manager/testservice"
	"github.com/stratumn/go-indigonode/core/netutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testService(ctx context.Context, t *testing.T) *Service {
	serv := &Service{}
	config := serv.Config().(Config)
	config.PrometheusEndpoint = fmt.Sprintf("/ip4/127.0.0.1/tcp/%d", netutil.RandomPort())

	require.NoError(t, serv.SetConfig(config), "serv.SetConfig(config)")

	return serv
}

func TestService_strings(t *testing.T) {
	testservice.CheckStrings(t, &Service{})
}

func TestService_Expose(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	serv := testService(ctx, t)
	exposed := testservice.Expose(ctx, t, serv, time.Second)

	assert.IsType(t, &Metrics{}, exposed, "exposed type")
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
		"invalid Prometheus endpoint",
		func(c *Config) { c.PrometheusEndpoint = "http://example.com" },
		errAny,
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			serv := Service{}
			config := serv.Config().(Config)
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

func TestMetrics_AddPeriodicHandler(t *testing.T) {
	mtrx := newMetrics(nil, nil)

	ch := make(chan struct{}, 1)
	remove := mtrx.AddPeriodicHandler(func(metrics.MetricSink) {
		ch <- struct{}{}
	})

	ctx, cancel := context.WithCancel(context.Background())
	doneCh := make(chan struct{})
	go func() {
		mtrx.start(ctx, 100*time.Millisecond)
		close(doneCh)
	}()

	select {
	case <-time.After(time.Second):
		assert.Fail(t, "periodic function not called")
	case <-ch:
	}

	remove()

	select {
	case <-time.After(200 * time.Millisecond):
	case <-ch:
		assert.Fail(t, "periodic function called")
	}

	cancel()

	select {
	case <-time.After(time.Second):
		assert.Fail(t, "metrics did not stop")
	case <-doneCh:
	}
}
