package circle

import (
	"context"
	"testing"
	"time"

	"github.com/stratumn/alice/core/protocol/circle/lib"
	"github.com/stratumn/alice/core/protocol/circle/lib/mocklib"
	pb "github.com/stratumn/alice/pb/circle"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestCircleNormal(t *testing.T) {
	mockController := gomock.NewController(t)
	defer mockController.Finish()

	ctx := context.Background()

	mockStorage := mocklib.NewMockStorage(mockController)
	mockNode := mocklib.NewMockNode(mockController)
	mockReady := mocklib.NewMockReady(mockController)

	mockReady.EXPECT().Messages().Return(nil)
	mockReady.EXPECT().Entries().Return(nil)
	mockReady.EXPECT().CommittedEntries().Return(nil)

	mockStorage.EXPECT().Append(nil)
	mockNode.EXPECT().Advance()

	c := circleProcess{
		storage: mockStorage,
		node:    mockNode,
	}

	c.eventCircle(ctx, mockReady)

}

func TestProcessMessageNormal(t *testing.T) {
	mockController := gomock.NewController(t)
	defer mockController.Finish()

	ctx := context.Background()

	mockMessage := mocklib.NewMockMessage(mockController)
	mockLocalPeer := mocklib.NewMockPeer(mockController)
	mockRemotePeer := mocklib.NewMockPeer(mockController)

	mockLocalPeer.EXPECT().Context().AnyTimes().Return(testLocalPeer)
	mockLocalPeer.EXPECT().ID().AnyTimes().Return(testNodeID)
	mockRemotePeer.EXPECT().Context().AnyTimes().Return(testRemotePeer)
	mockRemotePeer.EXPECT().ID().AnyTimes().Return(testRemoteNodeID)

	mockMessage.EXPECT().To().AnyTimes().Return(testRemoteNodeID)
	mockMessage.EXPECT().Data().AnyTimes().Return(testInternodeMessage)

	testMessage := pb.Internode{
		PeerId:  &pb.PeerID{Address: testRemotePeer},
		Message: testInternodeMessage,
	}

	msgCircleToNetChan := make(chan pb.Internode)
	doneChan := make(chan struct{})

	go func() {
		msg := <-msgCircleToNetChan
		assert.Equal(t, testMessage, msg)
		doneChan <- struct{}{}
	}()

	c := circleProcess{
		peers:              []lib.Peer{mockLocalPeer, mockRemotePeer},
		msgCircleToNetChan: msgCircleToNetChan,
	}

	c.processMessage(ctx, mockMessage)

	select {
	case <-doneChan:
	case <-time.NewTimer(1 * time.Second).C:
		t.Error("timeout")
	}
}

func TestProcessMessageRecepientNotFound(t *testing.T) {
	mockController := gomock.NewController(t)
	defer mockController.Finish()

	ctx := context.Background()

	mockMessage := mocklib.NewMockMessage(mockController)
	mockLocalPeer := mocklib.NewMockPeer(mockController)

	mockLocalPeer.EXPECT().Context().AnyTimes().Return(testLocalPeer)
	mockLocalPeer.EXPECT().ID().AnyTimes().Return(testNodeID)

	mockMessage.EXPECT().To().AnyTimes().Return(testRemoteNodeID)
	mockMessage.EXPECT().Data().AnyTimes().Return(testInternodeMessage)

	doneChan := make(chan struct{})

	c := circleProcess{
		peers: []lib.Peer{mockLocalPeer},
	}

	go func() {
		c.processMessage(ctx, mockMessage)
		doneChan <- struct{}{}
	}()

	select {
	case <-doneChan:
	case <-time.NewTimer(1 * time.Second).C:
		t.Error("timeout")
	}
}

func TestProcessCommittedEntryNormal(t *testing.T) {
	mockController := gomock.NewController(t)
	defer mockController.Finish()

	ctx := context.Background()

	mockEntryTwo := mocklib.NewMockEntry(mockController)

	mockEntryTwo.EXPECT().Type().AnyTimes().Return(lib.EntryNormal)
	mockEntryTwo.EXPECT().Data().AnyTimes().Return(testEntryTwo)

	c := circleProcess{
		committed: [][]byte{testEntryOne},
	}

	c.processCommittedEntry(ctx, mockEntryTwo)

	assert.Equal(t, c.committed, [][]byte{testEntryOne, testEntryTwo})

}

func TestProcessCommittedEntryNormalEmpty(t *testing.T) {
	mockController := gomock.NewController(t)
	defer mockController.Finish()

	ctx := context.Background()

	mockEntryTwo := mocklib.NewMockEntry(mockController)

	mockEntryTwo.EXPECT().Type().AnyTimes().Return(lib.EntryNormal)
	mockEntryTwo.EXPECT().Data().AnyTimes().Return(nil)

	c := circleProcess{
		committed: [][]byte{testEntryOne},
	}

	c.processCommittedEntry(ctx, mockEntryTwo)

	assert.Equal(t, c.committed, [][]byte{testEntryOne})

}

func TestProcessCommittedEntryAddNode(t *testing.T) {
	mockController := gomock.NewController(t)
	defer mockController.Finish()

	ctx := context.Background()

	mockLib := mocklib.NewMockLib(mockController)
	mockNode := mocklib.NewMockNode(mockController)
	mockEntry := mocklib.NewMockEntry(mockController)
	mockConfChange := mocklib.NewMockConfChange(mockController)
	mockLocalPeer := mocklib.NewMockPeer(mockController)
	mockRemotePeer := mocklib.NewMockPeer(mockController)

	mockLocalPeer.EXPECT().Context().AnyTimes().Return(testLocalPeer)
	mockLocalPeer.EXPECT().ID().AnyTimes().Return(testNodeID)
	mockRemotePeer.EXPECT().Context().AnyTimes().Return(testRemotePeer)
	mockRemotePeer.EXPECT().ID().AnyTimes().Return(testRemoteNodeID)

	mockEntry.EXPECT().Type().AnyTimes().Return(lib.EntryConfChange)
	mockEntry.EXPECT().Data().AnyTimes().Return(testConfigChangeEntry)
	mockConfChange.EXPECT().Type().AnyTimes().Return(lib.ConfChangeAddNode)
	mockConfChange.EXPECT().NodeID().AnyTimes().Return(testRemoteNodeID)
	mockConfChange.EXPECT().Context().AnyTimes().Return(testRemotePeer)

	mockLib.EXPECT().NewConfChange(uint64(0), lib.ConfChangeAddNode, uint64(0), nil).Return(mockConfChange)

	mockConfChange.EXPECT().Unmarshal(testConfigChangeEntry)
	mockNode.EXPECT().ApplyConfChange(mockConfChange)

	mockLib.EXPECT().NewPeer(testRemoteNodeID, testRemotePeer).Return(mockRemotePeer)

	c := circleProcess{
		lib:            mockLib,
		node:           mockNode,
		peers:          []lib.Peer{mockLocalPeer},
		configChangeID: testConfigChangeID,
		lastNodeID:     testLastNodeID,
	}

	c.processCommittedEntry(ctx, mockEntry)

	assert.Equal(t, c.configChangeID, testConfigChangeID+1)
	assert.Equal(t, c.lastNodeID, testLastNodeID+1)
	assert.Equal(t, c.peers, []lib.Peer{mockLocalPeer, mockRemotePeer})

}

func TestProcessCommittedEntryAddExistingNode(t *testing.T) {
	mockController := gomock.NewController(t)
	defer mockController.Finish()

	ctx := context.Background()

	mockLib := mocklib.NewMockLib(mockController)
	mockNode := mocklib.NewMockNode(mockController)
	mockEntry := mocklib.NewMockEntry(mockController)
	mockConfChange := mocklib.NewMockConfChange(mockController)
	mockLocalPeer := mocklib.NewMockPeer(mockController)
	mockRemotePeer := mocklib.NewMockPeer(mockController)

	mockLocalPeer.EXPECT().Context().AnyTimes().Return(testLocalPeer)
	mockLocalPeer.EXPECT().ID().AnyTimes().Return(testNodeID)
	mockRemotePeer.EXPECT().Context().AnyTimes().Return(testRemotePeer)
	mockRemotePeer.EXPECT().ID().AnyTimes().Return(testRemoteNodeID)

	mockEntry.EXPECT().Type().AnyTimes().Return(lib.EntryConfChange)
	mockEntry.EXPECT().Data().AnyTimes().Return(testConfigChangeEntry)
	mockConfChange.EXPECT().Type().AnyTimes().Return(lib.ConfChangeAddNode)
	mockConfChange.EXPECT().NodeID().AnyTimes().Return(testRemoteNodeID)
	mockConfChange.EXPECT().Context().AnyTimes().Return(testRemotePeer)

	mockLib.EXPECT().NewConfChange(uint64(0), lib.ConfChangeAddNode, uint64(0), nil).Return(mockConfChange)

	mockConfChange.EXPECT().Unmarshal(testConfigChangeEntry)
	mockNode.EXPECT().ApplyConfChange(mockConfChange)

	c := circleProcess{
		lib:            mockLib,
		node:           mockNode,
		peers:          []lib.Peer{mockLocalPeer, mockRemotePeer},
		configChangeID: testConfigChangeID,
		lastNodeID:     testLastNodeID,
	}

	c.processCommittedEntry(ctx, mockEntry)

	assert.Equal(t, c.configChangeID, testConfigChangeID+1)
	assert.Equal(t, c.lastNodeID, testLastNodeID+1)
	assert.Equal(t, c.peers, []lib.Peer{mockLocalPeer, mockRemotePeer})

}

func TestProcessCommittedEntryRemoveNode(t *testing.T) {
	mockController := gomock.NewController(t)
	defer mockController.Finish()

	ctx := context.Background()

	mockLib := mocklib.NewMockLib(mockController)
	mockNode := mocklib.NewMockNode(mockController)
	mockEntry := mocklib.NewMockEntry(mockController)
	mockConfChange := mocklib.NewMockConfChange(mockController)
	mockLocalPeer := mocklib.NewMockPeer(mockController)
	mockRemotePeer := mocklib.NewMockPeer(mockController)

	mockLocalPeer.EXPECT().Context().AnyTimes().Return(testLocalPeer)
	mockLocalPeer.EXPECT().ID().AnyTimes().Return(testNodeID)
	mockRemotePeer.EXPECT().Context().AnyTimes().Return(testRemotePeer)
	mockRemotePeer.EXPECT().ID().AnyTimes().Return(testRemoteNodeID)

	mockEntry.EXPECT().Type().AnyTimes().Return(lib.EntryConfChange)
	mockEntry.EXPECT().Data().AnyTimes().Return(testConfigChangeEntry)
	mockConfChange.EXPECT().Type().AnyTimes().Return(lib.ConfChangeRemoveNode)
	mockConfChange.EXPECT().NodeID().AnyTimes().Return(testRemoteNodeID)
	mockConfChange.EXPECT().Context().AnyTimes().Return(testRemotePeer)

	mockLib.EXPECT().NewConfChange(uint64(0), lib.ConfChangeAddNode, uint64(0), nil).Return(mockConfChange)

	mockConfChange.EXPECT().Unmarshal(testConfigChangeEntry)
	mockNode.EXPECT().ApplyConfChange(mockConfChange)

	c := circleProcess{
		lib:            mockLib,
		node:           mockNode,
		nodeID:         testNodeID,
		localPeer:      testLocalPeer,
		peers:          []lib.Peer{mockLocalPeer, mockRemotePeer},
		configChangeID: testConfigChangeID,
		lastNodeID:     testLastNodeID,
	}

	c.processCommittedEntry(ctx, mockEntry)

	assert.Equal(t, c.configChangeID, testConfigChangeID+1)
	assert.Equal(t, c.lastNodeID, testLastNodeID)
	assert.Equal(t, c.peers, []lib.Peer{mockLocalPeer})

}

func TestProcessCommittedEntryRemoveNodeAndStop(t *testing.T) {
	mockController := gomock.NewController(t)
	defer mockController.Finish()

	ctx := context.Background()

	mockLib := mocklib.NewMockLib(mockController)
	mockNode := mocklib.NewMockNode(mockController)
	mockEntry := mocklib.NewMockEntry(mockController)
	mockConfChange := mocklib.NewMockConfChange(mockController)
	mockLocalPeer := mocklib.NewMockPeer(mockController)
	mockRemotePeer := mocklib.NewMockPeer(mockController)

	mockLocalPeer.EXPECT().Context().AnyTimes().Return(testLocalPeer)
	mockLocalPeer.EXPECT().ID().AnyTimes().Return(testNodeID)
	mockRemotePeer.EXPECT().Context().AnyTimes().Return(testRemotePeer)
	mockRemotePeer.EXPECT().ID().AnyTimes().Return(testRemoteNodeID)

	mockEntry.EXPECT().Type().AnyTimes().Return(lib.EntryConfChange)
	mockEntry.EXPECT().Data().AnyTimes().Return(testConfigChangeEntry)
	mockConfChange.EXPECT().Type().AnyTimes().Return(lib.ConfChangeRemoveNode)
	mockConfChange.EXPECT().NodeID().AnyTimes().Return(testNodeID)
	mockConfChange.EXPECT().Context().AnyTimes().Return(testLocalPeer)

	mockLib.EXPECT().NewConfChange(uint64(0), lib.ConfChangeAddNode, uint64(0), nil).Return(mockConfChange)

	mockConfChange.EXPECT().Unmarshal(testConfigChangeEntry)
	mockNode.EXPECT().ApplyConfChange(mockConfChange)

	callStopChan := make(chan hubCallStop)
	doneChan := make(chan struct{})

	go func() {
		<-callStopChan
		doneChan <- struct{}{}
	}()

	c := circleProcess{
		lib:            mockLib,
		node:           mockNode,
		nodeID:         testNodeID,
		localPeer:      testLocalPeer,
		peers:          []lib.Peer{mockLocalPeer, mockRemotePeer},
		configChangeID: testConfigChangeID,
		lastNodeID:     testLastNodeID,
		callStopChan:   callStopChan,
	}

	c.processCommittedEntry(ctx, mockEntry)

	assert.Equal(t, c.configChangeID, testConfigChangeID+1)
	assert.Equal(t, c.lastNodeID, testLastNodeID)

	select {
	case <-doneChan:
	case <-time.NewTimer(1 * time.Second).C:
		t.Error("timeout")
	}

}
