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

package bootstrap

import (
	"context"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stratumn/alice/core/manager/testservice"
	"github.com/stratumn/alice/core/service/bootstrap/mockbootstrap"

	inet "gx/ipfs/QmNa31VPzC561NWwRsJLE7nGYZYuuD2QfpK2b1q9BK54J1/go-libp2p-net"
	pstore "gx/ipfs/QmPgDWmTmuzvP7QE5zwo1TmjbJme9pmZHNujB2453jkCTr/go-libp2p-peerstore"
	peer "gx/ipfs/QmXYjuNuxVzXKJCfWasQk1RqkhVLDM9jtUKhqc2WPQmFSB/go-libp2p-peer"
)

func testService(ctx context.Context, t *testing.T, host Host) *Service {
	serv := &Service{}
	config := serv.Config().(Config)

	config.Addresses = []string{
		"/ip4/127.0.0.1/tcp/54983/ipfs/QmXMTDMJht9Q1KCYgJqVRzNG9kpXoujLXJkPouSowTiwwr",
	}

	config.MinPeerThreshold = 1

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

func expectHost(ctx context.Context, t *testing.T, net *mockbootstrap.MockNetwork, host *mockbootstrap.MockHost) {
	seedID, err := peer.IDB58Decode("QmXMTDMJht9Q1KCYgJqVRzNG9kpXoujLXJkPouSowTiwwr")
	if err != nil {
		t.Fatalf(`peer.IDB58Decode("QmXMTDMJht9Q1KCYgJqVRzNG9kpXoujLXJkPouSowTiwwr"): error: %s`, err)
	}

	ps := pstore.NewPeerstore()

	host.EXPECT().Network().Return(net).AnyTimes()
	host.EXPECT().Peerstore().Return(ps).AnyTimes()

	net.EXPECT().Peers().Return(ps.Peers())
	net.EXPECT().Connectedness(seedID).Return(inet.NotConnected)
	host.EXPECT().Connect(gomock.Any(), gomock.Any()).Return(nil)
	net.EXPECT().Peers().Return([]peer.ID{seedID})
	net.EXPECT().Peers().Return([]peer.ID{seedID})
}

func TestService_Expose(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	host := mockbootstrap.NewMockHost(ctrl)
	net := mockbootstrap.NewMockNetwork(ctrl)
	expectHost(ctx, t, net, host)

	serv := testService(ctx, t, host)
	exposed := testservice.Expose(ctx, t, serv, time.Second)

	ok := exposed != nil
	if got, want := ok, true; got != want {
		t.Errorf("ok = %v want %v", got, want)
	}
}

func TestService_Run(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	host := mockbootstrap.NewMockHost(ctrl)
	net := mockbootstrap.NewMockNetwork(ctrl)
	expectHost(ctx, t, net, host)

	serv := testService(ctx, t, host)
	testservice.TestRun(ctx, t, serv, time.Second)
}
