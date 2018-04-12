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

package store_test

import (
	"context"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stratumn/alice/core/manager/testservice"
	"github.com/stratumn/alice/core/service/indigo/store"
	"github.com/stratumn/alice/core/service/indigo/store/mockstore"
	"github.com/stretchr/testify/require"
)

func testService(ctx context.Context, t *testing.T, host *mockstore.MockHost) *store.Service {
	serv := &store.Service{}
	config := serv.Config().(store.Config)
	config.Version = "1.0.0"
	config.NetworkID = "42"

	require.NoError(t, serv.SetConfig(config), "serv.SetConfig(config)")

	deps := map[string]interface{}{
		"host": host,
	}

	require.NoError(t, serv.Plug(deps), "serv.Plug(deps)")

	return serv
}

func TestService_Strings(t *testing.T) {
	testservice.CheckStrings(t, &store.Service{})
}

func TestService_Run(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	host := mockstore.NewMockHost(ctrl)
	// We don't use host yet, no expectations are needed for now.

	serv := testService(ctx, t, host)
	testservice.TestRun(ctx, t, serv, time.Second)
}
