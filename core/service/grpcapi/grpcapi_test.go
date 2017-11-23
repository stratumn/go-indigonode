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

	testutil "gx/ipfs/QmQGX417WoxKxDJeHqouMEmmH4G1RCENNSzkZYHrXy3Xb3/go-libp2p-netutil"
)

func testService(ctx context.Context, t *testing.T, mgr Manager) *Service {
	serv := &Service{}
	config := serv.Config().(Config)
	config.Address = fmt.Sprintf("/ip4/127.0.0.1/tcp/%d", netutil.RandomPort())

	if err := serv.SetConfig(config); err != nil {
		t.Fatalf("serv.SetConfig(config): error: %s", err)
	}

	deps := map[string]interface{}{
		"manager": mgr,
	}

	if err := serv.Plug(deps); err != nil {
		t.Fatalf("serv.Plug(deps): error: %s", err)
	}

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

	tt := []struct {
		name string
		set  func(*Config)
		err  error
	}{{
		"invalid address",
		func(c *Config) { c.Address = "http://example.com" },
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
		"manager",
		func(c *Config) { c.Manager = "mymgr" },
		[]string{"mymgr"},
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

	for _, test := range tt {
		serv := Service{}
		config := serv.Config().(Config)
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
	if err != nil {
		t.Fatalf("srv.Inform(ctx, req): error: %s", err)
	}

	if got, want := res.Protocol, release.Protocol; got != want {
		t.Errorf("res.Protocol = %v want %v", got, want)
	}
}
