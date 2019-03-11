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
	pb "github.com/stratumn/go-node/core/app/swarm/grpc"
	mockpb "github.com/stratumn/go-node/core/app/swarm/grpc/mockgrpc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	peer "github.com/libp2p/go-libp2p-peer"
	ma "github.com/multiformats/go-multiaddr"
	pstore "github.com/libp2p/go-libp2p-peerstore"
	swarm "github.com/libp2p/go-libp2p-swarm"
	swarmtesting "github.com/libp2p/go-libp2p-swarm/testing"
)

func testGRPCServer(ctx context.Context, t *testing.T) grpcServer {
	swm := (*swarm.Swarm)(swarmtesting.GenSwarm(t, ctx))

	return grpcServer{
		GetSwarm: func() *swarm.Swarm { return swm },
	}
}

func testGRPCServerUnavailable() grpcServer {
	return grpcServer{
		GetSwarm: func() *swarm.Swarm { return nil },
	}
}

func TestGRPCServer_LocalPeer(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	srv := testGRPCServer(ctx, t)
	defer srv.GetSwarm().Close()

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

// testConnect ensures two swarms networks are connected.
func testConnect(ctx context.Context, t *testing.T, n1, n2 *swarm.Swarm) {
	pi1 := n1.Peerstore().PeerInfo(n1.LocalPeer())
	pi2 := n2.Peerstore().PeerInfo(n2.LocalPeer())

	n1.Peerstore().AddAddr(pi2.ID, pi2.Addrs[0], pstore.PermanentAddrTTL)
	n2.Peerstore().AddAddr(pi1.ID, pi1.Addrs[0], pstore.PermanentAddrTTL)

	_, err := n1.DialPeer(ctx, pi2.ID)
	require.NoError(t, err, "n1.Dial(ctx, pi2.ID)")
}

func TestGRPCServer_Peers(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	srv := testGRPCServer(ctx, t)

	n1 := srv.GetSwarm()
	defer n1.Close()

	n2 := swarmtesting.GenSwarm(t, ctx)
	defer n2.Close()

	testConnect(ctx, t, n1, n2)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	req, ss := &pb.PeersReq{}, mockpb.NewMockSwarm_PeersServer(ctrl)

	ss.EXPECT().Send(&pb.Peer{Id: []byte(n2.LocalPeer())})

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

	if !bytes.Equal(c.Addr.Bytes(), msg.RemoteAddress) {
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
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	srv := testGRPCServer(ctx, t)

	n1 := srv.GetSwarm()
	defer n1.Close()

	n2 := swarmtesting.GenSwarm(t, ctx)
	defer n2.Close()

	testConnect(ctx, t, n1, n2)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	req := &pb.ConnectionsReq{}
	ss := mockpb.NewMockSwarm_ConnectionsServer(ctrl)

	ss.EXPECT().Context().Return(ctx).AnyTimes()
	ss.EXPECT().Send(connMatcher{n2.LocalPeer(), n2.ListenAddresses()[0]})

	assert.NoError(t, srv.Connections(req, ss))
}

func TestGRPCServer_Connections_peer(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	srv := testGRPCServer(ctx, t)

	n1 := srv.GetSwarm()
	defer n1.Close()

	n2 := swarmtesting.GenSwarm(t, ctx)
	defer n2.Close()

	n3 := swarmtesting.GenSwarm(t, ctx)
	defer n3.Close()

	testConnect(ctx, t, n1, n2)
	testConnect(ctx, t, n1, n3)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	req := &pb.ConnectionsReq{PeerId: []byte(n2.LocalPeer())}
	ss := mockpb.NewMockSwarm_ConnectionsServer(ctrl)

	ss.EXPECT().Context().Return(ctx).AnyTimes()
	ss.EXPECT().Send(connMatcher{n2.LocalPeer(), n2.ListenAddresses()[0]})

	assert.NoError(t, srv.Connections(req, ss))
}

func TestGRPCServer_Connections_unavailable(t *testing.T) {
	srv := testGRPCServerUnavailable()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	req, ss := &pb.ConnectionsReq{}, mockpb.NewMockSwarm_ConnectionsServer(ctrl)

	assert.Equal(t, ErrUnavailable, errors.Cause(srv.Connections(req, ss)))
}
