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

package signal

import (
	"context"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/pkg/errors"
	"github.com/stratumn/alice/core/manager/testservice"
	"github.com/stratumn/alice/core/service/signal/mocksignal"
)

func testService(ctx context.Context, t *testing.T, mgr *mocksignal.MockManager) *Service {
	serv := &Service{}
	config := serv.Config().(Config)

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

func TestService_Run(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mgr := mocksignal.NewMockManager(ctrl)
	serv := testService(ctx, t, mgr)
	testservice.TestRun(ctx, t, serv, time.Second)
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

	mgr := mocksignal.NewMockManager(ctrl)

	tt := []struct {
		name string
		set  func(*Config)
		deps map[string]interface{}
		err  error
	}{{
		"valid manager",
		func(c *Config) { c.Manager = "mymgr" },
		map[string]interface{}{
			"mymgr": mgr,
		},
		nil,
	}, {
		"invalid manager",
		func(c *Config) { c.Manager = "mymgr" },
		map[string]interface{}{
			"mymgr": struct{}{},
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
