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

package signal

import (
	"context"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/pkg/errors"
	"github.com/stratumn/alice/core/manager/testservice"
	"github.com/stratumn/alice/core/service/signal/mocksignal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testService(ctx context.Context, t *testing.T, mgr *mocksignal.MockManager) *Service {
	serv := &Service{}
	config := serv.Config().(Config)

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

	mgr := mocksignal.NewMockManager(ctrl)
	serv := testService(ctx, t, mgr)
	testservice.TestRun(ctx, t, serv, time.Second)
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

	mgr := mocksignal.NewMockManager(ctrl)

	tests := []struct {
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
