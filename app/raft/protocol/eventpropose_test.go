package protocol

import (
	"context"
	"testing"

	"github.com/stratumn/go-indigonode/app/raft/pb"
	"github.com/stratumn/go-indigonode/app/raft/protocol/lib/mocklib"

	"github.com/golang/mock/gomock"
)

func TestProposeNormal(t *testing.T) {
	mockController := gomock.NewController(t)
	defer mockController.Finish()

	ctx := context.Background()

	proposal := []byte("hello")

	mockNode := mocklib.NewMockNode(mockController)

	mockNode.EXPECT().Propose(ctx, proposal)

	c := circleProcess{
		node: mockNode,
	}

	msg := hubCallPropose{Proposal: pb.Proposal{Data: proposal}}

	c.eventPropose(ctx, msg)

}

func TestProposeStopped(t *testing.T) {
	mockController := gomock.NewController(t)
	defer mockController.Finish()

	ctx := context.Background()

	proposal := []byte("hello")

	c := circleProcess{}

	msg := hubCallPropose{Proposal: pb.Proposal{Data: proposal}}

	c.eventPropose(ctx, msg)
}
