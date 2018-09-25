// Copyright © 2017-2018 Stratumn SAS
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as
// published by the Free Software Foundation, either version 3 of the
// License, or (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <https://www.gnu.org/licenses/>.

package protocol

import (
	"context"
	"sync"

	"github.com/pkg/errors"
	"github.com/stratumn/go-node/core/app/bootstrap/pb"
	"github.com/stratumn/go-node/core/app/bootstrap/protocol/proposal"
	"github.com/stratumn/go-node/core/monitoring"
	"github.com/stratumn/go-node/core/protector"
	protectorpb "github.com/stratumn/go-node/core/protector/pb"
	"github.com/stratumn/go-node/core/streamutil"

	"gx/ipfs/QmPvyPwuCgJ7pDmrKDxRtsScJgBaM5h4EpRL2qQJsmXf4n/go-libp2p-crypto"
	"gx/ipfs/QmQsErDt8Qgw1XrsXf2BpEzDgGWtB1YLsTAARBup5b6B9W/go-libp2p-peer"
	"gx/ipfs/QmYmsdtJ3HsodkePE3eU3TsCaP2YvPZJ4LoXnNkDE5Tpt7/go-multiaddr"
	inet "gx/ipfs/QmZNJyx9GGCX4GeuHnLB8fxaxMLs4MjTjHokxfQcCd6Nve/go-libp2p-net"
	"gx/ipfs/QmZNkThpqfVXs9GNbexPrfBbXSLNYeKrE7jwFM2oqHbyqN/go-libp2p-protocol"
	"gx/ipfs/Qmda4cPRvSRyox3SqgJN6DfSZGU5TtHufPTp9uXjFj71X6/go-libp2p-peerstore"
	ihost "gx/ipfs/QmeMYW7Nj8jnnEfs9qhm7SxKkoDPUWXu3MsxX6BFwz34tf/go-libp2p-host"
)

var (
	// PrivateCoordinatorHandshakePID is the protocol for handling handshake
	// messages and sending the network participants list.
	// Only the network coordinator should implement this protocol.
	PrivateCoordinatorHandshakePID = protocol.ID("/indigo/node/bootstrap/private/coordinator/handshake/v1.0.0")

	// PrivateCoordinatorProposePID is the protocol for receiving network update
	// proposals from peers.
	// Only the network coordinator should implement this protocol.
	PrivateCoordinatorProposePID = protocol.ID("/indigo/node/bootstrap/private/coordinator/propose/v1.0.0")

	// PrivateCoordinatorVotePID is the protocol for receiving votes
	// from network participants.
	// Only the network coordinator should implement this protocol.
	PrivateCoordinatorVotePID = protocol.ID("/indigo/node/bootstrap/private/coordinator/vote/v1.0.0")
)

// Errors used by the coordinator.
var (
	ErrUnknownNode = errors.New("unknown node: no addresses available")
)

// CoordinatorHandler is the handler for the coordinator
// of a private network.
type CoordinatorHandler struct {
	host           ihost.Host
	streamProvider streamutil.Provider
	networkConfig  protector.NetworkConfig
	proposalStore  proposal.Store
}

// NewCoordinatorHandler returns a Handler for a coordinator node.
func NewCoordinatorHandler(
	host ihost.Host,
	streamProvider streamutil.Provider,
	networkConfig protector.NetworkConfig,
	proposalStore proposal.Store,
) Handler {
	handler := CoordinatorHandler{
		host:           host,
		streamProvider: streamProvider,
		networkConfig:  networkConfig,
		proposalStore:  proposalStore,
	}

	host.SetStreamHandler(
		PrivateCoordinatorHandshakePID,
		streamutil.WithAutoClose("bootstrap", "HandleHandshake", handler.HandleHandshake),
	)

	host.SetStreamHandler(
		PrivateCoordinatorProposePID,
		streamutil.WithAutoClose("bootstrap", "HandlePropose", handler.HandlePropose),
	)

	host.SetStreamHandler(
		PrivateCoordinatorVotePID,
		streamutil.WithAutoClose("bootstrap", "HandleVote", handler.HandleVote),
	)

	return &handler
}

// Close removes the protocol handlers.
func (h *CoordinatorHandler) Close(ctx context.Context) {
	_, span := monitoring.StartSpan(ctx, "bootstrap", "Close")
	defer span.End()

	h.host.RemoveStreamHandler(PrivateCoordinatorHandshakePID)
	h.host.RemoveStreamHandler(PrivateCoordinatorProposePID)
	h.host.RemoveStreamHandler(PrivateCoordinatorVotePID)
}

// ValidateSender rejects unauthorized requests.
// This should already be done at the connection level by our protector
// component, but it's always better to have multi-level security.
func (h *CoordinatorHandler) ValidateSender(ctx context.Context, peerID peer.ID) error {
	ctx, span := monitoring.StartSpan(ctx, "bootstrap", "ValidateSender", monitoring.SpanOptionPeerID(peerID))
	defer span.End()

	networkState := h.networkConfig.NetworkState(ctx)
	allowed := h.networkConfig.IsAllowed(ctx, peerID)

	// Once the bootstrap is complete, we reject non-white-listed peers.
	if !allowed && networkState == protectorpb.NetworkState_PROTECTED {
		span.SetStatus(monitoring.NewStatus(
			monitoring.StatusCodePermissionDenied,
			protector.ErrConnectionRefused.Error()))
		return protector.ErrConnectionRefused
	}

	return nil
}

// HandleHandshake handles an incoming handshake and responds with the network
// configuration if handshake succeeds.
func (h *CoordinatorHandler) HandleHandshake(
	ctx context.Context,
	span *monitoring.Span,
	stream inet.Stream,
	codec streamutil.Codec,
) error {
	remoteID := stream.Conn().RemotePeer()
	err := h.ValidateSender(ctx, remoteID)
	if err != nil {
		return err
	}

	var hello pb.Hello
	if err := codec.Decode(&hello); err != nil {
		return protector.ErrConnectionRefused
	}

	allowed := h.networkConfig.IsAllowed(ctx, remoteID)

	// We should not reveal network participants to unwanted peers.
	if !allowed {
		return codec.Encode(&protectorpb.NetworkConfig{})
	}

	networkConfig := h.networkConfig.Copy(ctx)
	return codec.Encode(&networkConfig)
}

// HandlePropose handles an incoming network update proposal.
func (h *CoordinatorHandler) HandlePropose(
	ctx context.Context,
	span *monitoring.Span,
	stream inet.Stream,
	codec streamutil.Codec,
) error {
	remotePeer := stream.Conn().RemotePeer()
	err := h.ValidateSender(ctx, remotePeer)
	if err != nil {
		return err
	}

	var nodeID pb.NodeIdentity
	if err := codec.Decode(&nodeID); err != nil {
		return protector.ErrConnectionRefused
	}

	peerID, err := peer.IDFromBytes(nodeID.PeerId)
	if err != nil {
		return codec.Encode(&pb.Ack{Error: proposal.ErrInvalidPeerID.Error()})
	}

	// Populate address from PeerStore.
	if len(nodeID.PeerAddr) == 0 {
		pi := h.host.Peerstore().PeerInfo(peerID)
		if len(pi.Addrs) > 0 {
			nodeID.PeerAddr = pi.Addrs[0].Bytes()
		}
	}

	if h.networkConfig.NetworkState(ctx) == protectorpb.NetworkState_BOOTSTRAP {
		r, err := proposal.NewAddRequest(&nodeID)
		if err != nil {
			return codec.Encode(&pb.Ack{Error: err.Error()})
		}

		if r.PeerID != remotePeer {
			return codec.Encode(&pb.Ack{Error: proposal.ErrInvalidPeerAddr.Error()})
		}

		err = h.proposalStore.AddRequest(ctx, r)
		if err != nil {
			return codec.Encode(&pb.Ack{Error: err.Error()})
		}
	} else {
		allowed := h.networkConfig.IsAllowed(ctx, peerID)
		var req *proposal.Request
		if allowed {
			req, err = proposal.NewRemoveRequest(&nodeID)
		} else {
			req, err = proposal.NewAddRequest(&nodeID)
		}

		if err != nil {
			return codec.Encode(&pb.Ack{Error: err.Error()})
		}

		err = h.proposalStore.AddRequest(ctx, req)
		if err != nil {
			return codec.Encode(&pb.Ack{Error: err.Error()})
		}

		if req.Type == proposal.RemoveNode {
			h.SendProposal(ctx, req)
		}
	}

	return codec.Encode(&pb.Ack{})
}

// HandleVote handles an incoming vote.
func (h *CoordinatorHandler) HandleVote(
	ctx context.Context,
	span *monitoring.Span,
	stream inet.Stream,
	codec streamutil.Codec,
) error {
	err := h.ValidateSender(ctx, stream.Conn().RemotePeer())
	if err != nil {
		return err
	}

	if h.networkConfig.NetworkState(ctx) == protectorpb.NetworkState_BOOTSTRAP {
		return codec.Encode(&pb.Ack{Error: ErrInvalidOperation.Error()})
	}

	var msg pb.Vote
	if err := codec.Decode(&msg); err != nil {
		return protector.ErrConnectionRefused
	}

	vote := &proposal.Vote{}
	err = vote.FromProtoVote(&msg)
	if err != nil {
		return codec.Encode(&pb.Ack{Error: err.Error()})
	}

	err = h.proposalStore.AddVote(ctx, vote)
	if err != nil {
		return codec.Encode(&pb.Ack{Error: err.Error()})
	}

	thresholdReached, err := h.voteThresholdReached(ctx, vote.PeerID)
	if err != nil {
		return codec.Encode(&pb.Ack{Error: err.Error()})
	}

	if thresholdReached {
		err = h.Accept(ctx, vote.PeerID)
		if err != nil {
			return codec.Encode(&pb.Ack{Error: err.Error()})
		}
	}

	return codec.Encode(&pb.Ack{})
}

func (h *CoordinatorHandler) voteThresholdReached(ctx context.Context, peerID peer.ID) (bool, error) {
	votes, err := h.proposalStore.GetVotes(ctx, peerID)
	if err != nil {
		return false, err
	}

	allowed := h.networkConfig.AllowedPeers(ctx)

	// Since all participants except the coordinator and the node that will be removed
	// need to agree, if we're missing more than 2 votes it's impossible that the
	// threshold was reached.
	if len(votes) < len(allowed)-2 {
		return false, nil
	}

	votesMap := make(map[peer.ID]struct{})
	for _, v := range votes {
		pk, err := crypto.UnmarshalPublicKey(v.Signature.PublicKey)
		if err != nil {
			return false, err
		}

		voterID, err := peer.IDFromPublicKey(pk)
		if err != nil {
			return false, err
		}

		votesMap[voterID] = struct{}{}
	}

	for _, p := range allowed {
		if p == peerID || p == h.host.ID() {
			continue
		}

		_, ok := votesMap[p]
		if !ok {
			return false, nil
		}
	}

	return true, nil
}

// Handshake sends the current network configuration to all participants.
func (h *CoordinatorHandler) Handshake(ctx context.Context) error {
	ctx, span := monitoring.StartSpan(ctx, "bootstrap", "Handshake")
	defer span.End()

	h.SendNetworkConfig(ctx)
	return nil
}

// AddNode adds the node to the network configuration and notifies network
// participants.
func (h *CoordinatorHandler) AddNode(ctx context.Context, peerID peer.ID, addr multiaddr.Multiaddr, info []byte) error {
	ctx, span := monitoring.StartSpan(ctx, "bootstrap", "AddNode", monitoring.SpanOptionPeerID(peerID))
	defer span.End()

	if h.networkConfig.IsAllowed(ctx, peerID) {
		return nil
	}

	pstore := h.host.Peerstore()
	if addr != nil {
		pstore.AddAddr(peerID, addr, peerstore.PermanentAddrTTL)
	}

	pi := pstore.PeerInfo(peerID)
	if len(pi.Addrs) == 0 {
		span.SetStatus(monitoring.NewStatus(monitoring.StatusCodeNotFound, ErrUnknownNode.Error()))
		return ErrUnknownNode
	}

	err := h.networkConfig.AddPeer(ctx, peerID, pi.Addrs)
	if err != nil {
		span.SetUnknownError(err)
		return err
	}

	h.SendNetworkConfig(ctx)

	return nil
}

// RemoveNode removes a node from the network configuration
// and notifies network participants.
func (h *CoordinatorHandler) RemoveNode(ctx context.Context, peerID peer.ID) error {
	ctx, span := monitoring.StartSpan(ctx, "bootstrap", "RemoveNode", monitoring.SpanOptionPeerID(peerID))
	defer span.End()

	if peerID == h.host.ID() {
		span.SetStatus(monitoring.NewStatus(monitoring.StatusCodeInvalidArgument, ErrInvalidOperation.Error()))
		return ErrInvalidOperation
	}

	if !h.networkConfig.IsAllowed(ctx, peerID) {
		return nil
	}

	err := h.networkConfig.RemovePeer(ctx, peerID)
	if err != nil {
		span.SetUnknownError(err)
		return err
	}

	Disconnect(ctx, h.host, peerID)

	h.SendNetworkConfig(ctx)

	return nil
}

// Accept accepts a proposal to add or remove a node
// and notifies network participants.
func (h *CoordinatorHandler) Accept(ctx context.Context, peerID peer.ID) error {
	ctx, span := monitoring.StartSpan(ctx, "bootstrap", "Accept", monitoring.SpanOptionPeerID(peerID))
	defer span.End()

	r, err := h.proposalStore.Get(ctx, peerID)
	if err != nil {
		span.SetUnknownError(err)
		return err
	}

	if r == nil {
		span.SetStatus(monitoring.NewStatus(
			monitoring.StatusCodeNotFound,
			proposal.ErrMissingRequest.Error()))
		return proposal.ErrMissingRequest
	}

	err = h.proposalStore.Remove(ctx, peerID)
	if err != nil {
		span.SetUnknownError(err)
		return err
	}

	switch r.Type {
	case proposal.AddNode:
		err = h.AddNode(ctx, peerID, r.PeerAddr, r.Info)
	case proposal.RemoveNode:
		err = h.RemoveNode(ctx, peerID)
	default:
		err = proposal.ErrInvalidRequestType
	}

	if err != nil {
		span.SetUnknownError(err)
		return err
	}

	return nil
}

// Reject ignores a proposal to add or remove a node.
func (h *CoordinatorHandler) Reject(ctx context.Context, peerID peer.ID) error {
	ctx, span := monitoring.StartSpan(ctx, "bootstrap", "Reject", monitoring.SpanOptionPeerID(peerID))
	defer span.End()

	err := h.proposalStore.Remove(ctx, peerID)
	if err != nil {
		span.SetUnknownError(err)
		return err
	}

	return nil
}

// CompleteBootstrap completes the bootstrap phase and notifies
// white-listed network participants.
func (h *CoordinatorHandler) CompleteBootstrap(ctx context.Context) error {
	ctx, span := monitoring.StartSpan(ctx, "bootstrap", "CompleteBootstrap")
	defer span.End()

	if h.networkConfig.NetworkState(ctx) == protectorpb.NetworkState_PROTECTED {
		return nil
	}

	err := h.networkConfig.SetNetworkState(ctx, protectorpb.NetworkState_PROTECTED)
	if err != nil {
		span.SetUnknownError(err)
		return err
	}

	DisconnectUnauthorized(ctx, h.host, h.networkConfig)

	h.SendNetworkConfig(ctx)

	return nil
}

// SendProposal sends a network update proposal to all participants.
// The proposal contains a random challenge and needs to be signed by
// participants to confirm their agreement.
func (h *CoordinatorHandler) SendProposal(ctx context.Context, req *proposal.Request) {
	ctx, span := monitoring.StartSpan(ctx, "bootstrap", "SendProposal", monitoring.SpanOptionPeerID(req.PeerID))
	defer span.End()

	updateReq := req.ToUpdateProposal()
	allowedPeers := h.networkConfig.AllowedPeers(ctx)
	wg := &sync.WaitGroup{}

	for _, peerID := range allowedPeers {
		if peerID == h.host.ID() {
			continue
		}

		wg.Add(1)

		go func(peerID peer.ID) {
			defer wg.Done()

			ctx, span := monitoring.StartSpan(ctx, "bootstrap", "SendProposalToPeer", monitoring.SpanOptionPeerID(peerID))
			defer span.End()

			stream, err := h.streamProvider.NewStream(
				ctx,
				h.host,
				streamutil.OptPeerID(peerID),
				streamutil.OptProtocolIDs(PrivateCoordinatedProposePID),
			)
			if err != nil {
				return
			}

			defer stream.Close()

			err = stream.Codec().Encode(updateReq)
			if err != nil {
				span.SetUnknownError(err)
			}
		}(peerID)
	}

	wg.Wait()
}

// SendNetworkConfig sends the current network configuration to all
// white-listed participants. It logs errors but swallows them.
func (h *CoordinatorHandler) SendNetworkConfig(ctx context.Context) {
	ctx, span := monitoring.StartSpan(ctx, "bootstrap", "SendNetworkConfig")
	defer span.End()

	networkConfig := h.networkConfig.Copy(ctx)
	allowedPeers := h.networkConfig.AllowedPeers(ctx)
	participants.Record(ctx, int64(len(allowedPeers)))

	wg := &sync.WaitGroup{}

	for _, peerID := range allowedPeers {
		if peerID == h.host.ID() {
			continue
		}

		wg.Add(1)

		go func(peerID peer.ID) {
			defer wg.Done()

			ctx, span := monitoring.StartSpan(ctx, "bootstrap", "SendNetworkConfigToPeer", monitoring.SpanOptionPeerID(peerID))
			defer span.End()

			stream, err := h.streamProvider.NewStream(
				ctx,
				h.host,
				streamutil.OptPeerID(peerID),
				streamutil.OptProtocolIDs(PrivateCoordinatedConfigPID),
			)
			if err != nil {
				return
			}

			defer stream.Close()

			err = stream.Codec().Encode(&networkConfig)
			if err != nil {
				span.SetUnknownError(err)
			}
		}(peerID)
	}

	wg.Wait()
}
