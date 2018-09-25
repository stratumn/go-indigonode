package protocol

import (
	"context"
	"testing"

	"github.com/stratumn/go-node/app/raft/pb"
	"github.com/stratumn/go-node/app/raft/protocol/lib/mocklib"

	"github.com/golang/mock/gomock"
)

func TestStepNormal(t *testing.T) {
	mockController := gomock.NewController(t)
	defer mockController.Finish()

	ctx := context.Background()

	mockNode := mocklib.NewMockNode(mockController)

	mockNode.EXPECT().Step(ctx, testInternodeMessage)

	c := circleProcess{
		node: mockNode,
	}

	msg := pb.Internode{PeerId: &pb.PeerID{Address: testRemotePeer}, Message: testInternodeMessage}

	c.eventStep(ctx, msg)

}

func TestStepStopped(t *testing.T) {
	mockController := gomock.NewController(t)
	defer mockController.Finish()

	ctx := context.Background()

	c := circleProcess{}

	msg := pb.Internode{PeerId: &pb.PeerID{Address: testRemotePeer}, Message: testInternodeMessage}
	c.eventStep(ctx, msg)
}
