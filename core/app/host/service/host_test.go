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
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/pkg/errors"
	swarm "github.com/stratumn/go-node/core/app/swarm/service"
	"github.com/stratumn/go-node/core/manager/testservice"
	"github.com/stratumn/go-node/core/p2p"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	ifconnmgr "gx/ipfs/QmWGGN1nysi1qgqto31bENwESkmZBY4YGK4sZC3qhnqhSv/go-libp2p-interface-connmgr"
	swarmtesting "gx/ipfs/QmeDpqUwwdye8ABKVMPXKuWwPVURFdqTqssbTUB39E2Nwd/go-libp2p-swarm/testing"
)

func testService(ctx context.Context, t *testing.T) *Service {
	serv := &Service{}
	config := serv.Config().(Config)
	config.ConnectionManager = ""

	require.NoError(t, serv.SetConfig(config), "serv.SetConfig(config)")

	deps := map[string]interface{}{
		"swarm": swarmtesting.GenSwarm(t, ctx),
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

	assert.IsType(t, &p2p.Host{}, exposed, "exposed type")
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
		[]string{"connmgr", "myswarm"},
	}, {
		"connmgr",
		func(c *Config) { c.ConnectionManager = "myconnmgr" },
		[]string{"myconnmgr", "swarm"},
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
			"myswarm":    swarmtesting.GenSwarm(t, context.Background()),
			"monitoring": time.Second,
		},
		nil,
	}, {
		"valid swarm",
		func(c *Config) { c.Network = "myswarm" },
		map[string]interface{}{
			"myswarm":    &swarm.Swarm{},
			"monitoring": time.Second,
		},
		nil,
	}, {
		"invalid network",
		func(c *Config) { c.Network = "myswarm" },
		map[string]interface{}{
			"myswarm":    struct{}{},
			"monitoring": time.Second,
		},
		ErrNotNetwork,
	}, {
		"valid connmgr",
		func(c *Config) { c.ConnectionManager = "myconnmgr" },
		map[string]interface{}{
			"myconnmgr":  ifconnmgr.NullConnMgr{},
			"swarm":      swarmtesting.GenSwarm(t, context.Background()),
			"monitoring": time.Second,
		},
		nil,
	}, {
		"invalid connmgr",
		func(c *Config) { c.ConnectionManager = "myconnmgr" },
		map[string]interface{}{
			"myconnmgr":  struct{}{},
			"swarm":      swarmtesting.GenSwarm(t, context.Background()),
			"monitoring": time.Second,
		},
		ErrNotConnManager,
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			serv := Service{}
			config := serv.Config().(Config)
			config.ConnectionManager = ""
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
