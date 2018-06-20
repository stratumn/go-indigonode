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

package bootstrap_test

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/pkg/errors"
	"github.com/stratumn/alice/core/protector"
	mockprotector "github.com/stratumn/alice/core/protector/mocks"
	"github.com/stratumn/alice/core/protector/protectortest"
	"github.com/stratumn/alice/core/protocol/bootstrap"
	"github.com/stratumn/alice/core/protocol/bootstrap/bootstraptest"
	"github.com/stratumn/alice/core/protocol/bootstrap/proposal"
	"github.com/stratumn/alice/core/protocol/bootstrap/proposal/mocks"
	"github.com/stratumn/alice/core/streamutil"
	"github.com/stratumn/alice/core/streamutil/mockstream"
	"github.com/stratumn/alice/core/streamutil/streamtest"
	pb "github.com/stratumn/alice/pb/bootstrap"
	protectorpb "github.com/stratumn/alice/pb/protector"
	"github.com/stratumn/alice/test"
	"github.com/stratumn/alice/test/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gx/ipfs/QmWWQ2Txc2c6tqjsBpzg5Ar652cHPGNsQQp2SejkNmkUMb/go-multiaddr"
	inet "gx/ipfs/QmXoz9o2PT3tEzf7hicegwex5UgVP54n3k82K7jrWFyN86/go-libp2p-net"
	netutil "gx/ipfs/Qmb6BsZf6Y3kxffXMNTubGPF1w1bkHtpvhfYbmnwP3NQyw/go-libp2p-netutil"
	bhost "gx/ipfs/Qmc64U41EEB4nPG7wxjEqFwKJajS2f8kk5q2TvUrQf78Xu/go-libp2p-blankhost"
	"gx/ipfs/QmcJukH2sAFjY3HdBKq35WDzWoL3UUu2gt9wdfqZTUyM74/go-libp2p-peer"
	"gx/ipfs/QmdeiKhUy1TVGBaKxt7y1QmBDLBdisSrLJ1x58Eoj4PXUh/go-libp2p-peerstore"
	ihost "gx/ipfs/QmfZTdmunzKzAGJrSvXXQbQ5kLLUiEMX5vdwux7iXkdk7D/go-libp2p-host"
)

func expectSetStreamHandler(host *mocks.MockHost) {
	host.EXPECT().SetStreamHandler(bootstrap.PrivateCoordinatorHandshakePID, gomock.Any()).Times(1)
	host.EXPECT().SetStreamHandler(bootstrap.PrivateCoordinatorProposePID, gomock.Any()).Times(1)
	host.EXPECT().SetStreamHandler(bootstrap.PrivateCoordinatorVotePID, gomock.Any()).Times(1)
}

func TestCoordinator_Close(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	host := mocks.NewMockHost(ctrl)
	expectSetStreamHandler(host)

	handler := bootstrap.NewCoordinatorHandler(host, nil, nil, nil)
	require.NotNil(t, handler)

	host.EXPECT().RemoveStreamHandler(bootstrap.PrivateCoordinatorHandshakePID).Times(1)
	host.EXPECT().RemoveStreamHandler(bootstrap.PrivateCoordinatorProposePID).Times(1)
	host.EXPECT().RemoveStreamHandler(bootstrap.PrivateCoordinatorVotePID).Times(1)

	handler.Close(context.Background())
}

func TestCoordinator_ValidateSender(t *testing.T) {
	ctx := context.Background()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	host := mocks.NewMockHost(ctrl)
	expectSetStreamHandler(host)

	networkCfg := protectortest.NewTestNetworkConfig(t, protectorpb.NetworkState_BOOTSTRAP)
	handler := bootstrap.NewCoordinatorHandler(host, nil, networkCfg, nil).(*bootstrap.CoordinatorHandler)

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

type HandleTestCase struct {
	name       string
	remotePeer peer.ID
	configure  func(*testing.T, *gomock.Controller, *mocks.MockHost, *mockstream.MockCodec, protector.NetworkConfig, proposal.Store)
	validate   func(*testing.T, protector.NetworkConfig, proposal.Store)
	err        error
}

func (ht *HandleTestCase) Run(t *testing.T, h func(*bootstrap.CoordinatorHandler) streamutil.AutoCloseHandler) {
	t.Run(ht.name, func(t *testing.T) {
		ctx := context.Background()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		host := mocks.NewMockHost(ctrl)
		expectSetStreamHandler(host)

		conn := mocks.NewMockConn(ctrl)
		conn.EXPECT().RemotePeer().Return(ht.remotePeer)

		stream := mocks.NewMockStream(ctrl)
		stream.EXPECT().Conn().Return(conn)

		networkCfg := protectortest.NewTestNetworkConfig(t, protectorpb.NetworkState_BOOTSTRAP)
		propStore := proposal.NewInMemoryStore()

		codec := mockstream.NewMockCodec(ctrl)
		ht.configure(t, ctrl, host, codec, networkCfg, propStore)

		handler := bootstrap.NewCoordinatorHandler(
			host,
			streamutil.NewStreamProvider(),
			networkCfg,
			propStore,
		).(*bootstrap.CoordinatorHandler)

		err := h(handler)(ctx, bootstraptest.NewEvent(), stream, codec)
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

	testCases := []HandleTestCase{
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
		tt.Run(t, func(handler *bootstrap.CoordinatorHandler) streamutil.AutoCloseHandler {
			return handler.HandleHandshake
		})
	}
}

func TestCoordinator_HandlePropose(t *testing.T) {
	hostID := test.GeneratePeerID(t)
	peer1 := test.GeneratePeerID(t)
	peer2 := test.GeneratePeerID(t)

	testCases := []HandleTestCase{
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
				h.EXPECT().NewStream(gomock.Any(), peer1, bootstrap.PrivateCoordinatedProposePID).Return(nil, errors.New("no stream"))
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
		tt.Run(t, func(handler *bootstrap.CoordinatorHandler) streamutil.AutoCloseHandler {
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

	removePeer1Req, err := proposal.NewRemoveRequest(&pb.NodeIdentity{PeerId: []byte(peer1)})
	require.NoError(t, err)

	testCases := []HandleTestCase{
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
				streamtest.ExpectEncodeAck(t, codec, bootstrap.ErrInvalidOperation)
			},
			func(*testing.T, protector.NetworkConfig, proposal.Store) {},
			nil,
		},
		{
			"invalid-vote",
			peer1,
			func(t *testing.T, _ *gomock.Controller, h *mocks.MockHost, codec *mockstream.MockCodec, cfg protector.NetworkConfig, ps proposal.Store) {
				cfg.AddPeer(context.Background(), peer1, test.GeneratePeerMultiaddrs(t, peer1))
				cfg.SetNetworkState(context.Background(), protectorpb.NetworkState_PROTECTED)

				ps.AddRequest(context.Background(), removePeer1Req)

				v, _ := proposal.NewVote(peer1Key, removePeer1Req)
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
				cfg.AddPeer(context.Background(), peer1, test.GeneratePeerMultiaddrs(t, peer1))
				cfg.SetNetworkState(context.Background(), protectorpb.NetworkState_PROTECTED)

				v, _ := proposal.NewVote(peer1Key, removePeer1Req)

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
				cfg.AddPeer(context.Background(), peer1, test.GeneratePeerMultiaddrs(t, peer1))
				cfg.AddPeer(context.Background(), peer2, test.GeneratePeerMultiaddrs(t, peer2))
				cfg.SetNetworkState(context.Background(), protectorpb.NetworkState_PROTECTED)

				ps.AddRequest(context.Background(), removePeer1Req)

				h.EXPECT().ID().Return(hostID).AnyTimes()

				v, _ := proposal.NewVote(peer1Key, removePeer1Req)
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
				cfg.AddPeer(context.Background(), peer1, test.GeneratePeerMultiaddrs(t, peer1))
				cfg.AddPeer(context.Background(), peer2, test.GeneratePeerMultiaddrs(t, peer2))
				cfg.SetNetworkState(context.Background(), protectorpb.NetworkState_PROTECTED)

				ps.AddRequest(context.Background(), removePeer1Req)

				network := mocks.NewMockNetwork(ctrl)
				network.EXPECT().Conns().Return(nil)

				h.EXPECT().ID().Return(hostID).AnyTimes()
				h.EXPECT().Network().Return(network)
				h.EXPECT().NewStream(gomock.Any(), peer2, bootstrap.PrivateCoordinatedConfigPID).Return(nil, errors.New("no stream"))

				v, _ := proposal.NewVote(peer2Key, removePeer1Req)
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
		tt.Run(t, func(handler *bootstrap.CoordinatorHandler) streamutil.AutoCloseHandler {
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

			p.EXPECT().NewStream(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
				func(_ context.Context, _ ihost.Host, opts ...streamutil.StreamOption) (streamutil.Stream, error) {
					streamOpts := &streamutil.StreamOptions{}
					for _, opt := range opts {
						opt(streamOpts)
					}

					assert.Equal(t, peer1, streamOpts.PeerID)
					assert.Len(t, streamOpts.PIDs, 1)
					assert.Equal(t, bootstrap.PrivateCoordinatedConfigPID, streamOpts.PIDs[0])
					assert.NotNil(t, streamOpts.Event)

					return stream, nil
				},
			)
		},
	}}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			host := mocks.NewMockHost(ctrl)
			expectSetStreamHandler(host)
			host.EXPECT().ID().Return(hostID).AnyTimes()

			networkConfig := protectortest.NewTestNetworkConfig(t, protectorpb.NetworkState_BOOTSTRAP)
			networkConfig.AddPeer(context.Background(), hostID, test.GeneratePeerMultiaddrs(t, hostID))

			streamProvider := mockstream.NewMockProvider(ctrl)
			tt.configure(t, ctrl, networkConfig, streamProvider)

			handler := bootstrap.NewCoordinatorHandler(
				host,
				streamProvider,
				networkConfig,
				nil,
			).(*bootstrap.CoordinatorHandler)

			handler.SendNetworkConfig(context.Background())
		})
	}
}

func TestCoordinator_SendProposal(t *testing.T) {
	hostID := test.GeneratePeerID(t)
	peer1 := test.GeneratePeerID(t)
	peer2 := test.GeneratePeerID(t)

	req, err := proposal.NewRemoveRequest(&pb.NodeIdentity{PeerId: []byte(peer1)})
	require.NoError(t, err)

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

			p.EXPECT().NewStream(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
				func(_ context.Context, _ ihost.Host, opts ...streamutil.StreamOption) (streamutil.Stream, error) {
					streamOpts := &streamutil.StreamOptions{}
					for _, opt := range opts {
						opt(streamOpts)
					}

					assert.Equal(t, peer1, streamOpts.PeerID)
					assert.Len(t, streamOpts.PIDs, 1)
					assert.Equal(t, bootstrap.PrivateCoordinatedProposePID, streamOpts.PIDs[0])
					assert.NotNil(t, streamOpts.Event)

					return stream, nil
				},
			)
		},
	}}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			host := mocks.NewMockHost(ctrl)
			expectSetStreamHandler(host)
			host.EXPECT().ID().Return(hostID).AnyTimes()

			networkConfig := protectortest.NewTestNetworkConfig(t, protectorpb.NetworkState_PROTECTED)
			networkConfig.AddPeer(context.Background(), hostID, test.GeneratePeerMultiaddrs(t, hostID))

			streamProvider := mockstream.NewMockProvider(ctrl)
			tt.configure(t, ctrl, networkConfig, streamProvider)

			handler := bootstrap.NewCoordinatorHandler(
				host,
				streamProvider,
				networkConfig,
				nil,
			).(*bootstrap.CoordinatorHandler)

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
			h.EXPECT().Peerstore().Times(1).Return(peerStore)
		},
		func(t *testing.T, cfg protector.NetworkConfig) {
			assert.False(t, cfg.IsAllowed(context.Background(), peer1))
		},
		bootstrap.ErrUnknownNode,
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
			h.EXPECT().Peerstore().Times(1).Return(peerStore)

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
			h.EXPECT().Peerstore().Times(1).Return(peerStore)

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
	}}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			host := mocks.NewMockHost(ctrl)
			expectSetStreamHandler(host)
			host.EXPECT().ID().AnyTimes().Return(coordinatorID)

			cfg := protectortest.NewTestNetworkConfig(t, protectorpb.NetworkState_BOOTSTRAP)
			cfg.AddPeer(context.Background(), coordinatorID, test.GeneratePeerMultiaddrs(t, coordinatorID))

			prov := mockstream.NewMockProvider(ctrl)

			tt.configure(t, ctrl, host, cfg, prov)

			handler := bootstrap.NewCoordinatorHandler(host, prov, cfg, nil)
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
		bootstrap.ErrInvalidOperation,
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
			conn1.EXPECT().RemotePeer().Return(peer1).Times(1)

			conn3 := mocks.NewMockConn(ctrl)
			conn3.EXPECT().RemotePeer().Return(peer3).Times(1)
			conn3.EXPECT().Close().Times(1)

			network := mocks.NewMockNetwork(ctrl)
			network.EXPECT().Conns().Return([]inet.Conn{conn1, conn3}).Times(1)

			h.EXPECT().Network().Return(network).Times(1)
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
			expectSetStreamHandler(host)
			host.EXPECT().ID().AnyTimes().Return(coordinatorID)

			cfg := protectortest.NewTestNetworkConfig(t, protectorpb.NetworkState_PROTECTED)
			cfg.AddPeer(context.Background(), coordinatorID, test.GeneratePeerMultiaddrs(t, coordinatorID))

			prov := mockstream.NewMockProvider(ctrl)

			tt.configure(t, ctrl, host, cfg, prov)

			handler := bootstrap.NewCoordinatorHandler(host, prov, cfg, nil)
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
	peer1Addr := test.GeneratePeerMultiaddr(t, peer1)
	peer2 := test.GeneratePeerID(t)

	addPeer1 := &proposal.Request{
		Type:     proposal.AddNode,
		PeerID:   peer1,
		PeerAddr: peer1Addr,
	}

	removePeer2 := &proposal.Request{
		Type:   proposal.RemoveNode,
		PeerID: peer2,
	}

	testCases := []struct {
		name      string
		acceptID  peer.ID
		configure func(*gomock.Controller, *mocks.MockHost, *mockprotector.MockNetworkConfig, *mockproposal.MockStore)
		err       error
	}{{
		"proposal-missing",
		peer1,
		func(ctrl *gomock.Controller, host *mocks.MockHost, cfg *mockprotector.MockNetworkConfig, store *mockproposal.MockStore) {
			store.EXPECT().Get(gomock.Any(), peer1).Times(1).Return(nil, nil)
		},
		proposal.ErrMissingRequest,
	}, {
		"peer-addr-missing",
		peer1,
		func(ctrl *gomock.Controller, host *mocks.MockHost, cfg *mockprotector.MockNetworkConfig, store *mockproposal.MockStore) {
			r := &proposal.Request{
				Type:   proposal.AddNode,
				PeerID: peer1,
			}
			store.EXPECT().Get(gomock.Any(), peer1).Times(1).Return(r, nil)
			store.EXPECT().Remove(gomock.Any(), peer1).Times(1)
		},
		proposal.ErrMissingPeerAddr,
	}, {
		"add-already-added",
		peer1,
		func(ctrl *gomock.Controller, host *mocks.MockHost, cfg *mockprotector.MockNetworkConfig, store *mockproposal.MockStore) {
			store.EXPECT().Get(gomock.Any(), peer1).Times(1).Return(addPeer1, nil)
			store.EXPECT().Remove(gomock.Any(), peer1).Times(1)

			cfg.EXPECT().IsAllowed(gomock.Any(), peer1).Times(1).Return(true)
		},
		nil,
	}, {
		"add-node",
		peer1,
		func(ctrl *gomock.Controller, host *mocks.MockHost, cfg *mockprotector.MockNetworkConfig, store *mockproposal.MockStore) {
			store.EXPECT().Get(gomock.Any(), peer1).Times(1).Return(addPeer1, nil)
			store.EXPECT().Remove(gomock.Any(), peer1).Times(1)

			cfg.EXPECT().IsAllowed(gomock.Any(), peer1).Times(1).Return(false)
			cfg.EXPECT().AddPeer(gomock.Any(), peer1, []multiaddr.Multiaddr{addPeer1.PeerAddr}).Times(1)
			cfg.EXPECT().Copy(gomock.Any()).Times(1)
			cfg.EXPECT().AllowedPeers(gomock.Any()).Times(1).Return([]peer.ID{peer1, peer2})

			stream := mocks.NewMockStream(ctrl)
			stream.EXPECT().Write(gomock.Any()).AnyTimes()
			stream.EXPECT().Conn().Times(2)
			stream.EXPECT().Close().Times(2)

			host.EXPECT().ID().AnyTimes().Return(coordinatorID)
			host.EXPECT().NewStream(gomock.Any(), peer1, bootstrap.PrivateCoordinatedConfigPID).Times(1).Return(stream, nil)
			host.EXPECT().NewStream(gomock.Any(), peer2, bootstrap.PrivateCoordinatedConfigPID).Times(1).Return(stream, nil)
		},
		nil,
	}, {
		"remove-node",
		peer2,
		func(ctrl *gomock.Controller, host *mocks.MockHost, cfg *mockprotector.MockNetworkConfig, store *mockproposal.MockStore) {
			store.EXPECT().Get(gomock.Any(), peer2).Times(1).Return(removePeer2, nil)
			store.EXPECT().Remove(gomock.Any(), peer2).Times(1)

			cfg.EXPECT().IsAllowed(gomock.Any(), peer2).Times(1).Return(true)
			cfg.EXPECT().RemovePeer(gomock.Any(), peer2).Times(1)
			cfg.EXPECT().Copy(gomock.Any()).Times(1)
			cfg.EXPECT().AllowedPeers(gomock.Any()).Times(1).Return([]peer.ID{peer1})

			network := mocks.NewMockNetwork(ctrl)
			network.EXPECT().Conns().Return(nil).Times(1)

			stream := mocks.NewMockStream(ctrl)
			stream.EXPECT().Write(gomock.Any()).AnyTimes()
			stream.EXPECT().Conn().Times(1)
			stream.EXPECT().Close().Times(1)

			host.EXPECT().ID().AnyTimes().Return(coordinatorID)
			host.EXPECT().Network().Return(network).Times(1)
			host.EXPECT().NewStream(gomock.Any(), peer1, bootstrap.PrivateCoordinatedConfigPID).Times(1).Return(stream, nil)
		},
		nil,
	}}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			host := mocks.NewMockHost(ctrl)
			expectSetStreamHandler(host)

			networkConfig := mockprotector.NewMockNetworkConfig(ctrl)
			store := mockproposal.NewMockStore(ctrl)

			tt.configure(ctrl, host, networkConfig, store)

			handler := bootstrap.NewCoordinatorHandler(
				host,
				streamutil.NewStreamProvider(),
				networkConfig,
				store,
			)

			err := handler.Accept(ctx, tt.acceptID)
			if tt.err != nil {
				assert.EqualError(t, err, tt.err.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCoordinator_Reject(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	peerID := test.GeneratePeerID(t)

	host := bhost.NewBlankHost(netutil.GenSwarmNetwork(t, ctx))
	store := mockproposal.NewMockStore(ctrl)
	store.EXPECT().Remove(gomock.Any(), peerID).Times(1)

	handler := bootstrap.NewCoordinatorHandler(
		host,
		streamutil.NewStreamProvider(),
		nil,
		store,
	)

	err := handler.Reject(ctx, peerID)
	require.NoError(t, err, "handler.Reject()")
}

func TestCoordinator_CompleteBootstrap(t *testing.T) {
	peer1 := test.GeneratePeerID(t)
	peer2 := test.GeneratePeerID(t)
	peer3 := test.GeneratePeerID(t)

	testCases := []struct {
		name   string
		expect func(*gomock.Controller, *mocks.MockHost, *mockprotector.MockNetworkConfig)
	}{{
		"already-completed",
		func(_ *gomock.Controller, _ *mocks.MockHost, cfg *mockprotector.MockNetworkConfig) {
			cfg.EXPECT().NetworkState(gomock.Any()).Return(protectorpb.NetworkState_PROTECTED).Times(1)
		},
	}, {
		"set-network-state",
		func(ctrl *gomock.Controller, host *mocks.MockHost, cfg *mockprotector.MockNetworkConfig) {
			cfg.EXPECT().NetworkState(gomock.Any()).Return(protectorpb.NetworkState_BOOTSTRAP).Times(1)
			cfg.EXPECT().SetNetworkState(gomock.Any(), protectorpb.NetworkState_PROTECTED).Times(1)
			cfg.EXPECT().AllowedPeers(gomock.Any()).Return([]peer.ID{}).Times(1)
			cfg.EXPECT().Copy(gomock.Any()).Times(1)

			mockNet := mocks.NewMockNetwork(ctrl)
			mockNet.EXPECT().Conns().Return([]inet.Conn{}).Times(1)

			host.EXPECT().Network().Return(mockNet).Times(1)
		},
	}, {
		"notify-participants",
		func(ctrl *gomock.Controller, host *mocks.MockHost, cfg *mockprotector.MockNetworkConfig) {
			cfg.EXPECT().NetworkState(gomock.Any()).Return(protectorpb.NetworkState_BOOTSTRAP).Times(1)
			cfg.EXPECT().SetNetworkState(gomock.Any(), protectorpb.NetworkState_PROTECTED).Times(1)
			cfg.EXPECT().AllowedPeers(gomock.Any()).Return([]peer.ID{peer1, peer2, peer3}).Times(1)
			cfg.EXPECT().Copy(gomock.Any()).Times(1)

			mockNet := mocks.NewMockNetwork(ctrl)
			mockNet.EXPECT().Conns().Return([]inet.Conn{}).Times(1)

			host.EXPECT().Network().Return(mockNet).Times(1)

			stream := mocks.NewMockStream(ctrl)
			stream.EXPECT().Write(gomock.Any()).AnyTimes()
			stream.EXPECT().Conn().Times(2)
			stream.EXPECT().Close().Times(2)

			host.EXPECT().ID().Return(peer1).AnyTimes()
			host.EXPECT().NewStream(gomock.Any(), peer2, bootstrap.PrivateCoordinatedConfigPID).Times(1).Return(stream, nil)
			host.EXPECT().NewStream(gomock.Any(), peer3, bootstrap.PrivateCoordinatedConfigPID).Times(1).Return(stream, nil)
		},
	}, {
		"disconnect-non-authorized-nodes",
		func(ctrl *gomock.Controller, host *mocks.MockHost, cfg *mockprotector.MockNetworkConfig) {
			cfg.EXPECT().NetworkState(gomock.Any()).Return(protectorpb.NetworkState_BOOTSTRAP).Times(1)
			cfg.EXPECT().SetNetworkState(gomock.Any(), protectorpb.NetworkState_PROTECTED).Times(1)
			cfg.EXPECT().AllowedPeers(gomock.Any()).Return([]peer.ID{peer1, peer2}).Times(1)
			cfg.EXPECT().Copy(gomock.Any()).Times(1)

			stream := mocks.NewMockStream(ctrl)
			stream.EXPECT().Write(gomock.Any()).AnyTimes()
			stream.EXPECT().Conn().Times(1)
			stream.EXPECT().Close().Times(1)

			host.EXPECT().ID().Return(peer1).AnyTimes()
			host.EXPECT().NewStream(gomock.Any(), peer2, bootstrap.PrivateCoordinatedConfigPID).Times(1).Return(stream, nil)

			unauthorizedConn := mocks.NewMockConn(ctrl)
			unauthorizedConn.EXPECT().RemotePeer().Return(peer3).Times(1)
			unauthorizedConn.EXPECT().Close().Times(1)
			cfg.EXPECT().IsAllowed(gomock.Any(), peer3).Return(false).Times(1)

			mockNet := mocks.NewMockNetwork(ctrl)
			mockNet.EXPECT().Conns().Return([]inet.Conn{unauthorizedConn}).Times(1)

			host.EXPECT().Network().Return(mockNet).Times(1)
		},
	}}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			host := mocks.NewMockHost(ctrl)
			expectSetStreamHandler(host)

			networkConfig := mockprotector.NewMockNetworkConfig(ctrl)

			tt.expect(ctrl, host, networkConfig)

			handler := bootstrap.NewCoordinatorHandler(
				host,
				streamutil.NewStreamProvider(),
				networkConfig,
				nil,
			)

			err := handler.CompleteBootstrap(ctx)
			require.NoError(t, err, "handler.CompleteBootstrap()")
		})
	}
}
