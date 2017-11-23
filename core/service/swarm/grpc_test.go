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

	testutil "gx/ipfs/QmQGX417WoxKxDJeHqouMEmmH4G1RCENNSzkZYHrXy3Xb3/go-libp2p-netutil"
	ma "gx/ipfs/QmXY77cVe7rVRQXZZQRioukUM7aRW3BTcAgJe12MCtb3Ji/go-multiaddr"
	peer "gx/ipfs/QmXYjuNuxVzXKJCfWasQk1RqkhVLDM9jtUKhqc2WPQmFSB/go-libp2p-peer"
	swarm "gx/ipfs/QmdQFrFnPrKRQtpeHKjZ3cVNwxmGKKS2TvhJTuN9C9yduh/go-libp2p-swarm"
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
	if err != nil {
		t.Fatalf("srv.LocalPeer(ctx, req): error: %s", err)
	}

	id, err := peer.IDFromBytes(res.Id)
	if err != nil {
		t.Fatalf("peer.IDFromBytes(res.ID): error: %s", err)
	}

	got := id.Pretty()
	want := srv.GetSwarm().LocalPeer().Pretty()

	if got != want {
		t.Errorf("id = %v want %v", got, want)
	}
}

func TestGRPCServer_LocalPeer_unavailable(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	srv := testGRPCServerUnavailable()

	req := &pb.LocalPeerReq{}
	_, err := srv.LocalPeer(ctx, req)

	if got, want := errors.Cause(err), ErrUnavailable; got != want {
		t.Errorf("srv.LocalPeer(ctx, req): error = %v want %v", got, want)
	}
}

func testConnect(ctx context.Context, t *testing.T, srv grpcServer) *p2p.Host {
	h := p2p.NewHost(
		ctx,
		testutil.GenSwarmNetwork(t, ctx),
		nil,
		nil,
		time.Second,
		nil,
		nil,
	)

	network := (*swarm.Network)(srv.GetSwarm())

	pi := network.Peerstore().PeerInfo(network.LocalPeer())

	err := h.Connect(ctx, pi)
	if err != nil {
		t.Fatalf("h.DialPeer(ctx, pid): error: %s", err)
	}

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

	err := srv.Peers(req, ss)
	if err != nil {
		t.Fatalf("srv.Peers(req, ss): error: %s", err)
	}
}

func TestGRPCServer_Peers_unavailable(t *testing.T) {
	srv := testGRPCServerUnavailable()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	req, ss := &pb.PeersReq{}, mockpb.NewMockSwarm_PeersServer(ctrl)

	err := srv.Peers(req, ss)

	if got, want := errors.Cause(err), ErrUnavailable; got != want {
		t.Errorf("srv.Peers(req, ss): error = %v want %v", got, want)
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
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	srv := testGRPCServer(ctx, t)

	h := testConnect(ctx, t, srv)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	req := &pb.ConnectionsReq{}
	ss := mockpb.NewMockSwarm_ConnectionsServer(ctrl)

	ss.EXPECT().Context().Return(ctx).AnyTimes()
	ss.EXPECT().Send(connMatcher{h.ID(), srv.GetSwarm().ListenAddresses()[0]})

	err := srv.Connections(req, ss)
	if err != nil {
		t.Fatalf("srv.Connections(req, ss): error: %s", err)
	}
}

func TestGRPCServer_Connections_peer(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
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

	err := srv.Connections(req, ss)
	if err != nil {
		t.Fatalf("srv.Connections(req, ss): error: %s", err)
	}
}

func TestGRPCServer_Connections_unavailable(t *testing.T) {
	srv := testGRPCServerUnavailable()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	req, ss := &pb.ConnectionsReq{}, mockpb.NewMockSwarm_ConnectionsServer(ctrl)

	err := srv.Connections(req, ss)

	if got, want := errors.Cause(err), ErrUnavailable; got != want {
		t.Errorf("srv.Connections(req, ss): error = %v want %v", got, want)
	}
}
