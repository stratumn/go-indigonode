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

	"github.com/golang/mock/gomock"
	"github.com/pkg/errors"
	pb "github.com/stratumn/go-indigonode/core/app/ping/grpc"
	mockpb "github.com/stratumn/go-indigonode/core/app/ping/grpc/mockgrpc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	peer "gx/ipfs/QmcJukH2sAFjY3HdBKq35WDzWoL3UUu2gt9wdfqZTUyM74/go-libp2p-peer"
	pstore "gx/ipfs/QmdeiKhUy1TVGBaKxt7y1QmBDLBdisSrLJ1x58Eoj4PXUh/go-libp2p-peerstore"
)

const testPID = "QmVhJVRSYHNSHgR9dJNbDxu6G7GPPqJAeiJoVRvcexGNf9"

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

	pid, err := peer.IDB58Decode(testPID)
	require.NoError(t, err, "peer.IDB58Decode(testPID)")

	req := &pb.PingReq{
		PeerId: []byte(pid),
		Times:  2,
	}
	ss := mockpb.NewMockPing_PingServer(ctrl)

	ss.EXPECT().Context().Return(ctx).AnyTimes()
	ss.EXPECT().Send(&pb.Response{Latency: 100000000}).Times(2)

	require.NoError(t, srv.Ping(req, ss))
}

func TestGRPCServer_Ping_unavailable(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	srv := testGRPCServerUnavailable()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	pid, err := peer.IDB58Decode(testPID)
	require.NoError(t, err, "peer.IDB58Decode(testPID)")

	req := &pb.PingReq{
		PeerId: []byte(pid),
		Times:  2,
	}
	ss := mockpb.NewMockPing_PingServer(ctrl)
	ss.EXPECT().Context().Return(ctx).AnyTimes()

	assert.Equal(t, ErrUnavailable, errors.Cause(srv.Ping(req, ss)))
}
