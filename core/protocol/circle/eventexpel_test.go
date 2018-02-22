package circle

import (
	"context"
	"testing"

	"github.com/stratumn/alice/core/protocol/circle/lib"
	"github.com/stratumn/alice/core/protocol/circle/lib/mocklib"
	pb "github.com/stratumn/alice/pb/circle"

	"github.com/golang/mock/gomock"
)

func TestExpelNormal(t *testing.T) {
	mockController := gomock.NewController(t)
	defer mockController.Finish()

	ctx := context.Background()

	mockLib := mocklib.NewMockLib(mockController)
	mockNode := mocklib.NewMockNode(mockController)
	mockLocalPeer := mocklib.NewMockPeer(mockController)
	mockRemotePeer := mocklib.NewMockPeer(mockController)
	mockConfChange := mocklib.NewMockConfChange(mockController)

	mockLocalPeer.EXPECT().Context().AnyTimes().Return(testLocalPeer)
	mockLocalPeer.EXPECT().ID().AnyTimes().Return(testNodeID)
	mockRemotePeer.EXPECT().Context().AnyTimes().Return(testRemotePeer)
	mockRemotePeer.EXPECT().ID().AnyTimes().Return(testRemoteNodeID)

	mockLib.EXPECT().NewConfChange(
		testConfigChangeID,
		lib.ConfChangeRemoveNode,
		testRemoteNodeID,
		nil,
	).Return(mockConfChange)

	mockNode.EXPECT().ProposeConfChange(
		ctx,
		mockConfChange,
	)

	c := circleProcess{
		lib:            mockLib,
		node:           mockNode,
		configChangeID: testConfigChangeID,
		peers:          []lib.Peer{mockLocalPeer, mockRemotePeer},
	}

	msg := hubCallExpel{PeerID: pb.PeerID{Address: testRemotePeer}}

	c.eventExpel(ctx, msg)

}

func TestExpelAlreadyNotFound(t *testing.T) {
	mockController := gomock.NewController(t)
	defer mockController.Finish()

	ctx := context.Background()

	mockLib := mocklib.NewMockLib(mockController)
	mockNode := mocklib.NewMockNode(mockController)
	mockLocalPeer := mocklib.NewMockPeer(mockController)

	mockLocalPeer.EXPECT().Context().AnyTimes().Return(testLocalPeer)
	mockLocalPeer.EXPECT().ID().AnyTimes().Return(testNodeID)

	c := circleProcess{
		lib:            mockLib,
		node:           mockNode,
		configChangeID: testConfigChangeID,
		peers:          []lib.Peer{mockLocalPeer},
	}

	msg := hubCallExpel{PeerID: pb.PeerID{Address: testRemotePeer}}

	c.eventExpel(ctx, msg)

}

func TestExpelStopped(t *testing.T) {
	mockController := gomock.NewController(t)
	defer mockController.Finish()

	ctx := context.Background()

	c := circleProcess{}

	msg := hubCallExpel{PeerID: pb.PeerID{Address: testRemotePeer}}

	c.eventExpel(ctx, msg)

}
