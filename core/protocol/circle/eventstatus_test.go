package circle

import (
	"context"
	"testing"

	"github.com/stratumn/alice/core/protocol/circle/lib/mocklib"
	pb "github.com/stratumn/alice/pb/circle"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestStatusRunning(t *testing.T) {
	mockController := gomock.NewController(t)
	defer mockController.Finish()

	mockNode := mocklib.NewMockNode(mockController)

	ctx := context.Background()

	c := circleProcess{
		node:   mockNode,
		nodeID: testNodeID,
	}

	statusInfoChan := make(chan pb.StatusInfo)
	doneChan := make(chan struct{})

	go func() {
		statusInfo := <-statusInfoChan
		assert.Equal(t, statusInfo, pb.StatusInfo{Running: true, Id: testNodeID})
		doneChan <- struct{}{}
	}()

	msg := hubCallStatus{StatusInfoChan: statusInfoChan}

	c.eventStatus(ctx, msg)

	<-doneChan

}

func TestStatusStopped(t *testing.T) {

	ctx := context.Background()

	c := circleProcess{}

	statusInfoChan := make(chan pb.StatusInfo)
	doneChan := make(chan struct{})

	go func() {
		statusInfo := <-statusInfoChan
		assert.Equal(t, statusInfo, pb.StatusInfo{Running: false, Id: 0})
		doneChan <- struct{}{}
	}()

	msg := hubCallStatus{StatusInfoChan: statusInfoChan}

	c.eventStatus(ctx, msg)

	<-doneChan
}
