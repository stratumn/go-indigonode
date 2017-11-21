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

package kaddht

import (
	"context"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stratumn/alice/core/manager/testservice"
	"github.com/stratumn/alice/core/service/kaddht/mockkaddht"

	testutil "gx/ipfs/QmQGX417WoxKxDJeHqouMEmmH4G1RCENNSzkZYHrXy3Xb3/go-libp2p-netutil"
	kaddht "gx/ipfs/QmWRBYr99v8sjrpbyNWMuGkQekn7b9ELoLSCe8Ny7Nxain/go-libp2p-kad-dht"
	ifconnmgr "gx/ipfs/QmYkCrTwivapqdB3JbwvwvxymseahVkcm46ThRMAA24zCr/go-libp2p-interface-connmgr"
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

func expectHost(ctx context.Context, t *testing.T, host *mockkaddht.MockHost) {
	swm := testutil.GenSwarmNetwork(t, ctx)

	host.EXPECT().ID().Return(swm.LocalPeer()).AnyTimes()
	host.EXPECT().Peerstore().Return(swm.Peerstore()).AnyTimes()
	host.EXPECT().ConnManager().Return(ifconnmgr.NullConnMgr{}).AnyTimes()
	host.EXPECT().Network().Return(swm).AnyTimes()
	host.EXPECT().SetStreamHandler(kaddht.ProtocolDHT, gomock.Any())
	host.EXPECT().SetStreamHandler(kaddht.ProtocolDHTOld, gomock.Any())
	host.EXPECT().SetRouter(gomock.Any())
	host.EXPECT().SetRouter(nil)
	host.EXPECT().RemoveStreamHandler(kaddht.ProtocolDHT)
	host.EXPECT().RemoveStreamHandler(kaddht.ProtocolDHTOld)
}

func TestServiceExpose(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	host := mockkaddht.NewMockHost(ctrl)
	expectHost(ctx, t, host)

	serv := testService(ctx, t, host)
	exposed := testservice.Expose(ctx, t, serv, time.Second)

	_, ok := exposed.(*kaddht.IpfsDHT)
	if got, want := ok, true; got != want {
		t.Errorf("ok = %v want %v", got, want)
	}
}

func TestServiceRun(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	host := mockkaddht.NewMockHost(ctrl)
	expectHost(ctx, t, host)

	serv := testService(ctx, t, host)
	testservice.TestRun(ctx, t, serv, time.Second)
}

func TestServiceRun_bootstrap(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	host := mockkaddht.NewMockHost(ctrl)
	expectHost(ctx, t, host)

	serv := testService(ctx, t, host)
	serv.Befriend("bootstrap", true)
	testservice.TestRun(ctx, t, serv, time.Second)
}
