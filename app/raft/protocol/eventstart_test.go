package protocol

import (
	"context"
	"testing"

	"github.com/stratumn/go-indigonode/app/raft/protocol/lib"
	"github.com/stratumn/go-indigonode/app/raft/protocol/lib/mocklib"

	"github.com/golang/mock/gomock"
)

func TestStartNormal(t *testing.T) {
	mockController := gomock.NewController(t)
	defer mockController.Finish()

	mockLib := mocklib.NewMockLib(mockController)
	mockStorage := mocklib.NewMockStorage(mockController)
	mockConfig := mocklib.NewMockConfig(mockController)
	mockPeer := mocklib.NewMockPeer(mockController)
	mockNode := mocklib.NewMockNode(mockController)

	mockPeers := []lib.Peer{mockPeer}

	mockLib.EXPECT().NewPeer(testNodeID, testLocalPeer).Return(mockPeer)
	mockLib.EXPECT().NewMemoryStorage().Return(mockStorage)
	mockLib.EXPECT().NewConfig(
		testNodeID,
		testElectionTick,
		testHeartbeatTick,
		mockStorage,
		testMaxSizePerMsg,
		testMaxInflightMsgs,
	).Return(mockConfig)
	mockLib.EXPECT().StartNode(mockConfig, mockPeers).Return(mockNode)
	mockNode.EXPECT().Ready()

	ctx := context.Background()

	c := circleProcess{
		lib:             mockLib,
		localPeer:       testLocalPeer,
		electionTick:    testElectionTick,
		heartbeatTick:   testHeartbeatTick,
		maxSizePerMsg:   testMaxSizePerMsg,
		maxInflightMsgs: testMaxInflightMsgs,
		tickerInterval:  testTickerInterval,
	}

	c.eventStart(ctx)

}

func TestStartAlreadyStarted(t *testing.T) {
	mockController := gomock.NewController(t)
	defer mockController.Finish()

	mockLib := mocklib.NewMockLib(mockController)
	mockNode := mocklib.NewMockNode(mockController)

	ctx := context.Background()

	c := circleProcess{
		lib:  mockLib,
		node: mockNode,
	}

	c.eventStart(ctx)

}
