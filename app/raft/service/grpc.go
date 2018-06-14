package service

import (
	"context"
	"errors"

	"github.com/stratumn/alice/app/raft/grpc"
	"github.com/stratumn/alice/app/raft/pb"
	"github.com/stratumn/alice/app/raft/protocol"
)

var (
	// ErrUnavailable is returned from gRPC methods when the service is not
	// available.
	ErrUnavailable = errors.New("the service is not available")
)

type grpcReceiver struct {
	hub *protocol.Hub
}

// Init connects GRPC receiver to the hub
func (r *grpcReceiver) Init(hub *protocol.Hub) {
	r.hub = hub
}

// Start starts the node
func (r *grpcReceiver) Start(_ context.Context, _ *grpc.Empty) (*grpc.Empty, error) {

	if r.hub == nil {
		return nil, ErrUnavailable
	}

	r.hub.Start()

	return &grpc.Empty{}, nil

}

// Stop stops the node
func (r *grpcReceiver) Stop(_ context.Context, _ *grpc.Empty) (*grpc.Empty, error) {

	if r.hub == nil {
		return nil, ErrUnavailable
	}

	r.hub.Stop()

	return &grpc.Empty{}, nil
}

// Status enqueries node status
func (r *grpcReceiver) Status(ctx context.Context, _ *grpc.Empty) (*grpc.StatusInfo, error) {
	if r.hub == nil {
		return nil, ErrUnavailable
	}

	statusInfo := r.hub.Status()

	return &grpc.StatusInfo{
		Running: statusInfo.Running,
		Id:      statusInfo.Id,
	}, nil

}

// Peers returns the list of local peers
func (r *grpcReceiver) Peers(_ *grpc.Empty, ss grpc.Raft_PeersServer) error {

	if r.hub == nil {
		return ErrUnavailable
	}

	peersChan := r.hub.Peers()

	for peer := range peersChan {
		if err := ss.Send(&grpc.Peer{Id: peer.Id, Address: peer.Address}); err != nil {
			return err
		}
	}

	return nil
}

// Discover returns the list of peers of a remote node
func (r *grpcReceiver) Discover(peerID *grpc.PeerID, ss grpc.Raft_DiscoverServer) error {

	if r.hub == nil {
		return ErrUnavailable
	}

	peersChan := r.hub.Discover(pb.PeerID{Address: peerID.Address})

	for peer := range peersChan {
		if err := ss.Send(&grpc.Peer{Id: peer.Id, Address: peer.Address}); err != nil {
			return err
		}
	}

	return nil

}

// Invite adds new node to the cluster
func (r *grpcReceiver) Invite(_ context.Context, peerID *grpc.PeerID) (*grpc.Empty, error) {
	if r.hub == nil {
		return nil, ErrUnavailable
	}

	r.hub.Invite(pb.PeerID{Address: peerID.Address})

	return &grpc.Empty{}, nil

}

// Join starts the node and joins the existing cluster
func (r *grpcReceiver) Join(_ context.Context, peerID *grpc.PeerID) (*grpc.Empty, error) {
	if r.hub == nil {
		return nil, ErrUnavailable
	}

	r.hub.Join(pb.PeerID{Address: peerID.Address})

	return &grpc.Empty{}, nil

}

// Expel removes a node from the cluster
func (r *grpcReceiver) Expel(_ context.Context, peerID *grpc.PeerID) (*grpc.Empty, error) {
	if r.hub == nil {
		return nil, ErrUnavailable
	}

	r.hub.Expel(pb.PeerID{Address: peerID.Address})

	return &grpc.Empty{}, nil

}

// Propose adds an entry in the log
func (r *grpcReceiver) Propose(_ context.Context, proposal *grpc.Proposal) (*grpc.Empty, error) {
	if r.hub == nil {
		return nil, ErrUnavailable
	}

	r.hub.Propose(pb.Proposal{Data: proposal.Data})

	return &grpc.Empty{}, nil

}

// Log lists all committed entries
func (r *grpcReceiver) Log(_ *grpc.Empty, ss grpc.Raft_LogServer) error {

	if r.hub == nil {
		return ErrUnavailable
	}

	entriesChan := r.hub.Log()

	for entry := range entriesChan {
		if err := ss.Send(&grpc.Entry{Index: entry.Index, Data: entry.Data}); err != nil {
			return err
		}
	}

	return nil
}
