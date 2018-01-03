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

package grpcapi

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/pkg/errors"
	"github.com/stratumn/alice/core/manager"
	"github.com/stratumn/alice/core/manager/testservice"
	"github.com/stratumn/alice/core/netutil"
	"github.com/stratumn/alice/core/service/grpcapi/mockgrpcapi"
	pb "github.com/stratumn/alice/grpc/grpcapi"
	"github.com/stratumn/alice/release"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	testutil "gx/ipfs/QmZTcPxK6VqrwY94JpKZPvEqAZ6tEr1rLrpcqJbbRZbg2V/go-libp2p-netutil"
)

func testService(ctx context.Context, t *testing.T, mgr Manager) *Service {
	serv := &Service{}
	config := serv.Config().(Config)
	config.Address = fmt.Sprintf("/ip4/127.0.0.1/tcp/%d", netutil.RandomPort())

	require.NoError(t, serv.SetConfig(config), "serv.SetConfig(config)")

	deps := map[string]interface{}{
		"manager": mgr,
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

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mgr := mockgrpcapi.NewMockManager(ctrl)
	mgr.EXPECT().List().Return([]string{})

	serv := testService(ctx, t, mgr)
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
		"manager",
		func(c *Config) { c.Manager = "mymgr" },
		[]string{"mymgr"},
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
		"valid manager",
		func(c *Config) { c.Manager = "mymgr" },
		map[string]interface{}{
			"mymgr": mockgrpcapi.NewMockManager(ctrl),
			"swarm": testutil.GenSwarmNetwork(t, context.Background()),
		},
		nil,
	}, {
		"invalid manager",
		func(c *Config) { c.Manager = "mymgr" },
		map[string]interface{}{
			"mymgr": struct{}{},
			"swarm": testutil.GenSwarmNetwork(t, context.Background()),
		},
		ErrNotManager,
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

type dummyService struct {
	id string
}

func (s dummyService) ID() string {
	return s.id
}

func (s dummyService) Name() string {
	return s.id
}

func (s dummyService) Desc() string {
	return s.id
}

func mockRegistrable(id string, reg Registrable) manager.Service {
	return struct {
		manager.Service
		Registrable
	}{
		dummyService{id},
		reg,
	}
}

func TestService_Registrable(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	p2p := mockgrpcapi.NewMockRegistrable(ctrl)
	p2pServ := mockRegistrable("p2p", p2p)

	mgr := mockgrpcapi.NewMockManager(ctrl)
	mgr.EXPECT().List().Return([]string{"p2p"})
	mgr.EXPECT().Find("p2p").Return(p2pServ, nil)
	p2p.EXPECT().AddToGRPCServer(gomock.Any())

	serv := testService(ctx, t, mgr)
	testservice.TestRun(ctx, t, serv, time.Second)
}

func TestService_Inform(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mgr := mockgrpcapi.NewMockManager(ctrl)
	serv := testService(ctx, t, mgr)

	req := &pb.InformReq{}
	res, err := serv.Inform(ctx, req)

	require.NoError(t, err)
	assert.Equal(t, release.Protocol, res.Protocol)
}
