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
	"google.golang.org/grpc"
)

func testService(ctx context.Context, t *testing.T) *Service {
	serv := &Service{}
	config := serv.Config().(Config)
	config.Address = fmt.Sprintf("/ip4/127.0.0.1/tcp/%d", netutil.RandomPort())

	require.NoError(t, serv.SetConfig(config), "serv.SetConfig(config)")

	deps := map[string]interface{}{
		"grpcapi": grpc.NewServer(),
	}

	require.NoError(t, serv.Plug(deps), "serv.Plug(deps)")

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
		"invalid address",
		func(c *Config) { c.Address = "http://example.com" },
		errAny,
	}, {
		"valid address",
		func(c *Config) { c.Address = "/ip4/127.0.0.1/tcp/8906" },
		nil,
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
		"grpcapi",
		func(c *Config) { c.Grpcapi = "mygrpcapi" },
		[]string{"mygrpcapi"},
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

	tests := []struct {
		name string
		set  func(*Config)
		deps map[string]interface{}
		err  error
	}{{
		"valid gRPC server",
		func(c *Config) { c.Grpcapi = "mygrpcapi" },
		map[string]interface{}{
			"mygrpcapi": grpc.NewServer(),
		},
		nil,
	}, {
		"invalid gRPC server",
		func(c *Config) { c.Grpcapi = "mygrpcapi" },
		map[string]interface{}{
			"mygrpcapi": struct{}{},
		},
		ErrNotGrpcServer,
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			serv := Service{}
			config := serv.Config().(Config)
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
