// Copyright © 2017-2018 Stratumn SAS
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as
// published by the Free Software Foundation, either version 3 of the
// License, or (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <https://www.gnu.org/licenses/>.

package service

import (
	"context"
	"testing"
	"time"

	"github.com/pkg/errors"
	pb "github.com/stratumn/go-node/app/clock/grpc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	peer "github.com/libp2p/go-libp2p-peer"
	pstore "github.com/libp2p/go-libp2p-peerstore"
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
