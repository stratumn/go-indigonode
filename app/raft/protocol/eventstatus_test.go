package protocol

import (
	"context"
	"testing"
	"time"

	"github.com/stratumn/go-indigonode/app/raft/pb"
	"github.com/stratumn/go-indigonode/app/raft/protocol/lib/mocklib"

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

	select {
	case <-doneChan:
	case <-time.NewTimer(1 * time.Second).C:
		t.Error("timeout")
	}

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

	select {
	case <-doneChan:
	case <-time.NewTimer(1 * time.Second).C:
		t.Error("timeout")
	}
}
