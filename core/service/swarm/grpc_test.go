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

package swarm

import (
	"bytes"
	"context"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/pkg/errors"
	"github.com/stratumn/alice/core/p2p"
	pb "github.com/stratumn/alice/grpc/swarm"
	mockpb "github.com/stratumn/alice/grpc/swarm/mockswarm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	swarm "gx/ipfs/QmUhvp4VoQ9cKDVLqAxciEKdm8ymBx2Syx4C1Tv6SmSTPa/go-libp2p-swarm"
	ma "gx/ipfs/QmW8s4zTsUoX1Q6CeYxVKPyqSKbF7H1YDUyTostBtZ8DaG/go-multiaddr"
	peer "gx/ipfs/QmWNY7dV54ZDYmTA1ykVdwNCqC11mpU4zSUp6XDpLTH9eG/go-libp2p-peer"
	testutil "gx/ipfs/QmZTcPxK6VqrwY94JpKZPvEqAZ6tEr1rLrpcqJbbRZbg2V/go-libp2p-netutil"
)

func testGRPCServer(ctx context.Context, t *testing.T) grpcServer {
	swm := (*swarm.Swarm)(testutil.GenSwarmNetwork(t, ctx))

	return grpcServer{func() *swarm.Swarm { return swm }}
}

func testGRPCServerUnavailable() grpcServer {
	return grpcServer{func() *swarm.Swarm { return nil }}
}

func TestGRPCServer_LocalPeer(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	srv := testGRPCServer(ctx, t)

	req := &pb.LocalPeerReq{}
	res, err := srv.LocalPeer(ctx, req)
	require.NoError(t, err, "srv.LocalPeer(ctx, req)")

	id, err := peer.IDFromBytes(res.Id)
	require.NoError(t, err, "peer.IDFromBytes(res.Id)")

	assert.Equal(t, srv.GetSwarm().LocalPeer(), id)
}

func TestGRPCServer_LocalPeer_unavailable(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	srv := testGRPCServerUnavailable()

	req := &pb.LocalPeerReq{}
	_, err := srv.LocalPeer(ctx, req)

	assert.Equal(t, ErrUnavailable, errors.Cause(err))
}

func testConnect(ctx context.Context, t *testing.T, srv grpcServer) *p2p.Host {
	h := p2p.NewHost(ctx, testutil.GenSwarmNetwork(t, ctx))
	network := (*swarm.Network)(srv.GetSwarm())
	pi := network.Peerstore().PeerInfo(network.LocalPeer())

	require.NoError(t, h.Connect(ctx, pi), "h.Connect(ctx, pi)")

	return h
}

func TestGRPCServer_Peers(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	srv := testGRPCServer(ctx, t)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	h := testConnect(ctx, t, srv)

	req, ss := &pb.PeersReq{}, mockpb.NewMockSwarm_PeersServer(ctrl)

	ss.EXPECT().Send(&pb.Peer{Id: []byte(h.ID())})

	assert.NoError(t, srv.Peers(req, ss))
}

func TestGRPCServer_Peers_unavailable(t *testing.T) {
	srv := testGRPCServerUnavailable()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	req, ss := &pb.PeersReq{}, mockpb.NewMockSwarm_PeersServer(ctrl)

	assert.Equal(t, ErrUnavailable, errors.Cause(srv.Peers(req, ss)))
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

	if !bytes.Equal(c.Addr.Bytes(), msg.LocalAddress) {
		return false
	}

	return true
}

func (c connMatcher) String() string {
	msg := pb.Connection{
		PeerId:       []byte(c.PeerID),
		LocalAddress: c.Addr.Bytes(),
	}

	return msg.String()
}

func TestGRPCServer_Connections(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	srv := testGRPCServer(ctx, t)

	h := testConnect(ctx, t, srv)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	req := &pb.ConnectionsReq{}
	ss := mockpb.NewMockSwarm_ConnectionsServer(ctrl)

	ss.EXPECT().Context().Return(ctx).AnyTimes()
	ss.EXPECT().Send(connMatcher{h.ID(), srv.GetSwarm().ListenAddresses()[0]})

	assert.NoError(t, srv.Connections(req, ss))
}

func TestGRPCServer_Connections_peer(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	srv := testGRPCServer(ctx, t)

	h1 := testConnect(ctx, t, srv)
	testConnect(ctx, t, srv)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	req := &pb.ConnectionsReq{PeerId: []byte(h1.ID())}
	ss := mockpb.NewMockSwarm_ConnectionsServer(ctrl)

	ss.EXPECT().Context().Return(ctx).AnyTimes()
	ss.EXPECT().Send(connMatcher{h1.ID(), srv.GetSwarm().ListenAddresses()[0]})

	assert.NoError(t, srv.Connections(req, ss))
}

func TestGRPCServer_Connections_unavailable(t *testing.T) {
	srv := testGRPCServerUnavailable()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	req, ss := &pb.ConnectionsReq{}, mockpb.NewMockSwarm_ConnectionsServer(ctrl)

	assert.Equal(t, ErrUnavailable, errors.Cause(srv.Connections(req, ss)))
}
