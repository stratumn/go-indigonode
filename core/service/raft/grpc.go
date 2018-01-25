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

	ErrInvalidCall = errors.New("invalid / unknown grpc call")
)

type grpcReceiver struct {
	msgStartC    chan<- protocol.MessageStart
	msgStopC     chan<- protocol.MessageStop
	msgStatusC   chan<- protocol.MessageStatus
	msgPeersC    chan<- protocol.MessagePeers
	msgDiscoverC chan<- protocol.MessageDiscover
	msgInviteC   chan<- protocol.MessageInvite
	msgJoinC     chan<- protocol.MessageJoin
	msgExpelC    chan<- protocol.MessageExpel
	msgProposeC  chan<- protocol.MessagePropose
	msgLogC      chan<- protocol.MessageLog
}

func (r *grpcReceiver) Init(
	msgStartC chan<- protocol.MessageStart,
	msgStopC chan<- protocol.MessageStop,
	msgStatusC chan<- protocol.MessageStatus,
	msgPeersC chan<- protocol.MessagePeers,
	msgDiscoverC chan<- protocol.MessageDiscover,
	msgInviteC chan<- protocol.MessageInvite,
	msgJoinC chan<- protocol.MessageJoin,
	msgExpelC chan<- protocol.MessageExpel,
	msgProposeC chan<- protocol.MessagePropose,
	msgLogC chan<- protocol.MessageLog,
) {

	r.msgStartC = msgStartC
	r.msgStopC = msgStopC
	r.msgStatusC = msgStatusC
	r.msgPeersC = msgPeersC
	r.msgDiscoverC = msgDiscoverC
	r.msgInviteC = msgInviteC
	r.msgJoinC = msgJoinC
	r.msgExpelC = msgExpelC
	r.msgProposeC = msgProposeC
	r.msgLogC = msgLogC
}

func (r *grpcReceiver) Start(_ context.Context, _ *pb.Empty) (*pb.Empty, error) {

	if r.msgStartC == nil {
		return nil, ErrUnavailable
	}

	r.msgStartC <- protocol.MessageStart{}

	return &pb.Empty{}, nil

}

func (r *grpcReceiver) Stop(_ context.Context, _ *pb.Empty) (*pb.Empty, error) {

	if r.msgStopC == nil {
		return nil, ErrUnavailable
	}

	r.msgStopC <- protocol.MessageStop{}

	return &pb.Empty{}, nil
}

func (r *grpcReceiver) Status(ctx context.Context, _ *pb.Empty) (*pb.StatusInfo, error) {
	if r.msgStatusC == nil {
		return nil, ErrUnavailable
	}

	statusInfoC := make(chan pb.StatusInfo)
	r.msgStatusC <- protocol.MessageStatus{
		StatusInfoC: statusInfoC,
	}

	select {
	case statusInfo := <-statusInfoC:
		return &statusInfo, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}

}

func (r *grpcReceiver) Peers(_ *pb.Empty, ss pb.Raft_PeersServer) error {

	if r.msgPeersC == nil {
		return ErrUnavailable
	}

	peersC := make(chan pb.Peer)
	r.msgPeersC <- protocol.MessagePeers{
		PeersC: peersC,
	}

	for peer := range peersC {
		if err := ss.Send(&peer); err != nil {
			return err
		}
	}

	return nil
}

func (r *grpcReceiver) Discover(peerID *pb.PeerID, ss pb.Raft_DiscoverServer) error {

	if r.msgDiscoverC == nil {
		return ErrUnavailable
	}

	peersC := make(chan pb.Peer)
	r.msgDiscoverC <- protocol.MessageDiscover{
		PeerID: *peerID,
		PeersC: peersC,
	}

	for peer := range peersC {
		if err := ss.Send(&peer); err != nil {
			return err
		}
	}

	return nil

}

func (r *grpcReceiver) Invite(_ context.Context, peerID *pb.PeerID) (*pb.Empty, error) {
	if r.msgInviteC == nil {
		return nil, ErrUnavailable
	}

	r.msgInviteC <- protocol.MessageInvite{
		PeerID: *peerID,
	}

	return &pb.Empty{}, nil

}

func (r *grpcReceiver) Join(_ context.Context, peerID *pb.PeerID) (*pb.Empty, error) {
	if r.msgJoinC == nil {
		return nil, ErrUnavailable
	}

	r.msgJoinC <- protocol.MessageJoin{
		PeerID: *peerID,
	}

	return &pb.Empty{}, nil

}

func (r *grpcReceiver) Expel(_ context.Context, peerID *pb.PeerID) (*pb.Empty, error) {
	if r.msgExpelC == nil {
		return nil, ErrUnavailable
	}

	r.msgExpelC <- protocol.MessageExpel{
		PeerID: *peerID,
	}

	return &pb.Empty{}, nil

}

func (r *grpcReceiver) Propose(_ context.Context, proposal *pb.Proposal) (*pb.Empty, error) {
	if r.msgProposeC == nil {
		return nil, ErrUnavailable
	}

	r.msgProposeC <- protocol.MessagePropose{
		Proposal: *proposal,
	}

	return &pb.Empty{}, nil

}

func (r *grpcReceiver) Log(_ *pb.Empty, ss pb.Raft_LogServer) error {

	if r.msgLogC == nil {
		return ErrUnavailable
	}

	entriesC := make(chan pb.Entry)
	r.msgLogC <- protocol.MessageLog{
		EntriesC: entriesC,
	}

	for entry := range entriesC {
		if err := ss.Send(&entry); err != nil {
			return err
		}
	}

	return nil
}
