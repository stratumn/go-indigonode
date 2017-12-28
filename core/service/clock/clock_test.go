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

	inet "gx/ipfs/QmU4vCDZTPLDqSDKguWbHCiUe46mZUtmM2g2suBZ9NE8ko/go-libp2p-net"
	testutil "gx/ipfs/QmZTcPxK6VqrwY94JpKZPvEqAZ6tEr1rLrpcqJbbRZbg2V/go-libp2p-netutil"
)

func testService(ctx context.Context, t *testing.T, host Host) *Service {
	serv := &Service{}
	config := serv.Config().(Config)

	if err := serv.SetConfig(config); err != nil {
		t.Fatalf("serv.SetConfig(config): error: %s", err)
	}

	deps := map[string]interface{}{
		"host": host,
	}

	if err := serv.Plug(deps); err != nil {
		t.Fatalf("serv.Plug(deps): error: %s", err)
	}

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

	_, ok := exposed.(*Clock)
	if got, want := ok, true; got != want {
		t.Errorf("ok = %v want %v", got, want)
	}
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
		serv := Service{}
		config := serv.Config().(Config)
		tt.set(&config)

		err := errors.Cause(serv.SetConfig(config))
		switch {
		case err != nil && tt.err == errAny:
		case err != tt.err:
			t.Errorf("%s: err = %v want %v", tt.name, err, tt.err)
		}
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

	for _, tt := range tests {
		serv := Service{}
		config := serv.Config().(Config)
		tt.set(&config)

		if err := serv.SetConfig(config); err != nil {
			t.Errorf("%s: serv.SetConfig(config): error: %s", tt.name, err)
			continue
		}

		needs := serv.Needs()
		for _, n := range tt.needs {
			if _, ok := needs[n]; !ok {
				t.Errorf("%s: needs[%q] = nil want struct{}{}", tt.name, n)
			}
		}
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
		serv := Service{}
		config := serv.Config().(Config)
		tt.set(&config)

		if err := serv.SetConfig(config); err != nil {
			t.Errorf("%s: serv.SetConfig(config): error: %s", tt.name, err)
			continue
		}

		err := errors.Cause(serv.Plug(tt.deps))
		switch {
		case err != nil && tt.err == errAny:
		case err != tt.err:
			t.Errorf("%s: err = %v want %v", tt.name, err, tt.err)
		}
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
	if err := h1.Connect(ctx, h2pi); err != nil {
		t.Fatal(err)
	}

	clockH2 := NewClock(h2, 10*time.Second)
	h2.SetStreamHandler(ProtocolID, func(stream inet.Stream) {
		clockH2.StreamHandler(ctx, stream)
	})

	clockH1 := &Clock{host: h1}
	remoteTime, err := clockH1.RemoteTime(ctx, h2.ID())
	if err != nil {
		t.Fatalf("clock.RemoteTime: error: %s", err)
	}
	if remoteTime == nil {
		t.Fatalf("clock.RemoteTime: expected time not to be nil")
	}
	if time.Now().UTC().Sub(*remoteTime) > time.Second {
		t.Errorf("clock.RemoteTime: unexpected time: %v", remoteTime)
	}
}
