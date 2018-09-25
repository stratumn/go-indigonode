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
	"io/ioutil"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/pkg/errors"
	"github.com/stratumn/go-node/core/app/kaddht/service/mockservice"
	"github.com/stratumn/go-node/core/manager/testservice"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	ifconnmgr "gx/ipfs/QmWGGN1nysi1qgqto31bENwESkmZBY4YGK4sZC3qhnqhSv/go-libp2p-interface-connmgr"
	kaddht "gx/ipfs/QmaXYSwxqJsX3EoGb1ZV2toZ9fXc8hWJPaBW1XAp1h2Tsp/go-libp2p-kad-dht"
	kaddhtopts "gx/ipfs/QmaXYSwxqJsX3EoGb1ZV2toZ9fXc8hWJPaBW1XAp1h2Tsp/go-libp2p-kad-dht/opts"
	swarmtesting "gx/ipfs/QmeDpqUwwdye8ABKVMPXKuWwPVURFdqTqssbTUB39E2Nwd/go-libp2p-swarm/testing"
)

func testService(ctx context.Context, t *testing.T, host Host) *Service {
	dir, err := ioutil.TempDir("", "")
	require.NoError(t, err, `ioutil.TempDir("", "")`)

	serv := &Service{}
	config := serv.Config().(Config)
	config.LevelDBPath = dir

	require.NoError(t, serv.SetConfig(config), "serv.SetConfig(config)")

	deps := map[string]interface{}{
		"host": host,
	}

	require.NoError(t, serv.Plug(deps), "serv.Plug(deps)")

	return serv
}

func expectHost(ctx context.Context, t *testing.T, host *mockservice.MockHost) {
	swm := swarmtesting.GenSwarm(t, ctx)

	host.EXPECT().ID().Return(swm.LocalPeer()).AnyTimes()
	host.EXPECT().Peerstore().Return(swm.Peerstore()).AnyTimes()
	host.EXPECT().ConnManager().Return(ifconnmgr.NullConnMgr{}).AnyTimes()
	host.EXPECT().Network().Return(swm).AnyTimes()
	host.EXPECT().SetStreamHandler(kaddhtopts.ProtocolDHT, gomock.Any())
	host.EXPECT().SetStreamHandler(kaddhtopts.ProtocolDHTOld, gomock.Any())
	host.EXPECT().SetRouter(gomock.Any())
	host.EXPECT().SetRouter(nil)
	host.EXPECT().RemoveStreamHandler(kaddhtopts.ProtocolDHT)
	host.EXPECT().RemoveStreamHandler(kaddhtopts.ProtocolDHTOld)
}

func TestService_strings(t *testing.T) {
	testservice.CheckStrings(t, &Service{})
}

func TestService_Expose(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	host := mockservice.NewMockHost(ctrl)
	expectHost(ctx, t, host)

	serv := testService(ctx, t, host)
	exposed := testservice.Expose(ctx, t, serv, time.Second)

	assert.IsType(t, &kaddht.IpfsDHT{}, exposed, "exposed type")
}

func TestService_Run(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	host := mockservice.NewMockHost(ctrl)
	expectHost(ctx, t, host)

	serv := testService(ctx, t, host)
	testservice.TestRun(ctx, t, serv, time.Second)
}

func TestService_Run_bootstrap(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	host := mockservice.NewMockHost(ctrl)
	expectHost(ctx, t, host)

	serv := testService(ctx, t, host)
	serv.Befriend("bootstrap", true)
	testservice.TestRun(ctx, t, serv, time.Second)
}

func TestService_SetConfig(t *testing.T) {
	errAny := errors.New("any error")

	tests := []struct {
		name string
		set  func(*Config)
		err  error
	}{{
		"invalid bootstrap interval",
		func(c *Config) { c.BootstrapInterval = "1" },
		errAny,
	}, {
		"invalid bootstrap timeout",
		func(c *Config) { c.BootstrapTimeout = "ten" },
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

	host := mockservice.NewMockHost(ctrl)

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

func TestService_Likes(t *testing.T) {
	tests := []struct {
		name  string
		set   func(*Config)
		likes []string
	}{{
		"host",
		func(c *Config) { c.Bootstrap = "mybootstrap" },
		[]string{"mybootstrap"},
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			serv := Service{}
			config := serv.Config().(Config)
			tt.set(&config)

			require.NoError(t, serv.SetConfig(config), "serv.SetConfig(config)")
			assert.Equal(t, toSet(tt.likes), serv.Likes())
		})
	}
}

func toSet(keys []string) map[string]struct{} {
	set := map[string]struct{}{}
	for _, v := range keys {
		set[v] = struct{}{}
	}

	return set
}
