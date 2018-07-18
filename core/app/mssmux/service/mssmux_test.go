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

	"github.com/pkg/errors"
	"github.com/stratumn/go-indigonode/core/manager/testservice"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	yamux "gx/ipfs/QmNWCEvi7bPRcvqAV8AKLGVNoQdArWi7NJayka2SM4XtRe/go-smux-yamux"
	smux "gx/ipfs/QmY9JXR3FupnYAYJWK9aMr9bCpqWKcToQ1tz8DVGTrHpHw/go-stream-muxer"
)

func testService(ctx context.Context, t *testing.T) *Service {
	serv := &Service{}
	config := serv.Config().(Config)

	require.NoError(t, serv.SetConfig(config), "serv.SetConfig(config)")

	deps := map[string]interface{}{
		"yamux": yamux.DefaultTransport,
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

	assert.Implements(t, (*smux.Transport)(nil), exposed, "exposed type")
}

func TestService_Run(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	serv := testService(ctx, t)
	testservice.TestRun(ctx, t, serv, time.Second)
}

func TestService_Needs(t *testing.T) {
	tests := []struct {
		name  string
		set   func(*Config)
		needs []string
	}{{
		"routes",
		func(c *Config) { c.Routes = map[string]string{"/test": "test"} },
		[]string{"test"},
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
		"valid stream muxer",
		func(c *Config) { c.Routes = map[string]string{"/yamux/v1.0.0": "yamux"} },
		map[string]interface{}{
			"yamux": yamux.DefaultTransport,
		},
		nil,
	}, {
		"invalid stream muxer",
		func(c *Config) { c.Routes = map[string]string{"/yamux/v1.0.0": "yamux"} },
		map[string]interface{}{
			"yamux": struct{}{},
		},
		ErrNotStreamMuxer,
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
