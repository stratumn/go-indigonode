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
	"github.com/stratumn/alice/core/service/bootstrap/mockbootstrap"
	pb "github.com/stratumn/alice/grpc/bootstrap"
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
	}
}

func testPrivateNetworkServer(handler *mockbootstrap.MockHandler) *grpcServer {
	return &grpcServer{
		GetNetworkMode: func() *protector.NetworkMode {
			return protector.NewCoordinatorNetworkMode()
		},
		GetProtocolHandler: func() protocol.Handler {
			return handler
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
			s := testPrivateNetworkServer(handler)

			nodeID := &pb.NodeIdentity{PeerId: []byte("b4tm4n")}
			_, err := s.AddNode(ctx, nodeID)
			require.EqualError(t, err, protectorpb.ErrInvalidPeerID.Error())
		})

		t.Run("invalid-peer-addr", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			handler := mockbootstrap.NewMockHandler(ctrl)
			s := testPrivateNetworkServer(handler)

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
			s := testPrivateNetworkServer(handler)

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
		s := testPrivateNetworkServer(handler)

		peerID := test.GeneratePeerID(t)
		message := &pb.PeerID{PeerId: []byte(peerID)}

		handler.EXPECT().Accept(gomock.Any(), peerID).Times(1)

		_, err := s.Accept(ctx, message)
		require.NoError(t, err)
	})
}
