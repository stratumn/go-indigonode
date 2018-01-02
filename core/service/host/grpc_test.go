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

package host

import (
	"bytes"
	"context"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/pkg/errors"
	"github.com/stratumn/alice/core/p2p"
	pb "github.com/stratumn/alice/grpc/host"
	mockpb "github.com/stratumn/alice/grpc/host/mockhost"

	ma "gx/ipfs/QmW8s4zTsUoX1Q6CeYxVKPyqSKbF7H1YDUyTostBtZ8DaG/go-multiaddr"
	peer "gx/ipfs/QmWNY7dV54ZDYmTA1ykVdwNCqC11mpU4zSUp6XDpLTH9eG/go-libp2p-peer"
	testutil "gx/ipfs/QmZTcPxK6VqrwY94JpKZPvEqAZ6tEr1rLrpcqJbbRZbg2V/go-libp2p-netutil"
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
	if err != nil {
		t.Fatalf("srv.ID(ctx, req): error: %s", err)
	}

	id, err := peer.IDFromBytes(res.Id)
	if err != nil {
		t.Fatalf("peer.IDFromBytes(res.ID): error: %s", err)
	}

	got := id.Pretty()
	want := srv.GetHost().ID().Pretty()

	if got != want {
		t.Errorf("id = %v want %v", got, want)
	}
}

func TestGRPCServer_ID_unavailable(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	srv := testGRPCServerUnavailable()

	req := &pb.IdReq{}
	_, err := srv.ID(ctx, req)

	if got, want := errors.Cause(err), ErrUnavailable; got != want {
		t.Errorf("srv.ID(ctx, req): error = %v want %v", got, want)
	}
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

	err := srv.Addresses(req, ss)
	if err != nil {
		t.Fatalf("srv.Addresses(req, ss): error: %s", err)
	}
}

func TestGRPCServer_Addresses_unavailable(t *testing.T) {
	srv := testGRPCServerUnavailable()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	req, ss := &pb.AddressesReq{}, mockpb.NewMockHost_AddressesServer(ctrl)

	err := srv.Addresses(req, ss)

	if got, want := errors.Cause(err), ErrUnavailable; got != want {
		t.Errorf("srv.Addresses(req, ss): error = %v want %v", got, want)
	}
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
	if err != nil {
		t.Fatalf("ma.NewMultiadd(%q): error: %s", addr, err)
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	req := &pb.ConnectReq{Address: maddr.Bytes()}
	ss := mockpb.NewMockHost_ConnectServer(ctrl)

	ss.EXPECT().Context().Return(ctx).AnyTimes()
	ss.EXPECT().Send(connMatcher{h2.ID(), h2.Addrs()[0]})

	err = srv.Connect(req, ss)
	if err != nil {
		t.Fatalf("srv.Connect(req, ss): error: %s", err)
	}
}

func TestGRPCServer_Connect_unavailable(t *testing.T) {
	srv := testGRPCServerUnavailable()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	req, ss := &pb.ConnectReq{}, mockpb.NewMockHost_ConnectServer(ctrl)

	err := srv.Connect(req, ss)

	if got, want := errors.Cause(err), ErrUnavailable; got != want {
		t.Errorf("srv.Connect(req, ss): error = %v want %v", got, want)
	}
}

func TestGRPCServer_Connect_invalidAddress(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	srv := testGRPCServer(ctx, t)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	req, ss := &pb.ConnectReq{}, mockpb.NewMockHost_ConnectServer(ctrl)

	if err := srv.Connect(req, ss); err == nil {
		t.Error("srv.Connect(req, ss): error = <nil> want <error>")
	}
}
