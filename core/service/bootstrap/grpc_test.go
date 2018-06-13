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

package bootstrap

import (
	"context"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stratumn/alice/core/protector"
	protocol "github.com/stratumn/alice/core/protocol/bootstrap"
	"github.com/stratumn/alice/core/protocol/bootstrap/proposal"
	"github.com/stratumn/alice/core/protocol/bootstrap/proposal/mocks"
	"github.com/stratumn/alice/core/service/bootstrap/mockbootstrap"
	pb "github.com/stratumn/alice/pb/bootstrap"
	protectorpb "github.com/stratumn/alice/pb/protector"
	"github.com/stratumn/alice/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testPublicNetworkServer() *grpcServer {
	return &grpcServer{
		GetNetworkMode: func() *protector.NetworkMode {
			return nil
		},
		GetProtocolHandler: func() protocol.Handler {
			return nil
		},
		GetProposalStore: func() proposal.Store {
			return nil
		},
	}
}

func testPrivateNetworkServer(mode *protector.NetworkMode, handler protocol.Handler, store proposal.Store) *grpcServer {
	return &grpcServer{
		GetNetworkMode: func() *protector.NetworkMode {
			return mode
		},
		GetProtocolHandler: func() protocol.Handler {
			return handler
		},
		GetProposalStore: func() proposal.Store {
			return store
		},
	}
}

func TestGRPCServer_AddNode(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	t.Run("public-network", func(t *testing.T) {
		s := testPublicNetworkServer()

		_, err := s.AddNode(ctx, nil)
		assert.EqualError(t, err, ErrNotAllowed.Error())
	})

	t.Run("private-network", func(t *testing.T) {
		t.Run("invalid-peer-id", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			handler := mockbootstrap.NewMockHandler(ctrl)
			s := testPrivateNetworkServer(protector.NewCoordinatorNetworkMode(), handler, nil)

			nodeID := &pb.NodeIdentity{PeerId: []byte("b4tm4n")}
			_, err := s.AddNode(ctx, nodeID)
			require.EqualError(t, err, protectorpb.ErrInvalidPeerID.Error())
		})

		t.Run("invalid-peer-addr", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			handler := mockbootstrap.NewMockHandler(ctrl)
			s := testPrivateNetworkServer(protector.NewCoordinatorNetworkMode(), handler, nil)

			nodeID := &pb.NodeIdentity{
				PeerId:   []byte(test.GeneratePeerID(t)),
				PeerAddr: []byte("not/a/multiaddr"),
			}
			_, err := s.AddNode(ctx, nodeID)
			require.EqualError(t, err, protectorpb.ErrInvalidPeerAddr.Error())
		})

		t.Run("valid-request", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			handler := mockbootstrap.NewMockHandler(ctrl)
			s := testPrivateNetworkServer(protector.NewCoordinatorNetworkMode(), handler, nil)

			peerID := test.GeneratePeerID(t)
			peerAddr := test.GeneratePeerMultiaddr(t, peerID)
			nodeID := &pb.NodeIdentity{
				PeerId:        []byte(peerID),
				PeerAddr:      peerAddr.Bytes(),
				IdentityProof: []byte("I'm the batman"),
			}

			handler.EXPECT().AddNode(gomock.Any(), peerID, peerAddr, nodeID.IdentityProof).Times(1)

			_, err := s.AddNode(ctx, nodeID)
			require.NoError(t, err)
		})
	})
}

func TestGRPCServer_RemoveNode(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	t.Run("public-network", func(t *testing.T) {
		s := testPublicNetworkServer()

		_, err := s.RemoveNode(ctx, nil)
		assert.EqualError(t, err, ErrNotAllowed.Error())
	})

	t.Run("private-network", func(t *testing.T) {
		t.Run("invalid-peer-id", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			handler := mockbootstrap.NewMockHandler(ctrl)
			s := testPrivateNetworkServer(protector.NewCoordinatorNetworkMode(), handler, nil)

			nodeID := &pb.NodeIdentity{PeerId: []byte("b4tm4n")}
			_, err := s.RemoveNode(ctx, nodeID)
			require.EqualError(t, err, protectorpb.ErrInvalidPeerID.Error())
		})

		t.Run("valid-request", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			handler := mockbootstrap.NewMockHandler(ctrl)
			s := testPrivateNetworkServer(protector.NewCoordinatorNetworkMode(), handler, nil)

			peerID := test.GeneratePeerID(t)
			nodeID := &pb.NodeIdentity{PeerId: []byte(peerID)}

			handler.EXPECT().RemoveNode(gomock.Any(), peerID).Times(1)

			_, err := s.RemoveNode(ctx, nodeID)
			require.NoError(t, err)
		})
	})
}

func TestGRPCServer_Accept(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	t.Run("public-network", func(t *testing.T) {
		s := testPublicNetworkServer()

		_, err := s.Accept(ctx, nil)
		assert.EqualError(t, err, ErrNotAllowed.Error())
	})

	t.Run("private-network", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		handler := mockbootstrap.NewMockHandler(ctrl)
		s := testPrivateNetworkServer(protector.NewCoordinatorNetworkMode(), handler, nil)

		peerID := test.GeneratePeerID(t)
		message := &pb.PeerID{PeerId: []byte(peerID)}

		handler.EXPECT().Accept(gomock.Any(), peerID).Times(1)

		_, err := s.Accept(ctx, message)
		require.NoError(t, err)
	})
}

func TestGRPCServer_Reject(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	t.Run("public-network", func(t *testing.T) {
		s := testPublicNetworkServer()

		_, err := s.Reject(ctx, nil)
		assert.EqualError(t, err, ErrNotAllowed.Error())
	})

	t.Run("private-network", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		handler := mockbootstrap.NewMockHandler(ctrl)
		s := testPrivateNetworkServer(protector.NewCoordinatorNetworkMode(), handler, nil)

		peerID := test.GeneratePeerID(t)
		message := &pb.PeerID{PeerId: []byte(peerID)}

		handler.EXPECT().Reject(gomock.Any(), peerID).Times(1)

		_, err := s.Reject(ctx, message)
		require.NoError(t, err)
	})
}

func TestGRPCServer_List(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	t.Run("public-network", func(t *testing.T) {
		s := testPublicNetworkServer()

		_, err := s.Reject(ctx, nil)
		assert.EqualError(t, err, ErrNotAllowed.Error())
	})

	t.Run("private-network", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		store := mockproposal.NewMockStore(ctrl)
		s := testPrivateNetworkServer(protector.NewCoordinatorNetworkMode(), nil, store)

		addReq := &proposal.Request{
			Type:     proposal.AddNode,
			PeerID:   test.GeneratePeerID(t),
			PeerAddr: test.GenerateMultiaddr(t),
			Info:     []byte("he's b4tm4n"),
		}

		removeReq := &proposal.Request{
			Type:   proposal.RemoveNode,
			PeerID: test.GeneratePeerID(t),
			Info:   []byte("he's not b4tm4n"),
		}

		store.EXPECT().List(gomock.Any()).Times(1).Return(
			[]*proposal.Request{addReq, removeReq},
			nil,
		)

		mockServer := mockbootstrap.NewMockBootstrap_ListServer(ctrl)
		mockServer.EXPECT().Context().Times(1).Return(ctx)
		mockServer.EXPECT().Send(gomock.Any()).Times(1).Do(func(req *pb.UpdateProposal) error {
			assert.Equal(t, pb.UpdateType_AddNode, req.UpdateType)
			assert.Equal(t, []byte(addReq.PeerID), req.NodeDetails.PeerId)
			assert.Equal(t, addReq.PeerAddr.Bytes(), req.NodeDetails.PeerAddr)
			assert.Equal(t, addReq.Info, req.NodeDetails.IdentityProof)
			return nil
		})
		mockServer.EXPECT().Send(gomock.Any()).Times(1).Do(func(req *pb.UpdateProposal) error {
			assert.Equal(t, pb.UpdateType_RemoveNode, req.UpdateType)
			assert.Equal(t, []byte(removeReq.PeerID), req.NodeDetails.PeerId)
			assert.Equal(t, removeReq.Info, req.NodeDetails.IdentityProof)
			return nil
		})

		err := s.List(nil, mockServer)
		require.NoError(t, err)
	})
}

func TestGRPCServer_Complete(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	t.Run("public-network", func(t *testing.T) {
		s := testPublicNetworkServer()

		_, err := s.Complete(ctx, nil)
		assert.EqualError(t, err, ErrNotAllowed.Error())
	})

	t.Run("private-network-coordinated", func(t *testing.T) {
		coordinatorID := test.GeneratePeerID(t)
		coordinatorAddr := test.GeneratePeerMultiaddr(t, coordinatorID)
		mode, _ := protector.NewCoordinatedNetworkMode(
			coordinatorID.Pretty(),
			[]string{coordinatorAddr.String()},
		)
		s := testPrivateNetworkServer(mode, nil, nil)

		_, err := s.Complete(ctx, nil)
		assert.EqualError(t, err, ErrNotAllowed.Error())
	})

	t.Run("private-network-coordinator", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		handler := mockbootstrap.NewMockHandler(ctrl)
		s := testPrivateNetworkServer(protector.NewCoordinatorNetworkMode(), handler, nil)

		handler.EXPECT().CompleteBootstrap(gomock.Any()).Times(1)

		_, err := s.Complete(ctx, nil)
		require.NoError(t, err)
	})
}
