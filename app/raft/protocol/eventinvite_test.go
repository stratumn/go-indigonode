package protocol

import (
	"context"
	"testing"

	"github.com/stratumn/alice/app/raft/pb"
	"github.com/stratumn/alice/app/raft/protocol/lib"
	"github.com/stratumn/alice/app/raft/protocol/lib/mocklib"

	"github.com/golang/mock/gomock"
)

func TestInviteNormal(t *testing.T) {
	mockController := gomock.NewController(t)
	defer mockController.Finish()

	ctx := context.Background()

	mockLib := mocklib.NewMockLib(mockController)
	mockNode := mocklib.NewMockNode(mockController)
	mockLocalPeer := mocklib.NewMockPeer(mockController)
	mockConfChange := mocklib.NewMockConfChange(mockController)

	mockLocalPeer.EXPECT().Context().AnyTimes().Return(testLocalPeer)

	mockLib.EXPECT().NewConfChange(
		testConfigChangeID,
		lib.ConfChangeAddNode,
		testLastNodeID+1,
		testRemotePeer,
	).Return(mockConfChange)

	mockNode.EXPECT().ProposeConfChange(
		ctx,
		mockConfChange,
	)

	c := circleProcess{
		lib:            mockLib,
		node:           mockNode,
		configChangeID: testConfigChangeID,
		lastNodeID:     testLastNodeID,
		peers:          []lib.Peer{mockLocalPeer},
	}

	msg := hubCallInvite{PeerID: pb.PeerID{Address: testRemotePeer}}

	c.eventInvite(ctx, msg)

}

func TestInviteAlreadyInvited(t *testing.T) {
	mockController := gomock.NewController(t)
	defer mockController.Finish()

	ctx := context.Background()

	mockLib := mocklib.NewMockLib(mockController)
	mockNode := mocklib.NewMockNode(mockController)
	mockLocalPeer := mocklib.NewMockPeer(mockController)
	mockRemotePeer := mocklib.NewMockPeer(mockController)

	mockLocalPeer.EXPECT().Context().AnyTimes().Return(testLocalPeer)
	mockRemotePeer.EXPECT().Context().AnyTimes().Return(testRemotePeer)

	c := circleProcess{
		lib:            mockLib,
		node:           mockNode,
		configChangeID: testConfigChangeID,
		peers:          []lib.Peer{mockLocalPeer, mockRemotePeer},
	}

	msg := hubCallInvite{PeerID: pb.PeerID{Address: testRemotePeer}}

	c.eventInvite(ctx, msg)

}

func TestInviteStopped(t *testing.T) {
	mockController := gomock.NewController(t)
	defer mockController.Finish()

	ctx := context.Background()

	c := circleProcess{}

	msg := hubCallInvite{PeerID: pb.PeerID{Address: testRemotePeer}}

	c.eventInvite(ctx, msg)

}
