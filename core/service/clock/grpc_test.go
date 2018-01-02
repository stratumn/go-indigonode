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

package clock

import (
	"context"
	"testing"
	"time"

	"github.com/pkg/errors"
	pb "github.com/stratumn/alice/grpc/clock"

	peer "gx/ipfs/QmWNY7dV54ZDYmTA1ykVdwNCqC11mpU4zSUp6XDpLTH9eG/go-libp2p-peer"
	pstore "gx/ipfs/QmYijbtjCxFEjSXaudaQAUz3LN5VKLssm8WCUsRoqzXmQR/go-libp2p-peerstore"
)

var testTime time.Time
var testPID peer.ID

func init() {
	v, err := time.Parse("2006-01-02 15:04:05.999999999 -0700 MST", "2017-12-13 17:22:06.403954641 +0100 CET")
	if err != nil {
		panic(err)
	}
	testTime = v.UTC()

	pid, err := peer.IDB58Decode("QmVhJVRSYHNSHgR9dJNbDxu6G7GPPqJAeiJoVRvcexGNf9")
	if err != nil {
		panic(err)
	}
	testPID = pid
}

func testGRPCServer(ctx context.Context, t *testing.T) grpcServer {
	return grpcServer{
		func(ctx context.Context) (*time.Time, error) {
			return &testTime, nil
		},
		func(ctx context.Context, pid peer.ID) (*time.Time, error) {
			return &testTime, nil
		},
		func(ctx context.Context, pi pstore.PeerInfo) error {
			return nil
		},
	}
}

func testGRPCServerUnavailable() grpcServer {
	return grpcServer{
		func(ctx context.Context) (*time.Time, error) {
			return nil, ErrUnavailable
		},
		func(ctx context.Context, pid peer.ID) (*time.Time, error) {
			return nil, ErrUnavailable
		},
		func(ctx context.Context, pi pstore.PeerInfo) error {
			return ErrUnavailable
		},
	}
}

func TestGRPCServer_Local(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	srv := testGRPCServer(ctx, t)

	req := &pb.LocalReq{}
	res, err := srv.Local(ctx, req)
	if err != nil {
		t.Fatalf("srv.Local(ctx, req): error: %s", err)
	}

	if got, want := res.Timestamp, testTime.UnixNano(); got != want {
		t.Errorf("res.Timestamp = %v want %v", got, want)
	}
}

func TestGRPCServer_Local_unavailable(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	srv := testGRPCServerUnavailable()

	req := &pb.LocalReq{}
	_, err := srv.Local(ctx, req)

	if got, want := errors.Cause(err), ErrUnavailable; got != want {
		t.Errorf("srv.Local(ctx, req): error = %v want %v", got, want)
	}
}

func TestGRPCServer_Remote(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	srv := testGRPCServer(ctx, t)

	req := &pb.RemoteReq{PeerId: []byte(testPID)}
	res, err := srv.Remote(ctx, req)
	if err != nil {
		t.Fatalf("srv.Remote(ctx, req): error: %s", err)
	}

	if got, want := res.Timestamp, testTime.UnixNano(); got != want {
		t.Errorf("res.Timestamp = %v want %v", got, want)
	}
}

func TestGRPCServer_Remote_unavailable(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	srv := testGRPCServerUnavailable()

	req := &pb.RemoteReq{PeerId: []byte(testPID)}
	_, err := srv.Remote(ctx, req)

	if got, want := errors.Cause(err), ErrUnavailable; got != want {
		t.Errorf("srv.Remote(ctx, req): error = %v want %v", got, want)
	}
}
