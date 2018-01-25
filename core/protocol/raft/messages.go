package raft

import (
	"github.com/coreos/etcd/raft/raftpb"
	pb "github.com/stratumn/alice/grpc/raft"
)

type MessageStart struct{}

type MessageStop struct{}

type MessageStatus struct {
	StatusInfoC chan<- pb.StatusInfo
}

type MessagePeers struct {
	PeersC chan<- pb.Peer
}

type MessageDiscover struct {
	PeerID pb.PeerID
	PeersC chan<- pb.Peer
}

type MessageInvite struct {
	PeerID pb.PeerID
}

type MessageJoin struct {
	PeerID pb.PeerID
}

type MessageExpel struct {
	PeerID pb.PeerID
}

type MessagePropose struct {
	Proposal pb.Proposal
}

type MessageLog struct {
	EntriesC chan<- pb.Entry
}

type MessageRaft struct {
	PeerID  pb.PeerID
	Message raftpb.Message
}
