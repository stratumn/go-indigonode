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

	"github.com/golang/mock/gomock"
	"github.com/pkg/errors"
	pb "github.com/stratumn/go-indigonode/core/app/grpcapi/grpc"
	"github.com/stratumn/go-indigonode/core/app/grpcapi/service/mockservice"
	"github.com/stratumn/go-indigonode/core/manager"
	"github.com/stratumn/go-indigonode/core/manager/testservice"
	"github.com/stratumn/go-indigonode/core/netutil"
	"github.com/stratumn/go-indigonode/release"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"

	testutil "gx/ipfs/QmfDapjsRAfzVpjeEm2tSmX19QpCrkLDXRCDDWJcbbUsFn/go-libp2p-netutil"
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

	mgr := mockservice.NewMockManager(ctrl)
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
			"mymgr": mockservice.NewMockManager(ctrl),
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

func TestService_Expose(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mgr := mockservice.NewMockManager(ctrl)
	mgr.EXPECT().List().Return([]string{})

	serv := testService(ctx, t, mgr)
	exposed := testservice.Expose(ctx, t, serv, time.Second)

	assert.IsType(t, &grpc.Server{}, exposed, "exposed type")
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

	p2p := mockservice.NewMockRegistrable(ctrl)
	p2pServ := mockRegistrable("p2p", p2p)

	mgr := mockservice.NewMockManager(ctrl)
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

	mgr := mockservice.NewMockManager(ctrl)
	serv := testService(ctx, t, mgr)

	req := &pb.InformReq{}
	res, err := serv.Inform(ctx, req)

	require.NoError(t, err)
	assert.Equal(t, release.Protocol, res.Protocol)
}
