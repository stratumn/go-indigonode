package protocol

import (
	"context"
	"testing"

	"github.com/stratumn/go-indigonode/app/raft/protocol/lib/mocklib"

	"github.com/golang/mock/gomock"
)

func TestStopNormal(t *testing.T) {
	mockController := gomock.NewController(t)
	defer mockController.Finish()

	mockNode := mocklib.NewMockNode(mockController)

	mockNode.EXPECT().Stop()

	ctx := context.Background()

	c := circleProcess{
		node: mockNode,
	}

	c.eventStop(ctx)

}

func TestStopAlreadyStopped(t *testing.T) {
	ctx := context.Background()
	c := circleProcess{}
	c.eventStop(ctx)
}

// TODO: testStopIdentical
