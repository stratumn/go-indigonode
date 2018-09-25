package protocol

import (
	"context"
	"testing"
	"time"

	"github.com/stratumn/go-node/app/raft/pb"
	"github.com/stratumn/go-node/app/raft/protocol/lib"
	"github.com/stratumn/go-node/app/raft/protocol/lib/mocklib"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestPeersNormal(t *testing.T) {
	mockController := gomock.NewController(t)
	defer mockController.Finish()

	mockNode := mocklib.NewMockNode(mockController)
	mockLocalPeer := mocklib.NewMockPeer(mockController)
	mockRemotePeer := mocklib.NewMockPeer(mockController)

	mockLocalPeer.EXPECT().ID().AnyTimes().Return(testNodeID)
	mockLocalPeer.EXPECT().Context().AnyTimes().Return(testLocalPeer)

	mockRemotePeer.EXPECT().ID().AnyTimes().Return(testRemoteNodeID)
	mockRemotePeer.EXPECT().Context().AnyTimes().Return(testRemotePeer)

	ctx := context.Background()

	testPeers := []pb.Peer{
		{Id: testNodeID, Address: testLocalPeer},
		{Id: testRemoteNodeID, Address: testRemotePeer},
	}

	c := circleProcess{
		node:   mockNode,
		nodeID: testNodeID,
		peers:  []lib.Peer{mockLocalPeer, mockRemotePeer},
	}

	peersChan := make(chan pb.Peer)
	doneChan := make(chan struct{})

	go func() {
		var peers []pb.Peer
		for p := range peersChan {
			peers = append(peers, p)
		}
		assert.Equal(t, testPeers, peers)
		doneChan <- struct{}{}
	}()

	msg := hubCallPeers{PeersChan: peersChan}

	c.eventPeers(ctx, msg)

	select {
	case <-doneChan:
	case <-time.NewTimer(1 * time.Second).C:
		t.Error("timeout")
	}

}
