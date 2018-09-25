package protocol

import (
	"context"
	"testing"

	"github.com/stratumn/go-node/app/raft/protocol/lib/mocklib"

	"github.com/golang/mock/gomock"
)

func TestTickNormal(t *testing.T) {
	mockController := gomock.NewController(t)
	defer mockController.Finish()

	ctx := context.Background()

	mockNode := mocklib.NewMockNode(mockController)

	mockNode.EXPECT().Tick()

	c := circleProcess{
		node: mockNode,
	}

	c.eventTick(ctx)

}

func TestTickStopped(t *testing.T) {
	mockController := gomock.NewController(t)
	defer mockController.Finish()

	ctx := context.Background()

	c := circleProcess{}

	c.eventTick(ctx)
}
