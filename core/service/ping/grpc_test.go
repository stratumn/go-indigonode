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

package ping

import (
	"context"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/pkg/errors"
	pb "github.com/stratumn/alice/grpc/ping"
	mockpb "github.com/stratumn/alice/grpc/ping/mockping"

	pstore "gx/ipfs/QmPgDWmTmuzvP7QE5zwo1TmjbJme9pmZHNujB2453jkCTr/go-libp2p-peerstore"
	peer "gx/ipfs/QmXYjuNuxVzXKJCfWasQk1RqkhVLDM9jtUKhqc2WPQmFSB/go-libp2p-peer"
)

func testGRPCServer(ctx context.Context, t *testing.T, times int) grpcServer {
	return grpcServer{
		func(ctx context.Context, pid peer.ID) (<-chan time.Duration, error) {
			ch := make(chan time.Duration, times)
			go func() {
				<-ctx.Done()
				close(ch)
			}()
			for i := 0; i < times; i++ {
				ch <- 100 * time.Millisecond
			}
			return ch, nil
		},
		func(ctx context.Context, pi pstore.PeerInfo) error {
			return nil
		},
	}
}

func testGRPCServerUnavailable() grpcServer {
	return grpcServer{
		func(ctx context.Context, pid peer.ID) (<-chan time.Duration, error) {
			return nil, ErrUnavailable
		},
		func(ctx context.Context, pi pstore.PeerInfo) error {
			return ErrUnavailable
		},
	}
}

func TestGRPCServer_Ping(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	srv := testGRPCServer(ctx, t, 2)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	pid, err := peer.IDB58Decode("QmVhJVRSYHNSHgR9dJNbDxu6G7GPPqJAeiJoVRvcexGNf9")
	if err != nil {
		t.Fatalf(`peer.IDB58Decode("QmVhJVRSYHNSHgR9dJNbDxu6G7GPPqJAeiJoVRvcexGNf9"): error: %s`, err)
	}

	req := &pb.PingReq{
		PeerId: []byte(pid),
		Times:  2,
	}
	ss := mockpb.NewMockPing_PingServer(ctrl)

	ss.EXPECT().Context().Return(ctx).AnyTimes()
	ss.EXPECT().Send(&pb.Response{Latency: 100000000}).Times(2)

	err = srv.Ping(req, ss)
	if err != nil {
		t.Fatalf("srv.Ping(req, ss): error: %s", err)
	}
}

func TestGRPCServer_Ping_unavailable(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	srv := testGRPCServerUnavailable()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	pid, err := peer.IDB58Decode("QmVhJVRSYHNSHgR9dJNbDxu6G7GPPqJAeiJoVRvcexGNf9")
	if err != nil {
		t.Fatalf(`peer.IDB58Decode("QmVhJVRSYHNSHgR9dJNbDxu6G7GPPqJAeiJoVRvcexGNf9"): error: %s`, err)
	}

	req := &pb.PingReq{
		PeerId: []byte(pid),
		Times:  2,
	}
	ss := mockpb.NewMockPing_PingServer(ctrl)
	ss.EXPECT().Context().Return(ctx).AnyTimes()

	err = srv.Ping(req, ss)

	if got, want := errors.Cause(err), ErrUnavailable; got != want {
		t.Errorf("srv.Ping(req, ss): error = %v want %v", got, want)
	}
}
