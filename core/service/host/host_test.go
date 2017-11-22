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

package host

import (
	"context"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/pkg/errors"
	"github.com/stratumn/alice/core/manager/testservice"
	"github.com/stratumn/alice/core/service/metrics"

	testutil "gx/ipfs/QmQGX417WoxKxDJeHqouMEmmH4G1RCENNSzkZYHrXy3Xb3/go-libp2p-netutil"
	ifconnmgr "gx/ipfs/QmYkCrTwivapqdB3JbwvwvxymseahVkcm46ThRMAA24zCr/go-libp2p-interface-connmgr"
)

func testService(ctx context.Context, t *testing.T) *Service {
	serv := &Service{}
	config := serv.Config().(Config)
	config.ConnectionManager = ""
	config.Metrics = ""

	if err := serv.SetConfig(config); err != nil {
		t.Fatalf("serv.SetConfig(config): error: %s", err)
	}

	deps := map[string]interface{}{
		"swarm": testutil.GenSwarmNetwork(t, ctx),
	}

	if err := serv.Plug(deps); err != nil {
		t.Fatalf("serv.Plug(deps): error: %s", err)
	}

	return serv
}

func TestService_Expose(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	serv := testService(ctx, t)
	exposed := testservice.Expose(ctx, t, serv, time.Second)

	_, ok := exposed.(*Host)
	if got, want := ok, true; got != want {
		t.Errorf("ok = %v want %v", got, want)
	}
}

func TestService_Run(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	serv := testService(ctx, t)
	testservice.TestRun(ctx, t, serv, time.Second)
}

func TestService_SetConfig(t *testing.T) {
	errAny := errors.New("any error")

	tt := []struct {
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

	for _, test := range tt {
		serv := Service{}
		config := serv.Config().(Config)
		test.set(&config)

		err := errors.Cause(serv.SetConfig(config))
		switch {
		case err != nil && test.err == errAny:
		case err != test.err:
			t.Errorf("%s: err = %v want %v", test.name, err, test.err)
		}
	}
}

func TestService_Needs(t *testing.T) {
	tt := []struct {
		name  string
		set   func(*Config)
		needs []string
	}{{
		"network",
		func(c *Config) { c.Network = "myswarm" },
		[]string{"myswarm"},
	}, {
		"connmgr",
		func(c *Config) { c.ConnectionManager = "myconnmgr" },
		[]string{"myconnmgr"},
	}, {
		"metrics",
		func(c *Config) { c.Metrics = "mymetrics" },
		[]string{"mymetrics"},
	}}

	for _, test := range tt {
		serv := Service{}
		config := serv.Config().(Config)
		test.set(&config)

		if err := serv.SetConfig(config); err != nil {
			t.Errorf("%s: serv.SetConfig(config): error: %s", test.name, err)
			continue
		}

		needs := serv.Needs()
		for _, n := range test.needs {
			if _, ok := needs[n]; !ok {
				t.Errorf("%s: needs[%q] = nil want struct{}{}", test.name, n)
			}
		}
	}
}

func TestService_Plug(t *testing.T) {
	errAny := errors.New("any error")

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tt := []struct {
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

	for _, test := range tt {
		serv := Service{}
		config := serv.Config().(Config)
		config.ConnectionManager = ""
		config.Metrics = ""
		test.set(&config)

		if err := serv.SetConfig(config); err != nil {
			t.Errorf("%s: serv.SetConfig(config): error: %s", test.name, err)
			continue
		}

		err := errors.Cause(serv.Plug(test.deps))
		switch {
		case err != nil && test.err == errAny:
		case err != test.err:
			t.Errorf("%s: err = %v want %v", test.name, err, test.err)
		}
	}
}
