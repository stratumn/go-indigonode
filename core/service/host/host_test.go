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

package host

import (
	"context"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/pkg/errors"
	"github.com/stratumn/alice/core/manager/testservice"
	"github.com/stratumn/alice/core/service/metrics"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	ifconnmgr "gx/ipfs/QmSAJm4QdTJ3EGF2cvgNcQyXTEbxqWSW1x4kCVV1aJQUQr/go-libp2p-interface-connmgr"
	testutil "gx/ipfs/QmZTcPxK6VqrwY94JpKZPvEqAZ6tEr1rLrpcqJbbRZbg2V/go-libp2p-netutil"
)

func testService(ctx context.Context, t *testing.T) *Service {
	serv := &Service{}
	config := serv.Config().(Config)
	config.ConnectionManager = ""
	config.Metrics = ""

	require.NoError(t, serv.SetConfig(config), "serv.SetConfig(config)")

	deps := map[string]interface{}{
		"swarm": testutil.GenSwarmNetwork(t, ctx),
	}

	require.NoError(t, serv.Plug(deps), "serv.Plug(deps)")

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

	assert.NotNil(t, exposed, "exposed type")
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
		"invalid negotiation timeout",
		func(c *Config) { c.NegotiationTimeout = "1" },
		errAny,
	}, {
		"invalid address netmasks",
		func(c *Config) { c.AddressesNetmasks = []string{"http://example.com"} },
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

func TestService_Needs(t *testing.T) {
	tests := []struct {
		name  string
		set   func(*Config)
		needs []string
	}{{
		"network",
		func(c *Config) { c.Network = "myswarm" },
		[]string{"connmgr", "metrics", "myswarm"},
	}, {
		"connmgr",
		func(c *Config) { c.ConnectionManager = "myconnmgr" },
		[]string{"myconnmgr", "metrics", "swarm"},
	}, {
		"metrics",
		func(c *Config) { c.Metrics = "mymetrics" },
		[]string{"connmgr", "mymetrics", "swarm"},
	}}

	toSet := func(keys []string) map[string]struct{} {
		set := map[string]struct{}{}
		for _, v := range keys {
			set[v] = struct{}{}
		}

		return set
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			serv := Service{}
			config := serv.Config().(Config)
			tt.set(&config)

			require.NoError(t, serv.SetConfig(config), "serv.SetConfig(config)")
			assert.Equal(t, toSet(tt.needs), serv.Needs())
		})
	}
}

func TestService_Plug(t *testing.T) {
	errAny := errors.New("any error")

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tests := []struct {
		name string
		set  func(*Config)
		deps map[string]interface{}
		err  error
	}{{
		"valid network",
		func(c *Config) { c.Network = "myswarm" },
		map[string]interface{}{
			"myswarm": testutil.GenSwarmNetwork(t, context.Background()),
		},
		nil,
	}, {
		"invalid network",
		func(c *Config) { c.Network = "myswarm" },
		map[string]interface{}{
			"myswarm": struct{}{},
		},
		ErrNotNetwork,
	}, {
		"valid connmgr",
		func(c *Config) { c.ConnectionManager = "myconnmgr" },
		map[string]interface{}{
			"myconnmgr": ifconnmgr.NullConnMgr{},
			"swarm":     testutil.GenSwarmNetwork(t, context.Background()),
		},
		nil,
	}, {
		"invalid connmgr",
		func(c *Config) { c.ConnectionManager = "myconnmgr" },
		map[string]interface{}{
			"myconnmgr": struct{}{},
			"swarm":     testutil.GenSwarmNetwork(t, context.Background()),
		},
		ErrNotConnManager,
	}, {
		"valid metrics",
		func(c *Config) { c.Metrics = "mymetrics" },
		map[string]interface{}{
			"mymetrics": &metrics.Metrics{},
			"swarm":     testutil.GenSwarmNetwork(t, context.Background()),
		},
		nil,
	}, {
		"invalid metrics",
		func(c *Config) { c.Metrics = "mymetrics" },
		map[string]interface{}{
			"mymetrics": struct{}{},
			"swarm":     testutil.GenSwarmNetwork(t, context.Background()),
		},
		ErrNotMetrics,
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			serv := Service{}
			config := serv.Config().(Config)
			config.ConnectionManager = ""
			config.Metrics = ""
			tt.set(&config)

			require.NoError(t, serv.SetConfig(config), "serv.SetConfig(config)")

			err := errors.Cause(serv.Plug(tt.deps))
			switch {
			case err != nil && tt.err == errAny:
			case err != tt.err:
				assert.Equal(t, tt.err, err)
			}
		})
	}
}
