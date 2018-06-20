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
	"time"

	"github.com/golang/mock/gomock"
	"github.com/pkg/errors"
	"github.com/stratumn/alice/core/protector"
	"github.com/stratumn/alice/core/protector/protectortest"
	"github.com/stratumn/alice/core/protocol/bootstrap"
	"github.com/stratumn/alice/core/protocol/bootstrap/bootstraptest"
	"github.com/stratumn/alice/core/protocol/bootstrap/proposal"
	"github.com/stratumn/alice/core/protocol/bootstrap/proposal/mocks"
	"github.com/stratumn/alice/core/protocol/bootstrap/proposaltest"
	"github.com/stratumn/alice/core/streamutil"
	"github.com/stratumn/alice/core/streamutil/mockstream"
	"github.com/stratumn/alice/core/streamutil/streamtest"
	bootstrappb "github.com/stratumn/alice/pb/bootstrap"
	protectorpb "github.com/stratumn/alice/pb/protector"
	"github.com/stratumn/alice/test"
	"github.com/stratumn/alice/test/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	inet "gx/ipfs/QmXoz9o2PT3tEzf7hicegwex5UgVP54n3k82K7jrWFyN86/go-libp2p-net"
	"gx/ipfs/QmcJukH2sAFjY3HdBKq35WDzWoL3UUu2gt9wdfqZTUyM74/go-libp2p-peer"
	"gx/ipfs/QmdeiKhUy1TVGBaKxt7y1QmBDLBdisSrLJ1x58Eoj4PXUh/go-libp2p-peerstore"
	ihost "gx/ipfs/QmfZTdmunzKzAGJrSvXXQbQ5kLLUiEMX5vdwux7iXkdk7D/go-libp2p-host"
)

// TODO: remove/move to helper packages

func newNetworkConfig(t *testing.T) protector.NetworkConfig {
	config, err := protector.NewInMemoryConfig(
		context.Background(),
		protectorpb.NewNetworkConfig(protectorpb.NetworkState_BOOTSTRAP),
	)
	require.NoError(t, err, "protector.NewInMemoryConfig()")
	return config
}

func waitUntilAllowed(t *testing.T, peerID peer.ID, networkConfig protector.NetworkConfig) {
	test.WaitUntil(t, 100*time.Millisecond, 20*time.Millisecond,
		func() error {
			if !networkConfig.IsAllowed(context.Background(), peerID) {
				return errors.New("peer not allowed")
			}

			return nil
		}, "peer not allowed in time")
}

func waitUntilNotAllowed(t *testing.T, peerID peer.ID, networkConfig protector.NetworkConfig) {
	test.WaitUntil(t, 100*time.Millisecond, 20*time.Millisecond,
		func() error {
			if networkConfig.IsAllowed(context.Background(), peerID) {
				return errors.New("still allowed")
			}

			return nil
		}, "peer not removed in time")
}

func waitUntilProposed(t *testing.T, s proposal.Store, peerID peer.ID) {
	test.WaitUntil(t, 100*time.Millisecond, 10*time.Millisecond,
		func() error {
			r, _ := s.Get(context.Background(), peerID)
			if r == nil {
				return errors.New("proposal not received yet")
			}

			return nil
		}, "proposal not received in time")
}

func waitUntilDisconnected(t *testing.T, host ihost.Host, peerID peer.ID) {
	test.WaitUntil(t, 100*time.Millisecond, 10*time.Millisecond,
		func() error {
			if host.Network().Connectedness(peerID) == inet.Connected {
				return errors.New("peers still connected")
			}

			return nil
		}, "peers not disconnected in time")
}

// END TODO

func expectCoordinatedHost(host *mocks.MockHost) {
	host.EXPECT().SetStreamHandler(bootstrap.PrivateCoordinatedConfigPID, gomock.Any())
	host.EXPECT().SetStreamHandler(bootstrap.PrivateCoordinatedProposePID, gomock.Any())
}

func TestCoordinated_Close(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	host := mocks.NewMockHost(ctrl)
	expectCoordinatedHost(host)

	mode := &protector.NetworkMode{
		CoordinatorID: test.GeneratePeerID(t),
	}
	handler := bootstrap.NewCoordinatedHandler(host, nil, mode, nil, nil)
	require.NotNil(t, handler)

	host.EXPECT().RemoveStreamHandler(bootstrap.PrivateCoordinatedConfigPID)
	host.EXPECT().RemoveStreamHandler(bootstrap.PrivateCoordinatedProposePID)

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
	h func(*bootstrap.CoordinatedHandler) streamutil.AutoCloseHandler,
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
		handler := bootstrap.NewCoordinatedHandler(host, nil, mode, cfg, s).(*bootstrap.CoordinatedHandler)
		err := h(handler)(ctx, bootstraptest.NewEvent(), stream, codec)

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
		tt.Run(t, func(handler *bootstrap.CoordinatedHandler) streamutil.AutoCloseHandler {
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
		tt.Run(t, func(handler *bootstrap.CoordinatedHandler) streamutil.AutoCloseHandler {
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
				bootstrap.PrivateCoordinatorHandshakePID,
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
				bootstrap.PrivateCoordinatorProposePID,
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

			pstore := peerstore.NewPeerstore()
			pstore.AddAddrs(coordinatorID, coordinatorAddrs, peerstore.PermanentAddrTTL)

			host := mocks.NewMockHost(ctrl)
			expectCoordinatedHost(host)
			host.EXPECT().Peerstore().Return(pstore)

			cfg := protectortest.NewTestNetworkConfig(t, protectorpb.NetworkState_BOOTSTRAP)
			cfg.AddPeer(context.Background(), coordinatorID, coordinatorAddrs)

			prov := mockstream.NewMockProvider(ctrl)

			tt.configure(t, ctrl, host, prov)

			mode := &protector.NetworkMode{CoordinatorID: coordinatorID}
			handler := bootstrap.NewCoordinatedHandler(host, prov, mode, cfg, nil)
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
	handler := bootstrap.NewCoordinatedHandler(host, p, mode, nil, nil)

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

	streamtest.ExpectStreamPeerAndProtocol(t, p, coordinatorID, bootstrap.PrivateCoordinatorProposePID, stream, nil)

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
	handler := bootstrap.NewCoordinatedHandler(host, p, mode, nil, nil)

	peerID := test.GeneratePeerID(t)

	codec := mockstream.NewMockCodec(ctrl)
	codec.EXPECT().Encode(&bootstrappb.NodeIdentity{PeerId: []byte(peerID)})

	stream := mockstream.NewMockStream(ctrl)
	stream.EXPECT().Codec().Return(codec)
	stream.EXPECT().Close()

	streamtest.ExpectStreamPeerAndProtocol(t, p, coordinatorID, bootstrap.PrivateCoordinatorProposePID, stream, nil)

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
		bootstrap.ErrInvalidOperation,
	}, {
		"vote-remove-node",
		peer1,
		func(t *testing.T, _ *gomock.Controller, _ *mocks.MockHost, p *mockstream.MockProvider, propStore proposal.Store) {
			r := proposaltest.NewRemoveRequest(t, peer1)
			err := propStore.AddRequest(context.Background(), r)
			require.NoError(t, err)

			streamtest.ExpectStreamPeerAndProtocol(t, p, coordinatorID, bootstrap.PrivateCoordinatorVotePID, nil, errors.New("no stream"))
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

			streamtest.ExpectStreamPeerAndProtocol(t, p, coordinatorID, bootstrap.PrivateCoordinatorVotePID, stream, nil)
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

			pstore := peerstore.NewPeerstore()
			require.NoError(t, pstore.AddPrivKey(hostID, hostKey))

			host := mocks.NewMockHost(ctrl)
			expectCoordinatedHost(host)
			host.EXPECT().Peerstore().Return(pstore).AnyTimes()
			host.EXPECT().ID().Return(hostID).AnyTimes()

			prov := mockstream.NewMockProvider(ctrl)
			propStore := proposal.NewInMemoryStore()

			tt.configure(t, ctrl, host, prov, propStore)

			handler := bootstrap.NewCoordinatedHandler(host, prov, mode, nil, propStore)
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
	handler := bootstrap.NewCoordinatedHandler(host, nil, mode, nil, propStore)

	peerID := test.GeneratePeerID(t)
	propStore.EXPECT().Remove(gomock.Any(), peerID)

	err := handler.Reject(context.Background(), peerID)
	require.NoError(t, err)
}
