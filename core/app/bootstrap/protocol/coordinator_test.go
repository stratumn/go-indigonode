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

package protocol_test

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/pkg/errors"
	"github.com/stratumn/go-indigonode/core/app/bootstrap/pb"
	"github.com/stratumn/go-indigonode/core/app/bootstrap/protocol"
	"github.com/stratumn/go-indigonode/core/app/bootstrap/protocol/proposal"
	"github.com/stratumn/go-indigonode/core/app/bootstrap/protocol/proposal/mocks"
	"github.com/stratumn/go-indigonode/core/app/bootstrap/protocol/proposaltest"
	"github.com/stratumn/go-indigonode/core/protector"
	protectorpb "github.com/stratumn/go-indigonode/core/protector/pb"
	"github.com/stratumn/go-indigonode/core/protector/protectortest"
	"github.com/stratumn/go-indigonode/core/streamutil"
	"github.com/stratumn/go-indigonode/core/streamutil/mockstream"
	"github.com/stratumn/go-indigonode/core/streamutil/streamtest"
	"github.com/stratumn/go-indigonode/test"
	"github.com/stratumn/go-indigonode/test/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gx/ipfs/QmQsErDt8Qgw1XrsXf2BpEzDgGWtB1YLsTAARBup5b6B9W/go-libp2p-peer"
	"gx/ipfs/QmYmsdtJ3HsodkePE3eU3TsCaP2YvPZJ4LoXnNkDE5Tpt7/go-multiaddr"
	inet "gx/ipfs/QmZNJyx9GGCX4GeuHnLB8fxaxMLs4MjTjHokxfQcCd6Nve/go-libp2p-net"
	"gx/ipfs/Qmda4cPRvSRyox3SqgJN6DfSZGU5TtHufPTp9uXjFj71X6/go-libp2p-peerstore"
)

func expectCoordinatorHost(host *mocks.MockHost) {
	host.EXPECT().SetStreamHandler(protocol.PrivateCoordinatorHandshakePID, gomock.Any())
	host.EXPECT().SetStreamHandler(protocol.PrivateCoordinatorProposePID, gomock.Any())
	host.EXPECT().SetStreamHandler(protocol.PrivateCoordinatorVotePID, gomock.Any())
}

func TestCoordinator_Close(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	host := mocks.NewMockHost(ctrl)
	expectCoordinatorHost(host)

	handler := protocol.NewCoordinatorHandler(host, nil, nil, nil)
	require.NotNil(t, handler)

	host.EXPECT().RemoveStreamHandler(protocol.PrivateCoordinatorHandshakePID)
	host.EXPECT().RemoveStreamHandler(protocol.PrivateCoordinatorProposePID)
	host.EXPECT().RemoveStreamHandler(protocol.PrivateCoordinatorVotePID)

	handler.Close(context.Background())
}

func TestCoordinator_ValidateSender(t *testing.T) {
	ctx := context.Background()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	host := mocks.NewMockHost(ctrl)
	expectCoordinatorHost(host)

	networkCfg := protectortest.NewTestNetworkConfig(t, protectorpb.NetworkState_BOOTSTRAP)
	handler := protocol.NewCoordinatorHandler(host, nil, networkCfg, nil).(*protocol.CoordinatorHandler)

	peerID := test.GeneratePeerID(t)
	err := handler.ValidateSender(ctx, peerID)
	require.NoError(t, err)

	networkCfg.SetNetworkState(ctx, protectorpb.NetworkState_PROTECTED)
	err = handler.ValidateSender(ctx, peerID)
	require.EqualError(t, err, protector.ErrConnectionRefused.Error())

	networkCfg.AddPeer(ctx, peerID, test.GeneratePeerMultiaddrs(t, peerID))
	err = handler.ValidateSender(ctx, peerID)
	require.NoError(t, err)
}

type CoordinatorHandleTestCase struct {
	name       string
	remotePeer peer.ID
	configure  func(*testing.T, *gomock.Controller, *mocks.MockHost, *mockstream.MockCodec, protector.NetworkConfig, proposal.Store)
	validate   func(*testing.T, protector.NetworkConfig, proposal.Store)
	err        error
}

func (ht *CoordinatorHandleTestCase) Run(
	t *testing.T,
	h func(*protocol.CoordinatorHandler) streamutil.AutoCloseHandler,
) {
	t.Run(ht.name, func(t *testing.T) {
		ctx := context.Background()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		host := mocks.NewMockHost(ctrl)
		expectCoordinatorHost(host)

		conn := mocks.NewMockConn(ctrl)
		conn.EXPECT().RemotePeer().Return(ht.remotePeer)

		stream := mocks.NewMockStream(ctrl)
		stream.EXPECT().Conn().Return(conn)

		networkCfg := protectortest.NewTestNetworkConfig(t, protectorpb.NetworkState_BOOTSTRAP)
		propStore := proposal.NewInMemoryStore()

		codec := mockstream.NewMockCodec(ctrl)
		ht.configure(t, ctrl, host, codec, networkCfg, propStore)

		handler := protocol.NewCoordinatorHandler(
			host,
			streamutil.NewStreamProvider(),
			networkCfg,
			propStore,
		).(*protocol.CoordinatorHandler)

		err := h(handler)(ctx, test.NewSpan(), stream, codec)
		if ht.err != nil {
			assert.EqualError(t, err, ht.err.Error())
		} else {
			require.NoError(t, err)
			ht.validate(t, networkCfg, propStore)
		}
	})
}

func TestCoordinator_HandleHandshake(t *testing.T) {
	peer1 := test.GeneratePeerID(t)
	peer1Addrs := test.GeneratePeerMultiaddrs(t, peer1)

	testCases := []CoordinatorHandleTestCase{
		{
			"reject-stream-error",
			peer1,
			func(_ *testing.T, _ *gomock.Controller, _ *mocks.MockHost, codec *mockstream.MockCodec, _ protector.NetworkConfig, _ proposal.Store) {
				codec.EXPECT().Decode(gomock.Any()).Return(errors.New("critical system failure"))
			},
			func(*testing.T, protector.NetworkConfig, proposal.Store) {},
			protector.ErrConnectionRefused,
		}, {
			"during-bootstrap-send-participants-to-white-listed-peer",
			peer1,
			func(t *testing.T, _ *gomock.Controller, _ *mocks.MockHost, codec *mockstream.MockCodec, cfg protector.NetworkConfig, _ proposal.Store) {
				cfg.AddPeer(context.Background(), peer1, peer1Addrs)

				codec.EXPECT().Decode(gomock.Any())
				codec.EXPECT().Encode(gomock.Any()).Do(func(n interface{}) error {
					sentCfg, ok := n.(*protectorpb.NetworkConfig)
					require.True(t, ok, "n.(*protectorpb.NetworkConfig)")

					_, ok = sentCfg.Participants[peer1.Pretty()]
					require.True(t, ok, "sentCfg.Participants[peer1.Pretty()]")

					return nil
				})
			},
			func(*testing.T, protector.NetworkConfig, proposal.Store) {},
			nil,
		}, {
			"during-bootstrap-do-not-send-participants-to-non-white-listed-peer",
			peer1,
			func(t *testing.T, _ *gomock.Controller, _ *mocks.MockHost, codec *mockstream.MockCodec, _ protector.NetworkConfig, _ proposal.Store) {
				codec.EXPECT().Decode(gomock.Any())
				codec.EXPECT().Encode(gomock.Any()).Do(func(n interface{}) error {
					sentCfg, ok := n.(*protectorpb.NetworkConfig)
					require.True(t, ok, "n.(*protectorpb.NetworkConfig)")
					require.Len(t, sentCfg.Participants, 0)

					return nil
				})
			},
			func(*testing.T, protector.NetworkConfig, proposal.Store) {},
			nil,
		}, {
			"after-bootstrap-send-participants-to-white-listed-peer",
			peer1,
			func(t *testing.T, _ *gomock.Controller, _ *mocks.MockHost, codec *mockstream.MockCodec, cfg protector.NetworkConfig, _ proposal.Store) {
				cfg.SetNetworkState(context.Background(), protectorpb.NetworkState_PROTECTED)
				cfg.AddPeer(context.Background(), peer1, peer1Addrs)

				codec.EXPECT().Decode(gomock.Any())
				codec.EXPECT().Encode(gomock.Any()).Do(func(n interface{}) error {
					sentCfg, ok := n.(*protectorpb.NetworkConfig)
					require.True(t, ok, "n.(*protectorpb.NetworkConfig)")

					_, ok = sentCfg.Participants[peer1.Pretty()]
					require.True(t, ok, "sentCfg.Participants[peer1.Pretty()]")

					return nil
				})
			},
			func(*testing.T, protector.NetworkConfig, proposal.Store) {},
			nil,
		}, {
			"after-bootstrap-reject-non-white-listed-peer",
			peer1,
			func(_ *testing.T, _ *gomock.Controller, _ *mocks.MockHost, _ *mockstream.MockCodec, cfg protector.NetworkConfig, _ proposal.Store) {
				cfg.SetNetworkState(context.Background(), protectorpb.NetworkState_PROTECTED)
			},
			func(*testing.T, protector.NetworkConfig, proposal.Store) {},
			protector.ErrConnectionRefused,
		}}

	for _, tt := range testCases {
		tt.Run(t, func(handler *protocol.CoordinatorHandler) streamutil.AutoCloseHandler {
			return handler.HandleHandshake
		})
	}
}

func TestCoordinator_HandlePropose(t *testing.T) {
	hostID := test.GeneratePeerID(t)
	peer1 := test.GeneratePeerID(t)
	peer2 := test.GeneratePeerID(t)

	testCases := []CoordinatorHandleTestCase{
		{
			"decode-error",
			peer1,
			func(t *testing.T, _ *gomock.Controller, h *mocks.MockHost, codec *mockstream.MockCodec, cfg protector.NetworkConfig, ps proposal.Store) {
				codec.EXPECT().Decode(gomock.Any()).Return(errors.New("heap corruption"))
			},
			func(*testing.T, protector.NetworkConfig, proposal.Store) {},
			protector.ErrConnectionRefused,
		},
		{
			"unauthorized",
			peer1,
			func(t *testing.T, _ *gomock.Controller, h *mocks.MockHost, codec *mockstream.MockCodec, cfg protector.NetworkConfig, ps proposal.Store) {
				cfg.SetNetworkState(context.Background(), protectorpb.NetworkState_PROTECTED)
			},
			func(*testing.T, protector.NetworkConfig, proposal.Store) {},
			protector.ErrConnectionRefused,
		},
		{
			"during-bootstrap-invalid-peer-id",
			peer1,
			func(t *testing.T, _ *gomock.Controller, h *mocks.MockHost, codec *mockstream.MockCodec, cfg protector.NetworkConfig, ps proposal.Store) {
				streamtest.ExpectDecodeNodeID(t, codec, &pb.NodeIdentity{PeerId: []byte("b4tm4n")})
				streamtest.ExpectEncodeAck(t, codec, proposal.ErrInvalidPeerID)
			},
			func(*testing.T, protector.NetworkConfig, proposal.Store) {},
			nil,
		},
		{
			"during-bootstrap-missing-peer-addr",
			peer1,
			func(t *testing.T, _ *gomock.Controller, h *mocks.MockHost, codec *mockstream.MockCodec, cfg protector.NetworkConfig, ps proposal.Store) {
				h.EXPECT().Peerstore().Return(peerstore.NewPeerstore())

				streamtest.ExpectDecodeNodeID(t, codec, &pb.NodeIdentity{PeerId: []byte(peer1)})
				streamtest.ExpectEncodeAck(t, codec, proposal.ErrMissingPeerAddr)
			},
			func(*testing.T, protector.NetworkConfig, proposal.Store) {},
			nil,
		},
		{
			"during-bootstrap-node-mismatch",
			peer1,
			func(t *testing.T, _ *gomock.Controller, h *mocks.MockHost, codec *mockstream.MockCodec, cfg protector.NetworkConfig, ps proposal.Store) {
				nodeID := &pb.NodeIdentity{
					PeerId:   []byte(peer2),
					PeerAddr: test.GeneratePeerMultiaddr(t, peer2).Bytes(),
				}
				streamtest.ExpectDecodeNodeID(t, codec, nodeID)
				streamtest.ExpectEncodeAck(t, codec, proposal.ErrInvalidPeerAddr)
			},
			func(*testing.T, protector.NetworkConfig, proposal.Store) {},
			nil,
		},
		{
			"during-bootstrap-addr-in-peerstore",
			peer1,
			func(t *testing.T, _ *gomock.Controller, h *mocks.MockHost, codec *mockstream.MockCodec, cfg protector.NetworkConfig, ps proposal.Store) {
				pstore := peerstore.NewPeerstore()
				pstore.AddAddr(peer1, test.GeneratePeerMultiaddr(t, peer1), peerstore.PermanentAddrTTL)
				h.EXPECT().Peerstore().Return(pstore)

				streamtest.ExpectDecodeNodeID(t, codec, &pb.NodeIdentity{PeerId: []byte(peer1)})
				streamtest.ExpectEncodeAck(t, codec, nil)
			},
			func(t *testing.T, cfg protector.NetworkConfig, s proposal.Store) {
				r, err := s.Get(context.Background(), peer1)
				require.NoError(t, err, "s.Get()")
				require.NotNil(t, r)
				assert.Equal(t, proposal.AddNode, r.Type)
			},
			nil,
		},
		{
			"during-bootstrap-addr-provided",
			peer1,
			func(t *testing.T, _ *gomock.Controller, h *mocks.MockHost, codec *mockstream.MockCodec, cfg protector.NetworkConfig, ps proposal.Store) {
				nodeID := &pb.NodeIdentity{
					PeerId:   []byte(peer1),
					PeerAddr: test.GeneratePeerMultiaddr(t, peer1).Bytes(),
				}
				streamtest.ExpectDecodeNodeID(t, codec, nodeID)
				streamtest.ExpectEncodeAck(t, codec, nil)
			},
			func(t *testing.T, cfg protector.NetworkConfig, s proposal.Store) {
				r, err := s.Get(context.Background(), peer1)
				require.NoError(t, err, "s.Get()")
				require.NotNil(t, r)
				assert.Equal(t, proposal.AddNode, r.Type)
			},
			nil,
		},
		{
			"after-bootstrap-add-node",
			peer1,
			func(t *testing.T, _ *gomock.Controller, h *mocks.MockHost, codec *mockstream.MockCodec, cfg protector.NetworkConfig, ps proposal.Store) {
				cfg.AddPeer(context.Background(), peer1, test.GeneratePeerMultiaddrs(t, peer1))
				cfg.SetNetworkState(context.Background(), protectorpb.NetworkState_PROTECTED)

				nodeID := &pb.NodeIdentity{
					PeerId:   []byte(peer2),
					PeerAddr: test.GeneratePeerMultiaddr(t, peer2).Bytes(),
				}

				streamtest.ExpectDecodeNodeID(t, codec, nodeID)
				streamtest.ExpectEncodeAck(t, codec, nil)
			},
			func(t *testing.T, cfg protector.NetworkConfig, s proposal.Store) {
				allowed := cfg.IsAllowed(context.Background(), peer2)
				assert.False(t, allowed)

				r, err := s.Get(context.Background(), peer2)
				require.NoError(t, err, "s.Get()")
				require.NotNil(t, r)
				assert.Equal(t, proposal.AddNode, r.Type)
			},
			nil,
		},
		{
			"after-bootstrap-remove-node",
			peer1,
			func(t *testing.T, _ *gomock.Controller, h *mocks.MockHost, codec *mockstream.MockCodec, cfg protector.NetworkConfig, ps proposal.Store) {
				cfg.AddPeer(context.Background(), peer1, test.GeneratePeerMultiaddrs(t, peer1))
				cfg.SetNetworkState(context.Background(), protectorpb.NetworkState_PROTECTED)

				nodeID := &pb.NodeIdentity{
					PeerId:   []byte(peer1),
					PeerAddr: test.GeneratePeerMultiaddr(t, peer1).Bytes(),
				}

				streamtest.ExpectDecodeNodeID(t, codec, nodeID)
				streamtest.ExpectEncodeAck(t, codec, nil)

				h.EXPECT().ID().Return(hostID)
				h.EXPECT().NewStream(gomock.Any(), peer1, protocol.PrivateCoordinatedProposePID).Return(nil, errors.New("no stream"))
			},
			func(t *testing.T, cfg protector.NetworkConfig, s proposal.Store) {
				allowed := cfg.IsAllowed(context.Background(), peer1)
				assert.True(t, allowed)

				r, err := s.Get(context.Background(), peer1)
				require.NoError(t, err, "s.Get()")
				require.NotNil(t, r)
				assert.Equal(t, proposal.RemoveNode, r.Type)
			},
			nil,
		},
	}

	for _, tt := range testCases {
		tt.Run(t, func(handler *protocol.CoordinatorHandler) streamutil.AutoCloseHandler {
			return handler.HandlePropose
		})
	}
}

func TestCoordinator_HandleVote(t *testing.T) {
	peer1Key := test.GeneratePrivateKey(t)
	peer1 := test.GetPeerIDFromKey(t, peer1Key)
	peer2Key := test.GeneratePrivateKey(t)
	peer2 := test.GetPeerIDFromKey(t, peer2Key)
	hostID := test.GeneratePeerID(t)

	removePeer1Req := proposaltest.NewRemoveRequest(t, peer1)

	testCases := []CoordinatorHandleTestCase{
		{
			"unauthorized",
			peer1,
			func(t *testing.T, _ *gomock.Controller, h *mocks.MockHost, codec *mockstream.MockCodec, cfg protector.NetworkConfig, ps proposal.Store) {
				cfg.SetNetworkState(context.Background(), protectorpb.NetworkState_PROTECTED)
			},
			func(*testing.T, protector.NetworkConfig, proposal.Store) {},
			protector.ErrConnectionRefused,
		},
		{
			"decode-error",
			peer1,
			func(t *testing.T, _ *gomock.Controller, h *mocks.MockHost, codec *mockstream.MockCodec, cfg protector.NetworkConfig, ps proposal.Store) {
				cfg.AddPeer(context.Background(), peer1, test.GeneratePeerMultiaddrs(t, peer1))
				cfg.SetNetworkState(context.Background(), protectorpb.NetworkState_PROTECTED)

				codec.EXPECT().Decode(gomock.Any()).Return(errors.New("fatal system failure"))
			},
			func(*testing.T, protector.NetworkConfig, proposal.Store) {},
			protector.ErrConnectionRefused,
		},
		{
			"during-bootstrap-reject",
			peer1,
			func(t *testing.T, _ *gomock.Controller, h *mocks.MockHost, codec *mockstream.MockCodec, cfg protector.NetworkConfig, ps proposal.Store) {
				streamtest.ExpectEncodeAck(t, codec, protocol.ErrInvalidOperation)
			},
			func(*testing.T, protector.NetworkConfig, proposal.Store) {},
			nil,
		},
		{
			"invalid-vote",
			peer1,
			func(t *testing.T, _ *gomock.Controller, h *mocks.MockHost, codec *mockstream.MockCodec, cfg protector.NetworkConfig, ps proposal.Store) {
				ctx := context.Background()
				cfg.AddPeer(ctx, peer1, test.GeneratePeerMultiaddrs(t, peer1))
				cfg.SetNetworkState(ctx, protectorpb.NetworkState_PROTECTED)

				ps.AddRequest(ctx, removePeer1Req)

				v, _ := proposal.NewVote(ctx, peer1Key, removePeer1Req)
				v.Signature.Signature = v.Signature.Signature[:10]

				streamtest.ExpectDecodeVote(t, codec, v.ToProtoVote())
				streamtest.ExpectEncodeAck(t, codec, proposal.ErrInvalidSignature)
			},
			func(t *testing.T, cfg protector.NetworkConfig, ps proposal.Store) {
				v, _ := ps.GetVotes(context.Background(), peer1)
				assert.Len(t, v, 0)
			},
			nil,
		},
		{
			"missing-request",
			peer1,
			func(t *testing.T, _ *gomock.Controller, h *mocks.MockHost, codec *mockstream.MockCodec, cfg protector.NetworkConfig, ps proposal.Store) {
				ctx := context.Background()
				cfg.AddPeer(ctx, peer1, test.GeneratePeerMultiaddrs(t, peer1))
				cfg.SetNetworkState(ctx, protectorpb.NetworkState_PROTECTED)

				v, _ := proposal.NewVote(ctx, peer1Key, removePeer1Req)

				streamtest.ExpectDecodeVote(t, codec, v.ToProtoVote())
				streamtest.ExpectEncodeAck(t, codec, proposal.ErrMissingRequest)
			},
			func(t *testing.T, cfg protector.NetworkConfig, ps proposal.Store) {
				v, _ := ps.GetVotes(context.Background(), peer1)
				assert.Len(t, v, 0)
			},
			nil,
		},
		{
			"vote-threshold-not-reached",
			peer1,
			func(t *testing.T, _ *gomock.Controller, h *mocks.MockHost, codec *mockstream.MockCodec, cfg protector.NetworkConfig, ps proposal.Store) {
				ctx := context.Background()
				cfg.AddPeer(ctx, peer1, test.GeneratePeerMultiaddrs(t, peer1))
				cfg.AddPeer(ctx, peer2, test.GeneratePeerMultiaddrs(t, peer2))
				cfg.SetNetworkState(ctx, protectorpb.NetworkState_PROTECTED)

				ps.AddRequest(ctx, removePeer1Req)

				h.EXPECT().ID().Return(hostID).AnyTimes()

				v, _ := proposal.NewVote(ctx, peer1Key, removePeer1Req)
				streamtest.ExpectDecodeVote(t, codec, v.ToProtoVote())
				streamtest.ExpectEncodeAck(t, codec, nil)
			},
			func(t *testing.T, cfg protector.NetworkConfig, ps proposal.Store) {
				assert.True(t, cfg.IsAllowed(context.Background(), peer1))

				v, err := ps.GetVotes(context.Background(), peer1)
				require.NoError(t, err)
				assert.Len(t, v, 1)
			},
			nil,
		},
		{
			"vote-threshold-reached",
			peer1,
			func(t *testing.T, ctrl *gomock.Controller, h *mocks.MockHost, codec *mockstream.MockCodec, cfg protector.NetworkConfig, ps proposal.Store) {
				ctx := context.Background()
				cfg.AddPeer(ctx, peer1, test.GeneratePeerMultiaddrs(t, peer1))
				cfg.AddPeer(ctx, peer2, test.GeneratePeerMultiaddrs(t, peer2))
				cfg.SetNetworkState(ctx, protectorpb.NetworkState_PROTECTED)

				ps.AddRequest(ctx, removePeer1Req)

				network := mocks.NewMockNetwork(ctrl)
				network.EXPECT().Conns().Return(nil)

				h.EXPECT().ID().Return(hostID).AnyTimes()
				h.EXPECT().Network().Return(network)
				h.EXPECT().NewStream(gomock.Any(), peer2, protocol.PrivateCoordinatedConfigPID).Return(nil, errors.New("no stream"))

				v, _ := proposal.NewVote(ctx, peer2Key, removePeer1Req)
				streamtest.ExpectDecodeVote(t, codec, v.ToProtoVote())
				streamtest.ExpectEncodeAck(t, codec, nil)
			},
			func(t *testing.T, cfg protector.NetworkConfig, ps proposal.Store) {
				assert.False(t, cfg.IsAllowed(context.Background(), peer1))

				v, _ := ps.GetVotes(context.Background(), peer1)
				assert.Len(t, v, 0)

				r, _ := ps.Get(context.Background(), peer1)
				assert.Nil(t, r)
			},
			nil,
		},
	}

	for _, tt := range testCases {
		tt.Run(t, func(handler *protocol.CoordinatorHandler) streamutil.AutoCloseHandler {
			return handler.HandleVote
		})
	}
}

func TestCoordinator_SendNetworkConfig(t *testing.T) {
	hostID := test.GeneratePeerID(t)
	peer1 := test.GeneratePeerID(t)
	peer2 := test.GeneratePeerID(t)

	testCases := []struct {
		name      string
		configure func(*testing.T, *gomock.Controller, protector.NetworkConfig, *mockstream.MockProvider)
	}{{
		"no-participants",
		func(*testing.T, *gomock.Controller, protector.NetworkConfig, *mockstream.MockProvider) {},
	}, {
		"stream-error",
		func(t *testing.T, ctrl *gomock.Controller, cfg protector.NetworkConfig, p *mockstream.MockProvider) {
			cfg.AddPeer(context.Background(), peer1, test.GeneratePeerMultiaddrs(t, peer1))

			p.EXPECT().NewStream(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errors.New("no stream"))
		},
	}, {
		"send-config",
		func(t *testing.T, ctrl *gomock.Controller, cfg protector.NetworkConfig, p *mockstream.MockProvider) {
			cfg.AddPeer(context.Background(), peer1, test.GeneratePeerMultiaddrs(t, peer1))
			cfg.AddPeer(context.Background(), peer2, test.GeneratePeerMultiaddrs(t, peer2))

			cfgPb := cfg.Copy(context.Background())
			codec := mockstream.NewMockCodec(ctrl)
			codec.EXPECT().Encode(&cfgPb).Times(2)

			stream := mockstream.NewMockStream(ctrl)
			stream.EXPECT().Close().Times(2)
			stream.EXPECT().Codec().Return(codec).Times(2)

			p.EXPECT().NewStream(gomock.Any(), gomock.Any(), gomock.Any()).Return(stream, nil).Times(2)
		},
	}, {
		"peer-and-protocol-id",
		func(t *testing.T, ctrl *gomock.Controller, cfg protector.NetworkConfig, p *mockstream.MockProvider) {
			cfg.AddPeer(context.Background(), peer1, test.GeneratePeerMultiaddrs(t, peer1))

			codec := mockstream.NewMockCodec(ctrl)
			codec.EXPECT().Encode(gomock.Any())

			stream := mockstream.NewMockStream(ctrl)
			stream.EXPECT().Close()
			stream.EXPECT().Codec().Return(codec)

			streamtest.ExpectStreamPeerAndProtocol(
				t,
				p,
				peer1,
				protocol.PrivateCoordinatedConfigPID,
				stream,
				nil,
			)
		},
	}}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			host := mocks.NewMockHost(ctrl)
			expectCoordinatorHost(host)
			host.EXPECT().ID().Return(hostID).AnyTimes()

			networkConfig := protectortest.NewTestNetworkConfig(t, protectorpb.NetworkState_BOOTSTRAP)
			networkConfig.AddPeer(context.Background(), hostID, test.GeneratePeerMultiaddrs(t, hostID))

			streamProvider := mockstream.NewMockProvider(ctrl)
			tt.configure(t, ctrl, networkConfig, streamProvider)

			handler := protocol.NewCoordinatorHandler(
				host,
				streamProvider,
				networkConfig,
				nil,
			).(*protocol.CoordinatorHandler)

			handler.SendNetworkConfig(context.Background())
		})
	}
}

func TestCoordinator_SendProposal(t *testing.T) {
	hostID := test.GeneratePeerID(t)
	peer1 := test.GeneratePeerID(t)
	peer2 := test.GeneratePeerID(t)

	req := proposaltest.NewRemoveRequest(t, peer1)

	testCases := []struct {
		name      string
		configure func(*testing.T, *gomock.Controller, protector.NetworkConfig, *mockstream.MockProvider)
	}{{
		"no-participants",
		func(*testing.T, *gomock.Controller, protector.NetworkConfig, *mockstream.MockProvider) {},
	}, {
		"stream-error",
		func(t *testing.T, ctrl *gomock.Controller, cfg protector.NetworkConfig, p *mockstream.MockProvider) {
			cfg.AddPeer(context.Background(), peer1, test.GeneratePeerMultiaddrs(t, peer1))

			p.EXPECT().NewStream(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errors.New("no stream"))
		},
	}, {
		"send-proposal",
		func(t *testing.T, ctrl *gomock.Controller, cfg protector.NetworkConfig, p *mockstream.MockProvider) {
			cfg.AddPeer(context.Background(), peer1, test.GeneratePeerMultiaddrs(t, peer1))
			cfg.AddPeer(context.Background(), peer2, test.GeneratePeerMultiaddrs(t, peer2))

			codec := mockstream.NewMockCodec(ctrl)
			codec.EXPECT().Encode(req.ToUpdateProposal()).Times(2)

			stream := mockstream.NewMockStream(ctrl)
			stream.EXPECT().Close().Times(2)
			stream.EXPECT().Codec().Return(codec).Times(2)

			p.EXPECT().NewStream(gomock.Any(), gomock.Any(), gomock.Any()).Return(stream, nil).Times(2)
		},
	}, {
		"peer-and-protocol-id",
		func(t *testing.T, ctrl *gomock.Controller, cfg protector.NetworkConfig, p *mockstream.MockProvider) {
			cfg.AddPeer(context.Background(), peer1, test.GeneratePeerMultiaddrs(t, peer1))

			codec := mockstream.NewMockCodec(ctrl)
			codec.EXPECT().Encode(gomock.Any())

			stream := mockstream.NewMockStream(ctrl)
			stream.EXPECT().Close()
			stream.EXPECT().Codec().Return(codec)

			streamtest.ExpectStreamPeerAndProtocol(
				t,
				p,
				peer1,
				protocol.PrivateCoordinatedProposePID,
				stream,
				nil,
			)
		},
	}}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			host := mocks.NewMockHost(ctrl)
			expectCoordinatorHost(host)
			host.EXPECT().ID().Return(hostID).AnyTimes()

			networkConfig := protectortest.NewTestNetworkConfig(t, protectorpb.NetworkState_PROTECTED)
			networkConfig.AddPeer(context.Background(), hostID, test.GeneratePeerMultiaddrs(t, hostID))

			streamProvider := mockstream.NewMockProvider(ctrl)
			tt.configure(t, ctrl, networkConfig, streamProvider)

			handler := protocol.NewCoordinatorHandler(
				host,
				streamProvider,
				networkConfig,
				nil,
			).(*protocol.CoordinatorHandler)

			handler.SendProposal(context.Background(), req)
		})
	}
}

func TestCoordinator_AddNode(t *testing.T) {
	coordinatorID := test.GeneratePeerID(t)
	peer1 := test.GeneratePeerID(t)
	peer1Addrs := []multiaddr.Multiaddr{test.GeneratePeerMultiaddr(t, peer1)}

	testCases := []struct {
		name        string
		addNodeID   peer.ID
		addNodeAddr multiaddr.Multiaddr
		configure   func(*testing.T, *gomock.Controller, *mocks.MockHost, protector.NetworkConfig, *mockstream.MockProvider)
		validate    func(*testing.T, protector.NetworkConfig)
		err         error
	}{{
		"node-addr-missing",
		peer1,
		nil,
		func(_ *testing.T, _ *gomock.Controller, h *mocks.MockHost, cfg protector.NetworkConfig, _ *mockstream.MockProvider) {
			// If the peer store doesn't have an address for the node,
			// and it wasn't provided we reject the request.
			peerStore := peerstore.NewPeerstore()
			h.EXPECT().Peerstore().Return(peerStore)
		},
		func(t *testing.T, cfg protector.NetworkConfig) {
			assert.False(t, cfg.IsAllowed(context.Background(), peer1))
		},
		protocol.ErrUnknownNode,
	}, {
		"node-already-white-listed",
		peer1,
		nil,
		func(_ *testing.T, _ *gomock.Controller, _ *mocks.MockHost, cfg protector.NetworkConfig, _ *mockstream.MockProvider) {
			// If the node is already white-listed we shouldn't notify participants.
			cfg.AddPeer(context.Background(), peer1, peer1Addrs)
		},
		func(t *testing.T, cfg protector.NetworkConfig) {
			assert.True(t, cfg.IsAllowed(context.Background(), peer1))
		},
		nil,
	}, {
		"new-node-added-from-peerstore",
		peer1,
		nil,
		func(t *testing.T, ctrl *gomock.Controller, h *mocks.MockHost, cfg protector.NetworkConfig, p *mockstream.MockProvider) {
			peerStore := peerstore.NewPeerstore()
			peerStore.AddAddrs(peer1, peer1Addrs, peerstore.PermanentAddrTTL)
			h.EXPECT().Peerstore().Return(peerStore)

			codec := mockstream.NewMockCodec(ctrl)
			streamtest.ExpectEncodeAllowed(t, codec, peer1)

			stream := mockstream.NewMockStream(ctrl)
			stream.EXPECT().Close()
			stream.EXPECT().Codec().Return(codec)

			p.EXPECT().NewStream(gomock.Any(), gomock.Any(), gomock.Any()).Return(stream, nil)
		},
		func(t *testing.T, cfg protector.NetworkConfig) {
			assert.True(t, cfg.IsAllowed(context.Background(), peer1))
		},
		nil,
	}, {
		"new-node-added-from-addr",
		peer1,
		peer1Addrs[0],
		func(t *testing.T, ctrl *gomock.Controller, h *mocks.MockHost, cfg protector.NetworkConfig, p *mockstream.MockProvider) {
			peerStore := peerstore.NewPeerstore()
			h.EXPECT().Peerstore().Return(peerStore)

			codec := mockstream.NewMockCodec(ctrl)
			streamtest.ExpectEncodeAllowed(t, codec, peer1)

			stream := mockstream.NewMockStream(ctrl)
			stream.EXPECT().Close()
			stream.EXPECT().Codec().Return(codec)

			p.EXPECT().NewStream(gomock.Any(), gomock.Any(), gomock.Any()).Return(stream, nil)
		},
		func(t *testing.T, cfg protector.NetworkConfig) {
			assert.True(t, cfg.IsAllowed(context.Background(), peer1))
		},
		nil,
	}, {
		"new-node-with-addr-and-peerstore",
		peer1,
		peer1Addrs[0],
		func(t *testing.T, ctrl *gomock.Controller, h *mocks.MockHost, cfg protector.NetworkConfig, p *mockstream.MockProvider) {
			peerStore := peerstore.NewPeerstore()
			previousAddr := test.GeneratePeerMultiaddr(t, peer1)
			peerStore.AddAddr(peer1, previousAddr, peerstore.PermanentAddrTTL)
			h.EXPECT().Peerstore().Return(peerStore)

			codec := mockstream.NewMockCodec(ctrl)
			streamtest.ExpectEncodeAllowed(t, codec, peer1)

			stream := mockstream.NewMockStream(ctrl)
			stream.EXPECT().Close()
			stream.EXPECT().Codec().Return(codec)

			p.EXPECT().NewStream(gomock.Any(), gomock.Any(), gomock.Any()).Return(stream, nil)
		},
		func(t *testing.T, cfg protector.NetworkConfig) {
			assert.True(t, cfg.IsAllowed(context.Background(), peer1))
			assert.Len(t, cfg.AllowedAddrs(context.Background(), peer1), 2)
		},
		nil,
	}}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			host := mocks.NewMockHost(ctrl)
			expectCoordinatorHost(host)
			host.EXPECT().ID().AnyTimes().Return(coordinatorID)

			cfg := protectortest.NewTestNetworkConfig(t, protectorpb.NetworkState_BOOTSTRAP)
			cfg.AddPeer(context.Background(), coordinatorID, test.GeneratePeerMultiaddrs(t, coordinatorID))

			prov := mockstream.NewMockProvider(ctrl)

			tt.configure(t, ctrl, host, cfg, prov)

			handler := protocol.NewCoordinatorHandler(host, prov, cfg, nil)
			err := handler.AddNode(context.Background(), tt.addNodeID, tt.addNodeAddr, []byte("I'm batman"))

			if tt.err != nil {
				assert.EqualError(t, err, tt.err.Error())
			} else {
				require.NoError(t, err)
				tt.validate(t, cfg)
			}
		})
	}
}

func TestCoordinator_RemoveNode(t *testing.T) {
	coordinatorID := test.GeneratePeerID(t)
	peer1 := test.GeneratePeerID(t)
	peer2 := test.GeneratePeerID(t)
	peer3 := test.GeneratePeerID(t)

	testCases := []struct {
		name         string
		removeNodeID peer.ID
		configure    func(*testing.T, *gomock.Controller, *mocks.MockHost, protector.NetworkConfig, *mockstream.MockProvider)
		validate     func(*testing.T, protector.NetworkConfig)
		err          error
	}{{
		"peer-not-in-network",
		peer3,
		func(*testing.T, *gomock.Controller, *mocks.MockHost, protector.NetworkConfig, *mockstream.MockProvider) {
		},
		func(t *testing.T, cfg protector.NetworkConfig) {
			assert.False(t, cfg.IsAllowed(context.Background(), peer3))
		},
		nil,
	}, {
		"remove-coordinator",
		coordinatorID,
		func(*testing.T, *gomock.Controller, *mocks.MockHost, protector.NetworkConfig, *mockstream.MockProvider) {
		},
		func(t *testing.T, cfg protector.NetworkConfig) {
			assert.True(t, cfg.IsAllowed(context.Background(), coordinatorID))
		},
		protocol.ErrInvalidOperation,
	}, {
		"remove-peer",
		peer3,
		func(t *testing.T, ctrl *gomock.Controller, h *mocks.MockHost, cfg protector.NetworkConfig, p *mockstream.MockProvider) {
			cfg.AddPeer(context.Background(), peer1, test.GeneratePeerMultiaddrs(t, peer1))
			cfg.AddPeer(context.Background(), peer2, test.GeneratePeerMultiaddrs(t, peer2))
			cfg.AddPeer(context.Background(), peer3, test.GeneratePeerMultiaddrs(t, peer3))

			codec := mockstream.NewMockCodec(ctrl)
			streamtest.ExpectEncodeAllowed(t, codec, peer1)
			streamtest.ExpectEncodeAllowed(t, codec, peer2)

			stream := mockstream.NewMockStream(ctrl)
			stream.EXPECT().Close().Times(2)
			stream.EXPECT().Codec().Return(codec).Times(2)

			p.EXPECT().NewStream(gomock.Any(), gomock.Any(), gomock.Any()).Return(stream, nil).Times(2)

			// After removing the peer, we should disconnect from it.
			conn1 := mocks.NewMockConn(ctrl)
			conn1.EXPECT().RemotePeer().Return(peer1)

			conn3 := mocks.NewMockConn(ctrl)
			conn3.EXPECT().RemotePeer().Return(peer3)
			conn3.EXPECT().Close()

			network := mocks.NewMockNetwork(ctrl)
			network.EXPECT().Conns().Return([]inet.Conn{conn1, conn3})

			h.EXPECT().Network().Return(network)
		},
		func(t *testing.T, cfg protector.NetworkConfig) {
			assert.True(t, cfg.IsAllowed(context.Background(), peer1))
			assert.True(t, cfg.IsAllowed(context.Background(), peer2))
			assert.False(t, cfg.IsAllowed(context.Background(), peer3))
		},
		nil,
	}}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			host := mocks.NewMockHost(ctrl)
			expectCoordinatorHost(host)
			host.EXPECT().ID().AnyTimes().Return(coordinatorID)

			cfg := protectortest.NewTestNetworkConfig(t, protectorpb.NetworkState_PROTECTED)
			cfg.AddPeer(context.Background(), coordinatorID, test.GeneratePeerMultiaddrs(t, coordinatorID))

			prov := mockstream.NewMockProvider(ctrl)

			tt.configure(t, ctrl, host, cfg, prov)

			handler := protocol.NewCoordinatorHandler(host, prov, cfg, nil)
			err := handler.RemoveNode(context.Background(), tt.removeNodeID)

			if tt.err != nil {
				assert.EqualError(t, err, tt.err.Error())
			} else {
				assert.NoError(t, err)
				tt.validate(t, cfg)
			}
		})
	}
}

func TestCoordinator_Accept(t *testing.T) {
	coordinatorID := test.GeneratePeerID(t)
	peer1 := test.GeneratePeerID(t)
	peer1Addrs := test.GeneratePeerMultiaddrs(t, peer1)
	peer2 := test.GeneratePeerID(t)

	testCases := []struct {
		name      string
		acceptID  peer.ID
		configure func(*testing.T, *gomock.Controller, *mocks.MockHost, *mockstream.MockProvider, *mockproposal.MockStore)
		validate  func(*testing.T, protector.NetworkConfig)
		err       error
	}{{
		"proposal-missing",
		peer2,
		func(_ *testing.T, _ *gomock.Controller, _ *mocks.MockHost, _ *mockstream.MockProvider, store *mockproposal.MockStore) {
			store.EXPECT().Get(gomock.Any(), peer2).Return(nil, nil)
		},
		func(t *testing.T, cfg protector.NetworkConfig) {
			assert.False(t, cfg.IsAllowed(context.Background(), peer2))
		},
		proposal.ErrMissingRequest,
	}, {
		"proposal-store-err",
		peer2,
		func(_ *testing.T, _ *gomock.Controller, _ *mocks.MockHost, _ *mockstream.MockProvider, store *mockproposal.MockStore) {
			r := proposaltest.NewAddRequest(t, peer2)
			store.EXPECT().Get(gomock.Any(), peer2).Return(r, nil)
			store.EXPECT().Remove(gomock.Any(), peer2).Return(errors.New("fatal error"))
		},
		func(t *testing.T, cfg protector.NetworkConfig) {
			assert.False(t, cfg.IsAllowed(context.Background(), peer2))
		},
		errors.New("fatal error"),
	}, {
		"add-node",
		peer2,
		func(_ *testing.T, _ *gomock.Controller, h *mocks.MockHost, prov *mockstream.MockProvider, store *mockproposal.MockStore) {
			r := proposaltest.NewAddRequest(t, peer2)
			store.EXPECT().Get(gomock.Any(), peer2).Return(r, nil)
			store.EXPECT().Remove(gomock.Any(), peer2)

			pstore := peerstore.NewPeerstore()
			h.EXPECT().Peerstore().Return(pstore)

			prov.EXPECT().NewStream(gomock.Any(), h, gomock.Any()).Return(nil, errors.New("no stream")).Times(2)
		},
		func(t *testing.T, cfg protector.NetworkConfig) {
			assert.True(t, cfg.IsAllowed(context.Background(), peer2))
		},
		nil,
	}, {
		"remove-node",
		peer1,
		func(_ *testing.T, ctrl *gomock.Controller, h *mocks.MockHost, prov *mockstream.MockProvider, store *mockproposal.MockStore) {
			r := proposaltest.NewRemoveRequest(t, peer1)
			store.EXPECT().Get(gomock.Any(), peer1).Return(r, nil)
			store.EXPECT().Remove(gomock.Any(), peer1)

			network := mocks.NewMockNetwork(ctrl)
			network.EXPECT().Conns().Return(nil)

			h.EXPECT().Network().Return(network)
		},
		func(t *testing.T, cfg protector.NetworkConfig) {
			assert.False(t, cfg.IsAllowed(context.Background(), peer2))
		},
		nil,
	}}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			host := mocks.NewMockHost(ctrl)
			expectCoordinatorHost(host)
			host.EXPECT().ID().AnyTimes().Return(coordinatorID)

			cfg := protectortest.NewTestNetworkConfig(t, protectorpb.NetworkState_PROTECTED)
			cfg.AddPeer(context.Background(), coordinatorID, test.GeneratePeerMultiaddrs(t, coordinatorID))
			cfg.AddPeer(context.Background(), peer1, peer1Addrs)

			store := mockproposal.NewMockStore(ctrl)
			prov := mockstream.NewMockProvider(ctrl)

			tt.configure(t, ctrl, host, prov, store)

			handler := protocol.NewCoordinatorHandler(host, prov, cfg, store)
			err := handler.Accept(context.Background(), tt.acceptID)

			if tt.err != nil {
				assert.EqualError(t, err, tt.err.Error())
			} else {
				assert.NoError(t, err)
				tt.validate(t, cfg)
			}
		})
	}
}

func TestCoordinator_Reject(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	peerID := test.GeneratePeerID(t)

	host := mocks.NewMockHost(ctrl)
	expectCoordinatorHost(host)

	store := mockproposal.NewMockStore(ctrl)
	store.EXPECT().Remove(gomock.Any(), peerID)

	handler := protocol.NewCoordinatorHandler(host, nil, nil, store)

	err := handler.Reject(ctx, peerID)
	require.NoError(t, err, "handler.Reject()")
}

func TestCoordinator_CompleteBootstrap(t *testing.T) {
	testCases := []struct {
		name      string
		configure func(*gomock.Controller, *mocks.MockHost, protector.NetworkConfig, *mockstream.MockProvider)
	}{{
		"already-completed",
		func(_ *gomock.Controller, _ *mocks.MockHost, cfg protector.NetworkConfig, _ *mockstream.MockProvider) {
			cfg.SetNetworkState(context.Background(), protectorpb.NetworkState_PROTECTED)
		},
	}, {
		"set-network-state",
		func(ctrl *gomock.Controller, h *mocks.MockHost, _ protector.NetworkConfig, _ *mockstream.MockProvider) {
			network := mocks.NewMockNetwork(ctrl)
			network.EXPECT().Conns().Return([]inet.Conn{})
			h.EXPECT().Network().Return(network)
		},
	}, {
		"notify-participants",
		func(ctrl *gomock.Controller, h *mocks.MockHost, cfg protector.NetworkConfig, prov *mockstream.MockProvider) {
			peerID := test.GeneratePeerID(t)
			cfg.AddPeer(context.Background(), peerID, test.GeneratePeerMultiaddrs(t, peerID))

			network := mocks.NewMockNetwork(ctrl)
			network.EXPECT().Conns().Return([]inet.Conn{})
			h.EXPECT().Network().Return(network)

			codec := mockstream.NewMockCodec(ctrl)
			streamtest.ExpectEncodeNetworkState(t, codec, protectorpb.NetworkState_PROTECTED)

			stream := mockstream.NewMockStream(ctrl)
			stream.EXPECT().Codec().Return(codec)
			stream.EXPECT().Close()

			prov.EXPECT().NewStream(gomock.Any(), h, gomock.Any()).Return(stream, nil)
		},
	}, {
		"disconnect-non-authorized-nodes",
		func(ctrl *gomock.Controller, h *mocks.MockHost, _ protector.NetworkConfig, _ *mockstream.MockProvider) {
			unauthorizedConn := mocks.NewMockConn(ctrl)
			unauthorizedConn.EXPECT().RemotePeer().Return(test.GeneratePeerID(t))
			unauthorizedConn.EXPECT().Close()

			network := mocks.NewMockNetwork(ctrl)
			network.EXPECT().Conns().Return([]inet.Conn{unauthorizedConn})
			h.EXPECT().Network().Return(network)
		},
	}}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			host := mocks.NewMockHost(ctrl)
			expectCoordinatorHost(host)

			hostID := test.GeneratePeerID(t)
			host.EXPECT().ID().Return(hostID).AnyTimes()

			cfg := protectortest.NewTestNetworkConfig(t, protectorpb.NetworkState_BOOTSTRAP)
			prov := mockstream.NewMockProvider(ctrl)

			tt.configure(ctrl, host, cfg, prov)

			handler := protocol.NewCoordinatorHandler(host, prov, cfg, nil)
			err := handler.CompleteBootstrap(context.Background())
			require.NoError(t, err, "handler.CompleteBootstrap()")
			assert.Equal(t, protectorpb.NetworkState_PROTECTED, cfg.NetworkState(context.Background()))
		})
	}
}
