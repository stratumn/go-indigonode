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
	bootstrappb "github.com/stratumn/go-node/core/app/bootstrap/pb"
	"github.com/stratumn/go-node/core/app/bootstrap/protocol"
	"github.com/stratumn/go-node/core/app/bootstrap/protocol/proposal"
	"github.com/stratumn/go-node/core/app/bootstrap/protocol/proposal/mocks"
	"github.com/stratumn/go-node/core/app/bootstrap/protocol/proposaltest"
	"github.com/stratumn/go-node/core/protector"
	protectorpb "github.com/stratumn/go-node/core/protector/pb"
	"github.com/stratumn/go-node/core/protector/protectortest"
	"github.com/stratumn/go-node/core/streamutil"
	"github.com/stratumn/go-node/core/streamutil/mockstream"
	"github.com/stratumn/go-node/core/streamutil/streamtest"
	"github.com/stratumn/go-node/test"
	"github.com/stratumn/go-node/test/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gx/ipfs/QmQsErDt8Qgw1XrsXf2BpEzDgGWtB1YLsTAARBup5b6B9W/go-libp2p-peer"
	inet "gx/ipfs/QmZNJyx9GGCX4GeuHnLB8fxaxMLs4MjTjHokxfQcCd6Nve/go-libp2p-net"
	"gx/ipfs/Qmda4cPRvSRyox3SqgJN6DfSZGU5TtHufPTp9uXjFj71X6/go-libp2p-peerstore"
	"gx/ipfs/Qmda4cPRvSRyox3SqgJN6DfSZGU5TtHufPTp9uXjFj71X6/go-libp2p-peerstore/pstoremem"
)

func expectCoordinatedHost(host *mocks.MockHost) {
	host.EXPECT().SetStreamHandler(protocol.PrivateCoordinatedConfigPID, gomock.Any())
	host.EXPECT().SetStreamHandler(protocol.PrivateCoordinatedProposePID, gomock.Any())
}

func TestCoordinated_Close(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	host := mocks.NewMockHost(ctrl)
	expectCoordinatedHost(host)

	mode := &protector.NetworkMode{
		CoordinatorID: test.GeneratePeerID(t),
	}
	handler := protocol.NewCoordinatedHandler(host, nil, mode, nil, nil)
	require.NotNil(t, handler)

	host.EXPECT().RemoveStreamHandler(protocol.PrivateCoordinatedConfigPID)
	host.EXPECT().RemoveStreamHandler(protocol.PrivateCoordinatedProposePID)

	handler.Close(context.Background())
}

type CoordinatedHandleTestCase struct {
	name          string
	coordinatorID peer.ID
	remotePeer    peer.ID
	configure     func(*testing.T, *gomock.Controller, *mocks.MockHost, *mockstream.MockCodec)
	validate      func(*testing.T, protector.NetworkConfig, proposal.Store)
	err           error
}

func (ht *CoordinatedHandleTestCase) Run(
	t *testing.T,
	h func(*protocol.CoordinatedHandler) streamutil.AutoCloseHandler,
) {
	t.Run(ht.name, func(t *testing.T) {
		ctx := context.Background()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		host := mocks.NewMockHost(ctrl)
		expectCoordinatedHost(host)

		conn := mocks.NewMockConn(ctrl)
		conn.EXPECT().RemotePeer().Return(ht.remotePeer)

		stream := mocks.NewMockStream(ctrl)
		stream.EXPECT().Conn().Return(conn)

		codec := mockstream.NewMockCodec(ctrl)

		cfg := protectortest.NewTestNetworkConfig(t, protectorpb.NetworkState_BOOTSTRAP)
		s := proposal.NewInMemoryStore()

		ht.configure(t, ctrl, host, codec)

		mode := &protector.NetworkMode{CoordinatorID: ht.coordinatorID}
		handler := protocol.NewCoordinatedHandler(host, nil, mode, cfg, s).(*protocol.CoordinatedHandler)
		err := h(handler)(ctx, test.NewSpan(), stream, codec)

		if ht.err != nil {
			assert.EqualError(t, err, ht.err.Error())
		} else {
			require.NoError(t, err)
			ht.validate(t, cfg, s)
		}
	})
}

func TestCoordinated_HandleConfigUpdate(t *testing.T) {
	coordinatorKey := test.GeneratePrivateKey(t)
	coordinatorID := test.GetPeerIDFromKey(t, coordinatorKey)
	peer1Key := test.GeneratePrivateKey(t)
	peer1 := test.GetPeerIDFromKey(t, peer1Key)

	testCases := []CoordinatedHandleTestCase{
		{
			"reject-not-coordinator",
			coordinatorID,
			peer1,
			func(*testing.T, *gomock.Controller, *mocks.MockHost, *mockstream.MockCodec) {
			},
			func(*testing.T, protector.NetworkConfig, proposal.Store) {},
			protector.ErrConnectionRefused,
		},
		{
			"invalid-config-content",
			coordinatorID,
			coordinatorID,
			func(t *testing.T, _ *gomock.Controller, _ *mocks.MockHost, codec *mockstream.MockCodec) {
				codec.EXPECT().Decode(gomock.Any()).Return(errors.New("BSOD"))
			},
			func(*testing.T, protector.NetworkConfig, proposal.Store) {},
			errors.Cause(errors.New("BSOD")),
		},
		{
			"invalid-config-signature",
			coordinatorID,
			coordinatorID,
			func(t *testing.T, _ *gomock.Controller, _ *mocks.MockHost, codec *mockstream.MockCodec) {
				cfg := &protectorpb.NetworkConfig{
					NetworkState: 42,
				}
				err := cfg.Sign(context.Background(), coordinatorKey)
				require.NoError(t, err, "cfg.Sign()")

				streamtest.ExpectDecodeConfig(t, codec, cfg)
			},
			func(t *testing.T, cfg protector.NetworkConfig, _ proposal.Store) {
				assert.Equal(t, protectorpb.NetworkState_BOOTSTRAP, cfg.NetworkState(context.Background()))
			},
			protectorpb.ErrInvalidNetworkState,
		},
		{
			"invalid-config-signer",
			coordinatorID,
			coordinatorID,
			func(t *testing.T, _ *gomock.Controller, _ *mocks.MockHost, codec *mockstream.MockCodec) {
				cfg := protectortest.NewTestNetworkConfig(
					t,
					protectorpb.NetworkState_PROTECTED,
					peer1,
				).Copy(context.Background())

				err := cfg.Sign(context.Background(), peer1Key)
				require.NoError(t, err, "cfg.Sign()")

				streamtest.ExpectDecodeConfig(t, codec, &cfg)
			},
			func(t *testing.T, cfg protector.NetworkConfig, _ proposal.Store) {
				assert.Equal(t, protectorpb.NetworkState_BOOTSTRAP, cfg.NetworkState(context.Background()))
				assert.False(t, cfg.IsAllowed(context.Background(), peer1))
			},
			protectorpb.ErrInvalidSignature,
		},
		{
			"valid-config",
			coordinatorID,
			coordinatorID,
			func(t *testing.T, ctrl *gomock.Controller, h *mocks.MockHost, codec *mockstream.MockCodec) {
				cfg := protectortest.NewTestNetworkConfig(
					t,
					protectorpb.NetworkState_PROTECTED,
					peer1,
				).Copy(context.Background())

				err := cfg.Sign(context.Background(), coordinatorKey)
				require.NoError(t, err, "cfg.Sign()")

				streamtest.ExpectDecodeConfig(t, codec, &cfg)

				// Disconnect from unauthorized peers
				peer2 := test.GeneratePeerID(t)
				unauthorizedConn := mocks.NewMockConn(ctrl)
				unauthorizedConn.EXPECT().RemotePeer().Return(peer2)
				unauthorizedConn.EXPECT().Close()

				network := mocks.NewMockNetwork(ctrl)
				network.EXPECT().Conns().Return([]inet.Conn{unauthorizedConn})

				h.EXPECT().Network().Return(network)
			},
			func(t *testing.T, cfg protector.NetworkConfig, _ proposal.Store) {
				assert.Equal(t, protectorpb.NetworkState_PROTECTED, cfg.NetworkState(context.Background()))
				assert.True(t, cfg.IsAllowed(context.Background(), peer1))
			},
			nil,
		},
	}

	for _, tt := range testCases {
		tt.Run(t, func(handler *protocol.CoordinatedHandler) streamutil.AutoCloseHandler {
			return handler.HandleConfigUpdate
		})
	}
}

func TestCoordinated_HandlePropose(t *testing.T) {
	coordinatorKey := test.GeneratePrivateKey(t)
	coordinatorID := test.GetPeerIDFromKey(t, coordinatorKey)
	peer1 := test.GeneratePeerID(t)

	testCases := []CoordinatedHandleTestCase{
		{
			"reject-not-coordinator",
			coordinatorID,
			peer1,
			func(*testing.T, *gomock.Controller, *mocks.MockHost, *mockstream.MockCodec) {
			},
			func(*testing.T, protector.NetworkConfig, proposal.Store) {},
			protector.ErrConnectionRefused,
		},
		{
			"invalid-request",
			coordinatorID,
			coordinatorID,
			func(t *testing.T, _ *gomock.Controller, _ *mocks.MockHost, codec *mockstream.MockCodec) {
				prop := &bootstrappb.UpdateProposal{
					UpdateType: bootstrappb.UpdateType_AddNode,
					NodeDetails: &bootstrappb.NodeIdentity{
						PeerId: []byte("b4tm4n"),
					},
				}
				streamtest.ExpectDecodeUpdateProp(t, codec, prop)
			},
			func(*testing.T, protector.NetworkConfig, proposal.Store) {},
			proposal.ErrInvalidPeerID,
		},
		{
			"valid-request",
			coordinatorID,
			coordinatorID,
			func(t *testing.T, _ *gomock.Controller, _ *mocks.MockHost, codec *mockstream.MockCodec) {
				prop := proposaltest.NewAddRequest(t, peer1)
				streamtest.ExpectDecodeUpdateProp(t, codec, prop.ToUpdateProposal())
			},
			func(t *testing.T, cfg protector.NetworkConfig, s proposal.Store) {
				r, err := s.Get(context.Background(), peer1)
				assert.NoError(t, err)
				assert.NotNil(t, r)
			},
			nil,
		},
	}

	for _, tt := range testCases {
		tt.Run(t, func(handler *protocol.CoordinatedHandler) streamutil.AutoCloseHandler {
			return handler.HandlePropose
		})
	}
}

func TestCoordinated_Handshake(t *testing.T) {
	coordinatorKey := test.GeneratePrivateKey(t)
	coordinatorID := test.GetPeerIDFromKey(t, coordinatorKey)
	coordinatorAddrs := test.GeneratePeerMultiaddrs(t, coordinatorID)

	peer1Key := test.GeneratePrivateKey(t)
	peer1 := test.GetPeerIDFromKey(t, peer1Key)

	testCases := []struct {
		name      string
		configure func(*testing.T, *gomock.Controller, *mocks.MockHost, *mockstream.MockProvider)
		validate  func(*testing.T, protector.NetworkConfig)
		err       error
	}{{
		"coordinator-unavailable",
		func(t *testing.T, _ *gomock.Controller, h *mocks.MockHost, _ *mockstream.MockProvider) {
			h.EXPECT().Connect(gomock.Any(), gomock.Any()).Return(errors.New("no conn"))
		},
		func(t *testing.T, cfg protector.NetworkConfig) {},
		protector.ErrConnectionRefused,
	}, {
		"coordinator-stream-error",
		func(t *testing.T, _ *gomock.Controller, h *mocks.MockHost, p *mockstream.MockProvider) {
			h.EXPECT().Connect(gomock.Any(), gomock.Any()).DoAndReturn(
				func(ctx context.Context, pi peerstore.PeerInfo) error {
					assert.Equal(t, coordinatorID, pi.ID)
					assert.ElementsMatch(t, coordinatorAddrs, pi.Addrs)
					return nil
				})

			streamtest.ExpectStreamPeerAndProtocol(
				t,
				p,
				coordinatorID,
				protocol.PrivateCoordinatorHandshakePID,
				nil,
				errors.New("no stream"),
			)
		},
		func(t *testing.T, cfg protector.NetworkConfig) {},
		protector.ErrConnectionRefused,
	}, {
		"coordinator-invalid-signature",
		func(t *testing.T, ctrl *gomock.Controller, h *mocks.MockHost, p *mockstream.MockProvider) {
			h.EXPECT().Connect(gomock.Any(), gomock.Any())

			cfg := protectortest.NewTestNetworkConfig(
				t,
				protectorpb.NetworkState_PROTECTED,
				peer1,
			).Copy(context.Background())
			cfg.Sign(context.Background(), coordinatorKey)
			cfg.Signature.Signature = cfg.Signature.Signature[0:10]

			codec := mockstream.NewMockCodec(ctrl)
			codec.EXPECT().Encode(&bootstrappb.Hello{})
			streamtest.ExpectDecodeConfig(t, codec, &cfg)

			stream := mockstream.NewMockStream(ctrl)
			stream.EXPECT().Codec().Return(codec).MinTimes(1)
			stream.EXPECT().Close()

			p.EXPECT().NewStream(gomock.Any(), gomock.Any(), gomock.Any()).Return(stream, nil)
		},
		func(t *testing.T, cfg protector.NetworkConfig) {
			assert.Equal(t, protectorpb.NetworkState_BOOTSTRAP, cfg.NetworkState(context.Background()))
			assert.False(t, cfg.IsAllowed(context.Background(), peer1))
		},
		protectorpb.ErrInvalidSignature,
	}, {
		"coordinator-invalid-signer",
		func(t *testing.T, ctrl *gomock.Controller, h *mocks.MockHost, p *mockstream.MockProvider) {
			h.EXPECT().Connect(gomock.Any(), gomock.Any())

			cfg := protectortest.NewTestNetworkConfig(
				t,
				protectorpb.NetworkState_PROTECTED,
				peer1,
			).Copy(context.Background())
			cfg.Sign(context.Background(), peer1Key)

			codec := mockstream.NewMockCodec(ctrl)
			codec.EXPECT().Encode(&bootstrappb.Hello{})
			streamtest.ExpectDecodeConfig(t, codec, &cfg)

			stream := mockstream.NewMockStream(ctrl)
			stream.EXPECT().Codec().Return(codec).MinTimes(1)
			stream.EXPECT().Close()

			p.EXPECT().NewStream(gomock.Any(), gomock.Any(), gomock.Any()).Return(stream, nil)
		},
		func(t *testing.T, cfg protector.NetworkConfig) {
			assert.Equal(t, protectorpb.NetworkState_BOOTSTRAP, cfg.NetworkState(context.Background()))
			assert.False(t, cfg.IsAllowed(context.Background(), peer1))
		},
		protectorpb.ErrInvalidSignature,
	}, {
		"coordinator-invalid-config",
		func(t *testing.T, ctrl *gomock.Controller, h *mocks.MockHost, p *mockstream.MockProvider) {
			h.EXPECT().Connect(gomock.Any(), gomock.Any())

			cfg := protectortest.NewTestNetworkConfig(
				t,
				protectorpb.NetworkState_PROTECTED,
				peer1,
			).Copy(context.Background())
			cfg.NetworkState = 42
			cfg.Sign(context.Background(), coordinatorKey)

			codec := mockstream.NewMockCodec(ctrl)
			codec.EXPECT().Encode(&bootstrappb.Hello{})
			streamtest.ExpectDecodeConfig(t, codec, &cfg)

			stream := mockstream.NewMockStream(ctrl)
			stream.EXPECT().Codec().Return(codec).MinTimes(1)
			stream.EXPECT().Close()

			p.EXPECT().NewStream(gomock.Any(), gomock.Any(), gomock.Any()).Return(stream, nil)
		},
		func(t *testing.T, cfg protector.NetworkConfig) {
			assert.Equal(t, protectorpb.NetworkState_BOOTSTRAP, cfg.NetworkState(context.Background()))
			assert.False(t, cfg.IsAllowed(context.Background(), peer1))
		},
		protectorpb.ErrInvalidNetworkState,
	}, {
		"coordinator-empty-config",
		func(t *testing.T, ctrl *gomock.Controller, h *mocks.MockHost, p *mockstream.MockProvider) {
			h.EXPECT().Connect(gomock.Any(), gomock.Any())

			codec := mockstream.NewMockCodec(ctrl)
			codec.EXPECT().Encode(&bootstrappb.Hello{})
			streamtest.ExpectDecodeConfig(t, codec, &protectorpb.NetworkConfig{})

			stream := mockstream.NewMockStream(ctrl)
			stream.EXPECT().Codec().Return(codec).MinTimes(1)
			stream.EXPECT().Close()

			p.EXPECT().NewStream(gomock.Any(), gomock.Any(), gomock.Any()).Return(stream, nil)

			// The node proposes to be added
			h.EXPECT().ID().Return(peer1)
			h.EXPECT().Addrs().Return(test.GeneratePeerMultiaddrs(t, peer1))

			streamtest.ExpectStreamPeerAndProtocol(
				t,
				p,
				coordinatorID,
				protocol.PrivateCoordinatorProposePID,
				nil,
				errors.New("no stream"),
			)
		},
		func(t *testing.T, cfg protector.NetworkConfig) {
			assert.Equal(t, protectorpb.NetworkState_BOOTSTRAP, cfg.NetworkState(context.Background()))
			assert.True(t, cfg.IsAllowed(context.Background(), coordinatorID))
		},
		errors.New("no stream"),
	}, {
		"coordinator-valid-config",
		func(t *testing.T, ctrl *gomock.Controller, h *mocks.MockHost, p *mockstream.MockProvider) {
			h.EXPECT().Connect(gomock.Any(), gomock.Any())

			cfg := protectortest.NewTestNetworkConfig(
				t,
				protectorpb.NetworkState_PROTECTED,
				coordinatorID,
				peer1,
			).Copy(context.Background())
			cfg.Sign(context.Background(), coordinatorKey)

			codec := mockstream.NewMockCodec(ctrl)
			codec.EXPECT().Encode(&bootstrappb.Hello{})
			streamtest.ExpectDecodeConfig(t, codec, &cfg)

			stream := mockstream.NewMockStream(ctrl)
			stream.EXPECT().Codec().Return(codec).MinTimes(1)
			stream.EXPECT().Close()

			p.EXPECT().NewStream(gomock.Any(), gomock.Any(), gomock.Any()).Return(stream, nil)
		},
		func(t *testing.T, cfg protector.NetworkConfig) {
			assert.Equal(t, protectorpb.NetworkState_PROTECTED, cfg.NetworkState(context.Background()))
			assert.Len(t, cfg.AllowedPeers(context.Background()), 2)
			assert.True(t, cfg.IsAllowed(context.Background(), coordinatorID))
			assert.True(t, cfg.IsAllowed(context.Background(), peer1))
		},
		nil,
	}}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			pstore := pstoremem.NewPeerstore()
			pstore.AddAddrs(coordinatorID, coordinatorAddrs, peerstore.PermanentAddrTTL)

			host := mocks.NewMockHost(ctrl)
			expectCoordinatedHost(host)
			host.EXPECT().Peerstore().Return(pstore)

			cfg := protectortest.NewTestNetworkConfig(t, protectorpb.NetworkState_BOOTSTRAP)
			cfg.AddPeer(context.Background(), coordinatorID, coordinatorAddrs)

			prov := mockstream.NewMockProvider(ctrl)

			tt.configure(t, ctrl, host, prov)

			mode := &protector.NetworkMode{CoordinatorID: coordinatorID}
			handler := protocol.NewCoordinatedHandler(host, prov, mode, cfg, nil)
			err := handler.Handshake(context.Background())

			if tt.err != nil {
				assert.EqualError(t, err, tt.err.Error())
			} else {
				require.NoError(t, err)
				tt.validate(t, cfg)
			}
		})
	}
}

func TestCoordinated_AddNode(t *testing.T) {
	coordinatorID := test.GeneratePeerID(t)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	host := mocks.NewMockHost(ctrl)
	expectCoordinatedHost(host)

	p := mockstream.NewMockProvider(ctrl)

	mode := &protector.NetworkMode{CoordinatorID: coordinatorID}
	handler := protocol.NewCoordinatedHandler(host, p, mode, nil, nil)

	peerID := test.GeneratePeerID(t)
	peerAddr := test.GeneratePeerMultiaddr(t, peerID)

	codec := mockstream.NewMockCodec(ctrl)
	codec.EXPECT().Encode(&bootstrappb.NodeIdentity{
		PeerId:        []byte(peerID),
		PeerAddr:      peerAddr.Bytes(),
		IdentityProof: []byte("b4tm4n"),
	})

	stream := mockstream.NewMockStream(ctrl)
	stream.EXPECT().Codec().Return(codec)
	stream.EXPECT().Close()

	streamtest.ExpectStreamPeerAndProtocol(t, p, coordinatorID, protocol.PrivateCoordinatorProposePID, stream, nil)

	err := handler.AddNode(context.Background(), peerID, peerAddr, []byte("b4tm4n"))
	require.NoError(t, err)
}

func TestCoordinated_RemoveNode(t *testing.T) {
	coordinatorID := test.GeneratePeerID(t)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	host := mocks.NewMockHost(ctrl)
	expectCoordinatedHost(host)

	p := mockstream.NewMockProvider(ctrl)

	mode := &protector.NetworkMode{CoordinatorID: coordinatorID}
	handler := protocol.NewCoordinatedHandler(host, p, mode, nil, nil)

	peerID := test.GeneratePeerID(t)

	codec := mockstream.NewMockCodec(ctrl)
	codec.EXPECT().Encode(&bootstrappb.NodeIdentity{PeerId: []byte(peerID)})

	stream := mockstream.NewMockStream(ctrl)
	stream.EXPECT().Codec().Return(codec)
	stream.EXPECT().Close()

	streamtest.ExpectStreamPeerAndProtocol(t, p, coordinatorID, protocol.PrivateCoordinatorProposePID, stream, nil)

	err := handler.RemoveNode(context.Background(), peerID)
	require.NoError(t, err)
}

func TestCoordinated_Accept(t *testing.T) {
	hostKey := test.GeneratePrivateKey(t)
	hostID := test.GetPeerIDFromKey(t, hostKey)

	coordinatorID := test.GeneratePeerID(t)
	mode := &protector.NetworkMode{
		CoordinatorID: coordinatorID,
	}

	peer1 := test.GeneratePeerID(t)

	testCases := []struct {
		name      string
		acceptID  peer.ID
		configure func(*testing.T, *gomock.Controller, *mocks.MockHost, *mockstream.MockProvider, proposal.Store)
		validate  func(*testing.T, proposal.Store)
		err       error
	}{{
		"missing-request",
		peer1,
		func(t *testing.T, _ *gomock.Controller, _ *mocks.MockHost, _ *mockstream.MockProvider, _ proposal.Store) {
		},
		func(t *testing.T, _ proposal.Store) {},
		proposal.ErrMissingRequest,
	}, {
		"add-node-request",
		peer1,
		func(t *testing.T, _ *gomock.Controller, _ *mocks.MockHost, _ *mockstream.MockProvider, propStore proposal.Store) {
			r := proposaltest.NewAddRequest(t, peer1)
			err := propStore.AddRequest(context.Background(), r)
			require.NoError(t, err)
		},
		func(t *testing.T, _ proposal.Store) {},
		protocol.ErrInvalidOperation,
	}, {
		"vote-remove-node",
		peer1,
		func(t *testing.T, _ *gomock.Controller, _ *mocks.MockHost, p *mockstream.MockProvider, propStore proposal.Store) {
			r := proposaltest.NewRemoveRequest(t, peer1)
			err := propStore.AddRequest(context.Background(), r)
			require.NoError(t, err)

			streamtest.ExpectStreamPeerAndProtocol(t, p, coordinatorID, protocol.PrivateCoordinatorVotePID, nil, errors.New("no stream"))
		},
		func(t *testing.T, propStore proposal.Store) {
			r, err := propStore.Get(context.Background(), peer1)
			require.NoError(t, err)
			assert.NotNil(t, r)
		},
		errors.New("no stream"),
	}, {
		"remove-request-from-store",
		peer1,
		func(t *testing.T, ctrl *gomock.Controller, _ *mocks.MockHost, p *mockstream.MockProvider, propStore proposal.Store) {
			r := proposaltest.NewRemoveRequest(t, peer1)
			err := propStore.AddRequest(context.Background(), r)
			require.NoError(t, err)

			codec := mockstream.NewMockCodec(ctrl)
			streamtest.ExpectEncodeVote(t, codec, r)

			stream := mockstream.NewMockStream(ctrl)
			stream.EXPECT().Codec().Return(codec)
			stream.EXPECT().Close()

			streamtest.ExpectStreamPeerAndProtocol(t, p, coordinatorID, protocol.PrivateCoordinatorVotePID, stream, nil)
		},
		func(t *testing.T, propStore proposal.Store) {
			r, _ := propStore.Get(context.Background(), peer1)
			assert.Nil(t, r)
		},
		nil,
	}}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			pstore := pstoremem.NewPeerstore()
			require.NoError(t, pstore.AddPrivKey(hostID, hostKey))

			host := mocks.NewMockHost(ctrl)
			expectCoordinatedHost(host)
			host.EXPECT().Peerstore().Return(pstore).AnyTimes()
			host.EXPECT().ID().Return(hostID).AnyTimes()

			prov := mockstream.NewMockProvider(ctrl)
			propStore := proposal.NewInMemoryStore()

			tt.configure(t, ctrl, host, prov, propStore)

			handler := protocol.NewCoordinatedHandler(host, prov, mode, nil, propStore)
			err := handler.Accept(context.Background(), tt.acceptID)

			if tt.err != nil {
				assert.EqualError(t, err, tt.err.Error())
			} else {
				require.NoError(t, err)
				tt.validate(t, propStore)
			}
		})
	}
}

func TestCoordinated_Reject(t *testing.T) {
	coordinatorID := test.GeneratePeerID(t)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	host := mocks.NewMockHost(ctrl)
	expectCoordinatedHost(host)

	propStore := mockproposal.NewMockStore(ctrl)

	mode := &protector.NetworkMode{CoordinatorID: coordinatorID}
	handler := protocol.NewCoordinatedHandler(host, nil, mode, nil, propStore)

	peerID := test.GeneratePeerID(t)
	propStore.EXPECT().Remove(gomock.Any(), peerID)

	err := handler.Reject(context.Background(), peerID)
	require.NoError(t, err)
}
