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

package swarm

import (
	"context"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stratumn/alice/core/manager/testservice"
	"github.com/stratumn/alice/core/service/swarm/mockswarm"

	swarm "gx/ipfs/QmdQFrFnPrKRQtpeHKjZ3cVNwxmGKKS2TvhJTuN9C9yduh/go-libp2p-swarm"
)

func testService(ctx context.Context, t *testing.T, smuxer Transport) *Service {
	serv := &Service{}
	config := serv.Config().(Config)
	config.Addresses = []string{"/ip4/0.0.0.0/tcp/35768"}
	config.Metrics = ""

	if err := serv.SetConfig(config); err != nil {
		t.Fatalf("serv.SetConfig(config): error: %s", err)
	}

	deps := map[string]interface{}{
		"mssmux": smuxer,
	}

	if err := serv.Plug(deps); err != nil {
		t.Fatalf("serv.Plug(deps): error: %s", err)
	}

	return serv
}

func expectTransport(smuxer *mockswarm.MockTransport) {
}

func TestService_Expose(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	smuxer := mockswarm.NewMockTransport(ctrl)
	expectTransport(smuxer)

	serv := testService(ctx, t, smuxer)
	exposed := testservice.Expose(ctx, t, serv, time.Second)

	_, ok := exposed.(*swarm.Swarm)
	if got, want := ok, true; got != want {
		t.Errorf("ok = %v want %v", got, want)
	}
}

func TestService_Run(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	smuxer := mockswarm.NewMockTransport(ctrl)
	expectTransport(smuxer)

	serv := testService(ctx, t, smuxer)
	testservice.TestRun(ctx, t, serv, time.Second)
}
