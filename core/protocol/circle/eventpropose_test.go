package circle

import (
	"context"
	"testing"

	"github.com/stratumn/alice/core/protocol/circle/lib/mocklib"
	pb "github.com/stratumn/alice/pb/circle"

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
