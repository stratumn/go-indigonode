package protocol

import (
	"context"
	"testing"
	"time"

	"github.com/stratumn/alice/app/raft/pb"
	"github.com/stratumn/alice/app/raft/protocol/lib"
	"github.com/stratumn/alice/app/raft/protocol/lib/mocklib"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestJoinNormal(t *testing.T) {
	mockController := gomock.NewController(t)
	defer mockController.Finish()

	mockLib := mocklib.NewMockLib(mockController)
	mockStorage := mocklib.NewMockStorage(mockController)
	mockConfig := mocklib.NewMockConfig(mockController)
	mockLocalPeer := mocklib.NewMockPeer(mockController)
	mockRemotePeer := mocklib.NewMockPeer(mockController)
	mockNode := mocklib.NewMockNode(mockController)

	mockPeers := []lib.Peer{mockLocalPeer, mockRemotePeer}

	mockLib.EXPECT().NewPeer(testNodeID, testLocalPeer).Return(mockLocalPeer)
	mockLib.EXPECT().NewPeer(testRemoteNodeID, testRemotePeer).Return(mockRemotePeer)
	mockLib.EXPECT().NewMemoryStorage().Return(mockStorage)
	mockLib.EXPECT().NewConfig(
		testNodeID,
		testElectionTick,
		testHeartbeatTick,
		mockStorage,
		testMaxSizePerMsg,
		testMaxInflightMsgs,
	).Return(mockConfig)
	mockLib.EXPECT().StartNode(mockConfig, nil).Return(mockNode)
	mockNode.EXPECT().Ready()

	ctx := context.Background()

	callDiscoverChan := make(chan hubCallDiscover)
	doneChan := make(chan struct{})

	go func() {
		callDiscover := <-callDiscoverChan
		assert.Equal(t, callDiscover.PeerID.Address, testRemotePeer)
		callDiscover.PeersChan <- pb.Peer{Id: testNodeID, Address: testLocalPeer}
		callDiscover.PeersChan <- pb.Peer{Id: testRemoteNodeID, Address: testRemotePeer}
		close(callDiscover.PeersChan)
		doneChan <- struct{}{}
	}()

	c := circleProcess{
		lib:              mockLib,
		localPeer:        testLocalPeer,
		electionTick:     testElectionTick,
		heartbeatTick:    testHeartbeatTick,
		maxSizePerMsg:    testMaxSizePerMsg,
		maxInflightMsgs:  testMaxInflightMsgs,
		tickerInterval:   testTickerInterval,
		callDiscoverChan: callDiscoverChan,
	}

	msg := hubCallJoin{PeerID: pb.PeerID{Address: testRemotePeer}}

	c.eventJoin(ctx, msg)

	assert.Equal(t, c.peers, mockPeers)

	select {
	case <-doneChan:
	case <-time.NewTimer(1 * time.Second).C:
		t.Error("timeout")
	}

}

func TestJoinAlreadyStarted(t *testing.T) {
	mockController := gomock.NewController(t)
	defer mockController.Finish()

	mockLib := mocklib.NewMockLib(mockController)
	mockNode := mocklib.NewMockNode(mockController)

	ctx := context.Background()

	c := circleProcess{
		lib:  mockLib,
		node: mockNode,
	}

	msg := hubCallJoin{PeerID: pb.PeerID{Address: testRemotePeer}}

	c.eventJoin(ctx, msg)

}

func TestJoinNotInvited(t *testing.T) {
	mockController := gomock.NewController(t)
	defer mockController.Finish()

	mockLib := mocklib.NewMockLib(mockController)
	mockRemotePeer := mocklib.NewMockPeer(mockController)

	mockLib.EXPECT().NewPeer(testRemoteNodeID, testRemotePeer).Return(mockRemotePeer)

	ctx := context.Background()

	callDiscoverChan := make(chan hubCallDiscover)
	doneChan := make(chan struct{})

	go func() {
		callDiscover := <-callDiscoverChan
		assert.Equal(t, callDiscover.PeerID.Address, testRemotePeer)
		callDiscover.PeersChan <- pb.Peer{Id: testRemoteNodeID, Address: testRemotePeer}
		close(callDiscover.PeersChan)
		doneChan <- struct{}{}

	}()

	c := circleProcess{
		lib:              mockLib,
		callDiscoverChan: callDiscoverChan,
	}

	msg := hubCallJoin{PeerID: pb.PeerID{Address: testRemotePeer}}

	c.eventJoin(ctx, msg)

	assert.Nil(t, c.peers)
	select {
	case <-doneChan:
	case <-time.NewTimer(1 * time.Second).C:
		t.Error("timeout")
	}

}
