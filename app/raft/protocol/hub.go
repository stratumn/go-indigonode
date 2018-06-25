package protocol

import (
	"github.com/stratumn/go-indigonode/app/raft/pb"
)

type hubCallStart struct{}

type hubCallStop struct{}

type hubCallStatus struct {
	StatusInfoChan chan<- pb.StatusInfo
}

type hubCallPeers struct {
	PeersChan chan<- pb.Peer
}

type hubCallDiscover struct {
	PeerID    pb.PeerID
	PeersChan chan<- pb.Peer
}

type hubCallInvite struct {
	PeerID pb.PeerID
}

type hubCallJoin struct {
	PeerID pb.PeerID
}

type hubCallExpel struct {
	PeerID pb.PeerID
}

type hubCallPropose struct {
	Proposal pb.Proposal
}

type hubCallLog struct {
	EntriesChan chan<- pb.Entry
}

// Hub centralizes communication between circle and net processes
type Hub struct {
	callStartChan      chan hubCallStart
	callStopChan       chan hubCallStop
	callStatusChan     chan hubCallStatus
	callPeersChan      chan hubCallPeers
	callDiscoverChan   chan hubCallDiscover
	callInviteChan     chan hubCallInvite
	callJoinChan       chan hubCallJoin
	callExpelChan      chan hubCallExpel
	callProposeChan    chan hubCallPropose
	callLogChan        chan hubCallLog
	msgNetToCircleChan chan pb.Internode
	msgCircleToNetChan chan pb.Internode
}

// NewHub creates an instance of Hub
func NewHub() *Hub {
	h := Hub{
		callStartChan:      make(chan hubCallStart),
		callStopChan:       make(chan hubCallStop),
		callStatusChan:     make(chan hubCallStatus),
		callPeersChan:      make(chan hubCallPeers),
		callDiscoverChan:   make(chan hubCallDiscover),
		callInviteChan:     make(chan hubCallInvite),
		callJoinChan:       make(chan hubCallJoin),
		callExpelChan:      make(chan hubCallExpel),
		callProposeChan:    make(chan hubCallPropose),
		callLogChan:        make(chan hubCallLog),
		msgNetToCircleChan: make(chan pb.Internode),
		msgCircleToNetChan: make(chan pb.Internode),
	}

	return &h
}

func (h *Hub) bindCircle(cp *circleProcess) {
	cp.callStartChan = h.callStartChan
	cp.callStopChan = h.callStopChan
	cp.callStatusChan = h.callStatusChan
	cp.callPeersChan = h.callPeersChan
	cp.callDiscoverChan = h.callDiscoverChan
	cp.callInviteChan = h.callInviteChan
	cp.callJoinChan = h.callJoinChan
	cp.callExpelChan = h.callExpelChan
	cp.callProposeChan = h.callProposeChan
	cp.callLogChan = h.callLogChan
	cp.msgNetToCircleChan = h.msgNetToCircleChan
	cp.msgCircleToNetChan = h.msgCircleToNetChan
}

func (h *Hub) bindNet(np *netProcess) {
	np.callDiscoverChan = h.callDiscoverChan
	np.callPeersChan = h.callPeersChan
	np.msgNetToCircleChan = h.msgNetToCircleChan
	np.msgCircleToNetChan = h.msgCircleToNetChan

}

// Discover returns the list of peers of a remote node
func (h *Hub) Discover(peerID pb.PeerID) <-chan pb.Peer {
	peersChan := make(chan pb.Peer)
	h.callDiscoverChan <- hubCallDiscover{
		PeerID:    peerID,
		PeersChan: peersChan,
	}

	return peersChan
}

// Peers returns the list of local peers
func (h *Hub) Peers() <-chan pb.Peer {
	peersChan := make(chan pb.Peer)
	h.callPeersChan <- hubCallPeers{
		PeersChan: peersChan,
	}
	return peersChan
}

// Start starts the node
func (h *Hub) Start() {
	h.callStartChan <- hubCallStart{}
}

// Stop stops the node
func (h *Hub) Stop() {
	h.callStopChan <- hubCallStop{}
}

// Status enqueries node status
func (h *Hub) Status() pb.StatusInfo {
	statusInfoChan := make(chan pb.StatusInfo)
	h.callStatusChan <- hubCallStatus{
		StatusInfoChan: statusInfoChan,
	}

	return <-statusInfoChan
}

// Invite adds new node to the cluster
func (h *Hub) Invite(peerID pb.PeerID) {
	h.callInviteChan <- hubCallInvite{
		PeerID: peerID,
	}
}

// Join starts the node and joins the existing cluster
func (h *Hub) Join(peerID pb.PeerID) {
	h.callJoinChan <- hubCallJoin{
		PeerID: peerID,
	}
}

// Expel removes a node from the cluster
func (h *Hub) Expel(peerID pb.PeerID) {
	h.callExpelChan <- hubCallExpel{
		PeerID: peerID,
	}
}

// Propose adds an entry in the log
func (h *Hub) Propose(proposal pb.Proposal) {
	h.callProposeChan <- hubCallPropose{
		Proposal: proposal,
	}
}

// Log lists all committed entries
func (h *Hub) Log() <-chan pb.Entry {
	entriesChan := make(chan pb.Entry)
	h.callLogChan <- hubCallLog{
		EntriesChan: entriesChan,
	}
	return entriesChan
}
