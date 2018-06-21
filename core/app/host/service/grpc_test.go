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
	"bytes"
	"context"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/pkg/errors"
	pb "github.com/stratumn/alice/core/app/host/grpc"
	mockpb "github.com/stratumn/alice/core/app/host/grpc/mockgrpc"
	"github.com/stratumn/alice/core/p2p"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	ma "gx/ipfs/QmWWQ2Txc2c6tqjsBpzg5Ar652cHPGNsQQp2SejkNmkUMb/go-multiaddr"
	testutil "gx/ipfs/Qmb6BsZf6Y3kxffXMNTubGPF1w1bkHtpvhfYbmnwP3NQyw/go-libp2p-netutil"
	peer "gx/ipfs/QmcJukH2sAFjY3HdBKq35WDzWoL3UUu2gt9wdfqZTUyM74/go-libp2p-peer"
)

func testGRPCServer(ctx context.Context, t *testing.T) grpcServer {
	h := p2p.NewHost(ctx, testutil.GenSwarmNetwork(t, ctx))

	return grpcServer{func() *p2p.Host { return h }}
}

func testGRPCServerUnavailable() grpcServer {
	return grpcServer{func() *p2p.Host { return nil }}
}

func TestGRPCServer_ID(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	srv := testGRPCServer(ctx, t)

	req := &pb.IdReq{}
	res, err := srv.ID(ctx, req)
	require.NoError(t, err)

	id, err := peer.IDFromBytes(res.Id)
	require.NoError(t, err, "peer.IDFromBytes(res.ID)")

	got := id.Pretty()
	want := srv.GetHost().ID().Pretty()

	assert.Equal(t, want, got)
}

func TestGRPCServer_ID_unavailable(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	srv := testGRPCServerUnavailable()

	req := &pb.IdReq{}
	_, err := srv.ID(ctx, req)

	assert.Equal(t, ErrUnavailable, errors.Cause(err))
}

func TestGRPCServer_Addresses(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	srv := testGRPCServer(ctx, t)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	req, ss := &pb.AddressesReq{}, mockpb.NewMockHost_AddressesServer(ctrl)

	for _, addr := range srv.GetHost().Addrs() {
		ss.EXPECT().Send(&pb.Address{Address: addr.Bytes()})
	}

	assert.NoError(t, srv.Addresses(req, ss))
}

func TestGRPCServer_Addresses_unavailable(t *testing.T) {
	srv := testGRPCServerUnavailable()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	req, ss := &pb.AddressesReq{}, mockpb.NewMockHost_AddressesServer(ctrl)

	assert.Equal(t, ErrUnavailable, errors.Cause(srv.Addresses(req, ss)))
}

type connMatcher struct {
	PeerID peer.ID
	Addr   ma.Multiaddr
}

func (c connMatcher) Matches(x interface{}) bool {
	msg, ok := x.(*pb.Connection)
	if !ok {
		return false
	}

	if !bytes.Equal([]byte(c.PeerID), msg.PeerId) {
		return false
	}

	if !bytes.Equal(c.Addr.Bytes(), msg.RemoteAddress) {
		return false
	}

	return true
}

func (c connMatcher) String() string {
	msg := pb.Connection{
		PeerId:        []byte(c.PeerID),
		RemoteAddress: c.Addr.Bytes(),
	}

	return msg.String()
}

func TestGRPCServer_Connect(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	srv := testGRPCServer(ctx, t)

	h2 := p2p.NewHost(ctx, testutil.GenSwarmNetwork(t, ctx))

	addr := h2.Addrs()[0].String() + "/ipfs/" + h2.ID().Pretty()
	maddr, err := ma.NewMultiaddr(addr)
	require.NoError(t, err, "ma.NewMultiaddr(addr)")

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	req := &pb.ConnectReq{Address: maddr.Bytes()}
	ss := mockpb.NewMockHost_ConnectServer(ctrl)

	ss.EXPECT().Context().Return(ctx).AnyTimes()
	ss.EXPECT().Send(connMatcher{h2.ID(), h2.Addrs()[0]})

	assert.NoError(t, srv.Connect(req, ss))
}

func TestGRPCServer_Connect_unavailable(t *testing.T) {
	srv := testGRPCServerUnavailable()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	req, ss := &pb.ConnectReq{}, mockpb.NewMockHost_ConnectServer(ctrl)

	assert.Equal(t, ErrUnavailable, errors.Cause(srv.Connect(req, ss)))
}

func TestGRPCServer_Connect_invalidAddress(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	srv := testGRPCServer(ctx, t)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	req, ss := &pb.ConnectReq{}, mockpb.NewMockHost_ConnectServer(ctrl)

	assert.Error(t, srv.Connect(req, ss))
}
