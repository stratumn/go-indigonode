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

package clock

import (
	"context"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/pkg/errors"
	"github.com/stratumn/alice/core/manager/testservice"
	"github.com/stratumn/alice/core/p2p"
	"github.com/stratumn/alice/core/service/clock/mockclock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	inet "gx/ipfs/QmU4vCDZTPLDqSDKguWbHCiUe46mZUtmM2g2suBZ9NE8ko/go-libp2p-net"
	testutil "gx/ipfs/QmZTcPxK6VqrwY94JpKZPvEqAZ6tEr1rLrpcqJbbRZbg2V/go-libp2p-netutil"
)

func testService(ctx context.Context, t *testing.T, host Host) *Service {
	serv := &Service{}
	config := serv.Config().(Config)

	require.NoError(t, serv.SetConfig(config), "serv.SetConfig(config)")

	deps := map[string]interface{}{
		"host": host,
	}

	require.NoError(t, serv.Plug(deps), "serv.Plug(deps)")

	return serv
}

func expectHost(ctx context.Context, t *testing.T, host *mockclock.MockHost) {
	host.EXPECT().SetStreamHandler(ProtocolID, gomock.Any())
	host.EXPECT().RemoveStreamHandler(ProtocolID)
}

func TestService_strings(t *testing.T) {
	testservice.CheckStrings(t, &Service{})
}

func TestService_Expose(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	host := mockclock.NewMockHost(ctrl)
	expectHost(ctx, t, host)

	serv := testService(ctx, t, host)
	exposed := testservice.Expose(ctx, t, serv, time.Second)

	assert.IsType(t, &Clock{}, exposed, "exposed type")
}

func TestService_Run(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	host := mockclock.NewMockHost(ctrl)
	expectHost(ctx, t, host)

	serv := testService(ctx, t, host)
	testservice.TestRun(ctx, t, serv, time.Second)
}

func TestService_SetConfig(t *testing.T) {
	errAny := errors.New("any error")

	tests := []struct {
		name string
		set  func(*Config)
		err  error
	}{{
		"invalid write timeout",
		func(c *Config) { c.WriteTimeout = "1" },
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
		"host",
		func(c *Config) { c.Host = "myhost" },
		[]string{"myhost"},
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

	host := mockclock.NewMockHost(ctrl)

	tests := []struct {
		name string
		set  func(*Config)
		deps map[string]interface{}
		err  error
	}{{
		"valid host",
		func(c *Config) { c.Host = "myhost" },
		map[string]interface{}{
			"myhost": host,
		},
		nil,
	}, {
		"invalid host",
		func(c *Config) { c.Host = "myhost" },
		map[string]interface{}{
			"myhost": struct{}{},
		},
		ErrNotHost,
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

func TestClock_RemoteTime(t *testing.T) {
	ctx := context.Background()
	h1 := p2p.NewHost(ctx, testutil.GenSwarmNetwork(t, ctx))
	h2 := p2p.NewHost(ctx, testutil.GenSwarmNetwork(t, ctx))
	defer h1.Close()
	defer h2.Close()

	// connect h1 to h2
	h2pi := h2.Peerstore().PeerInfo(h2.ID())
	require.NoError(t, h1.Connect(ctx, h2pi), "Connect()")

	clockH2 := NewClock(h2, 10*time.Second)
	h2.SetStreamHandler(ProtocolID, func(stream inet.Stream) {
		clockH2.StreamHandler(ctx, stream)
	})

	clockH1 := &Clock{host: h1}
	remoteTime, err := clockH1.RemoteTime(ctx, h2.ID())

	require.NoError(t, err, "RemoteTime()")
	require.NotNil(t, remoteTime)
	require.WithinDuration(t, time.Now().UTC(), *remoteTime, time.Second)
}
