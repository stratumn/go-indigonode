package raft

import (
	"github.com/coreos/etcd/raft/raftpb"
	pb "github.com/stratumn/alice/grpc/raft"
)

// MessageStart tells raft to start a single-node clusted
type MessageStart struct{}

// MessageStop tells raft to stop the node
type MessageStop struct{}

// MessageStatus requests raft status
type MessageStatus struct {
	StatusInfoChan chan<- pb.StatusInfo
}

// MessagePeers requests raft peers
type MessagePeers struct {
	PeersChan chan<- pb.Peer
}

// MessageDiscover returns node peers
type MessageDiscover struct {
	PeerID    pb.PeerID
	PeersChan chan<- pb.Peer
}

// MessageInvite adds a node to a cluster
type MessageInvite struct {
	PeerID pb.PeerID
}

// MessageJoin starts raft and joins the cluster
type MessageJoin struct {
	PeerID pb.PeerID
}

// MessageExpel removes a node from the cluster
type MessageExpel struct {
	PeerID pb.PeerID
}

// MessagePropose adds log entry
type MessagePropose struct {
	Proposal pb.Proposal
}

// MessageLog returns log entries
type MessageLog struct {
	EntriesChan chan<- pb.Entry
}

// MessageRaft used to communicate btw raft nodes
type MessageRaft struct {
	PeerID  pb.PeerID
	Message raftpb.Message
}
