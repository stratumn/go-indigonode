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

package service

import (
	"context"
	"testing"
	"time"

	"github.com/pkg/errors"
	pb "github.com/stratumn/alice/app/clock/grpc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	peer "gx/ipfs/QmcJukH2sAFjY3HdBKq35WDzWoL3UUu2gt9wdfqZTUyM74/go-libp2p-peer"
	pstore "gx/ipfs/QmdeiKhUy1TVGBaKxt7y1QmBDLBdisSrLJ1x58Eoj4PXUh/go-libp2p-peerstore"
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

	require.NoError(t, err)
	assert.Equal(t, testTime.UnixNano(), res.Timestamp)
}

func TestGRPCServer_Local_unavailable(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	srv := testGRPCServerUnavailable()

	req := &pb.LocalReq{}
	_, err := srv.Local(ctx, req)

	assert.Equal(t, ErrUnavailable, errors.Cause(err))
}

func TestGRPCServer_Remote(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	srv := testGRPCServer(ctx, t)

	req := &pb.RemoteReq{PeerId: []byte(testPID)}
	res, err := srv.Remote(ctx, req)

	require.NoError(t, err)
	assert.Equal(t, testTime.UnixNano(), res.Timestamp)
}

func TestGRPCServer_Remote_unavailable(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	srv := testGRPCServerUnavailable()

	req := &pb.RemoteReq{PeerId: []byte(testPID)}
	_, err := srv.Remote(ctx, req)

	assert.Equal(t, ErrUnavailable, errors.Cause(err))
}
