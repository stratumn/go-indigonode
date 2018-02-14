package raft

import (
	"context"
	"errors"

	protocol "github.com/stratumn/alice/core/protocol/raft"
	pb "github.com/stratumn/alice/grpc/raft"
)

var (
	// ErrUnavailable is returned from gRPC methods when the service is not
	// available.
	ErrUnavailable = errors.New("the service is not available")
)

type grpcReceiver struct {
	msgStartChan    chan<- protocol.MessageStart
	msgStopChan     chan<- protocol.MessageStop
	msgStatusChan   chan<- protocol.MessageStatus
	msgPeersChan    chan<- protocol.MessagePeers
	msgDiscoverChan chan<- protocol.MessageDiscover
	msgInviteChan   chan<- protocol.MessageInvite
	msgJoinChan     chan<- protocol.MessageJoin
	msgExpelChan    chan<- protocol.MessageExpel
	msgProposeChan  chan<- protocol.MessagePropose
	msgLogChan      chan<- protocol.MessageLog
}

func (r *grpcReceiver) Init(
	msgStartChan chan<- protocol.MessageStart,
	msgStopChan chan<- protocol.MessageStop,
	msgStatusChan chan<- protocol.MessageStatus,
	msgPeersChan chan<- protocol.MessagePeers,
	msgDiscoverChan chan<- protocol.MessageDiscover,
	msgInviteChan chan<- protocol.MessageInvite,
	msgJoinChan chan<- protocol.MessageJoin,
	msgExpelChan chan<- protocol.MessageExpel,
	msgProposeChan chan<- protocol.MessagePropose,
	msgLogChan chan<- protocol.MessageLog,
) {

	r.msgStartChan = msgStartChan
	r.msgStopChan = msgStopChan
	r.msgStatusChan = msgStatusChan
	r.msgPeersChan = msgPeersChan
	r.msgDiscoverChan = msgDiscoverChan
	r.msgInviteChan = msgInviteChan
	r.msgJoinChan = msgJoinChan
	r.msgExpelChan = msgExpelChan
	r.msgProposeChan = msgProposeChan
	r.msgLogChan = msgLogChan
}

func (r *grpcReceiver) Start(_ context.Context, _ *pb.Empty) (*pb.Empty, error) {

	if r.msgStartChan == nil {
		return nil, ErrUnavailable
	}

	r.msgStartChan <- protocol.MessageStart{}

	return &pb.Empty{}, nil

}

func (r *grpcReceiver) Stop(_ context.Context, _ *pb.Empty) (*pb.Empty, error) {

	if r.msgStopChan == nil {
		return nil, ErrUnavailable
	}

	r.msgStopChan <- protocol.MessageStop{}

	return &pb.Empty{}, nil
}

func (r *grpcReceiver) Status(ctx context.Context, _ *pb.Empty) (*pb.StatusInfo, error) {
	if r.msgStatusChan == nil {
		return nil, ErrUnavailable
	}

	statusInfoChan := make(chan pb.StatusInfo)
	r.msgStatusChan <- protocol.MessageStatus{
		StatusInfoChan: statusInfoChan,
	}

	select {
	case statusInfo := <-statusInfoChan:
		return &statusInfo, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}

}

func (r *grpcReceiver) Peers(_ *pb.Empty, ss pb.Raft_PeersServer) error {

	if r.msgPeersChan == nil {
		return ErrUnavailable
	}

	peersChan := make(chan pb.Peer)
	r.msgPeersChan <- protocol.MessagePeers{
		PeersChan: peersChan,
	}

	for peer := range peersChan {
		if err := ss.Send(&peer); err != nil {
			return err
		}
	}

	return nil
}

func (r *grpcReceiver) Discover(peerID *pb.PeerID, ss pb.Raft_DiscoverServer) error {

	if r.msgDiscoverChan == nil {
		return ErrUnavailable
	}

	peersChan := make(chan pb.Peer)
	r.msgDiscoverChan <- protocol.MessageDiscover{
		PeerID:    *peerID,
		PeersChan: peersChan,
	}

	for peer := range peersChan {
		if err := ss.Send(&peer); err != nil {
			return err
		}
	}

	return nil

}

func (r *grpcReceiver) Invite(_ context.Context, peerID *pb.PeerID) (*pb.Empty, error) {
	if r.msgInviteChan == nil {
		return nil, ErrUnavailable
	}

	r.msgInviteChan <- protocol.MessageInvite{
		PeerID: *peerID,
	}

	return &pb.Empty{}, nil

}

func (r *grpcReceiver) Join(_ context.Context, peerID *pb.PeerID) (*pb.Empty, error) {
	if r.msgJoinChan == nil {
		return nil, ErrUnavailable
	}

	r.msgJoinChan <- protocol.MessageJoin{
		PeerID: *peerID,
	}

	return &pb.Empty{}, nil

}

func (r *grpcReceiver) Expel(_ context.Context, peerID *pb.PeerID) (*pb.Empty, error) {
	if r.msgExpelChan == nil {
		return nil, ErrUnavailable
	}

	r.msgExpelChan <- protocol.MessageExpel{
		PeerID: *peerID,
	}

	return &pb.Empty{}, nil

}

func (r *grpcReceiver) Propose(_ context.Context, proposal *pb.Proposal) (*pb.Empty, error) {
	if r.msgProposeChan == nil {
		return nil, ErrUnavailable
	}

	r.msgProposeChan <- protocol.MessagePropose{
		Proposal: *proposal,
	}

	return &pb.Empty{}, nil

}

func (r *grpcReceiver) Log(_ *pb.Empty, ss pb.Raft_LogServer) error {

	if r.msgLogChan == nil {
		return ErrUnavailable
	}

	entriesChan := make(chan pb.Entry)
	r.msgLogChan <- protocol.MessageLog{
		EntriesChan: entriesChan,
	}

	for entry := range entriesChan {
		if err := ss.Send(&entry); err != nil {
			return err
		}
	}

	return nil
}
