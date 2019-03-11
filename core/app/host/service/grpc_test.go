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
	"bytes"
	"context"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/pkg/errors"
	pb "github.com/stratumn/go-node/core/app/host/grpc"
	mockpb "github.com/stratumn/go-node/core/app/host/grpc/mockgrpc"
	"github.com/stratumn/go-node/core/p2p"
	"github.com/stratumn/go-node/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/libp2p/go-libp2p-peer"
	ma "github.com/multiformats/go-multiaddr"
	"github.com/libp2p/go-libp2p-peerstore"
	swarmtesting "github.com/libp2p/go-libp2p-swarm/testing"
)

func testGRPCServer(ctx context.Context, t *testing.T) grpcServer {
	h := p2p.NewHost(ctx, swarmtesting.GenSwarm(t, ctx))

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

	h2 := p2p.NewHost(ctx, swarmtesting.GenSwarm(t, ctx))

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

func TestGRPCServer_AddPeerAddress(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	peerID := test.GeneratePeerID(t)
	peerAddr := test.GeneratePeerMultiaddr(t, peerID)

	srv := testGRPCServer(ctx, t)

	pid, err := srv.AddPeerAddress(ctx, &pb.AddPeerAddressReq{
		PeerId:  []byte(peerID),
		Address: peerAddr.Bytes(),
	})

	require.NoError(t, err)
	require.NotNil(t, pid)
	assert.Equal(t, []byte(peerID), pid.Id)

	addrs := srv.GetHost().Peerstore().Addrs(peerID)
	require.Len(t, addrs, 1)
	assert.Equal(t, peerAddr, addrs[0])
}

func TestGRPCServer_PeerAddresses(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	srv := testGRPCServer(ctx, t)

	peerID := test.GeneratePeerID(t)
	peerAddr1 := test.GeneratePeerMultiaddr(t, peerID)
	peerAddr2 := test.GeneratePeerMultiaddr(t, peerID)

	peerStore := srv.GetHost().Peerstore()
	peerStore.AddAddr(peerID, peerAddr1, peerstore.PermanentAddrTTL)
	peerStore.AddAddr(peerID, peerAddr2, peerstore.PermanentAddrTTL)

	ss := mockpb.NewMockHost_PeerAddressesServer(ctrl)
	ss.EXPECT().Send(&pb.Address{Address: peerAddr1.Bytes()})
	ss.EXPECT().Send(&pb.Address{Address: peerAddr2.Bytes()})

	err := srv.PeerAddresses(&pb.PeerAddressesReq{PeerId: []byte(peerID)}, ss)
	require.NoError(t, err)
}

func TestGRPCServer_ClearPeerAddresses(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	srv := testGRPCServer(ctx, t)

	peerID := test.GeneratePeerID(t)
	peerAddr := test.GeneratePeerMultiaddr(t, peerID)

	peerStore := srv.GetHost().Peerstore()
	peerStore.AddAddr(peerID, peerAddr, peerstore.PermanentAddrTTL)

	ss := mockpb.NewMockHost_ClearPeerAddressesServer(ctrl)

	err := srv.ClearPeerAddresses(&pb.PeerAddressesReq{PeerId: []byte(peerID)}, ss)
	require.NoError(t, err)

	assert.Len(t, peerStore.Addrs(peerID), 0)
}
