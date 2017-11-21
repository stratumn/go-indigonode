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

package pruner

import (
	"context"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stratumn/alice/core/manager/testservice"
	"github.com/stratumn/alice/core/service/pruner/mockpruner"
)

func testService(ctx context.Context, t *testing.T, mgr *mockpruner.MockManager) *Service {
	serv := &Service{}
	config := serv.Config().(Config)
	config.Interval = "100ms"

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

func expectManager(mgr *mockpruner.MockManager) {
}

func TestService_Run(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mgr := mockpruner.NewMockManager(ctrl)
	expectManager(mgr)

	serv := testService(ctx, t, mgr)
	testservice.TestRun(ctx, t, serv, time.Second)
}

func TestService_job(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mgr := mockpruner.NewMockManager(ctrl)
	expectManager(mgr)

	serv := testService(ctx, t, mgr)

	testservice.TestRunning(ctx, t, serv, time.Second, func() {
		mgr.EXPECT().Prune().MinTimes(1).MaxTimes(5)
		time.Sleep(500 * time.Millisecond)
	})
}
