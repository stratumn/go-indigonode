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

	"github.com/pkg/errors"
	"github.com/stratumn/go-node/core/app/bootstrap/pb"
	"github.com/stratumn/go-node/core/app/bootstrap/protocol/proposal"
	"github.com/stratumn/go-node/core/monitoring"
	"github.com/stratumn/go-node/core/protector"
	protectorpb "github.com/stratumn/go-node/core/protector/pb"
	"github.com/stratumn/go-node/core/streamutil"

	"gx/ipfs/QmQsErDt8Qgw1XrsXf2BpEzDgGWtB1YLsTAARBup5b6B9W/go-libp2p-peer"
	"gx/ipfs/QmYmsdtJ3HsodkePE3eU3TsCaP2YvPZJ4LoXnNkDE5Tpt7/go-multiaddr"
	inet "gx/ipfs/QmZNJyx9GGCX4GeuHnLB8fxaxMLs4MjTjHokxfQcCd6Nve/go-libp2p-net"
	"gx/ipfs/QmZNkThpqfVXs9GNbexPrfBbXSLNYeKrE7jwFM2oqHbyqN/go-libp2p-protocol"
	ihost "gx/ipfs/QmeMYW7Nj8jnnEfs9qhm7SxKkoDPUWXu3MsxX6BFwz34tf/go-libp2p-host"
)

var (
	// PrivateCoordinatedConfigPID is the protocol for receiving network
	// configuration updates in a private network.
	// Network participants should implement this protocol.
	PrivateCoordinatedConfigPID = protocol.ID("/indigo/node/bootstrap/private/coordinated/config/v1.0.0")

	// PrivateCoordinatedProposePID is the protocol for receiving network update
	// proposals from peers.
	// Network participants should implement this protocol.
	PrivateCoordinatedProposePID = protocol.ID("/indigo/node/bootstrap/private/coordinated/propose/v1.0.0")
)

// CoordinatedHandler is the handler for a non-coordinator node
// in a private network that has a coordinator.
type CoordinatedHandler struct {
	coordinatorID  peer.ID
	host           ihost.Host
	streamProvider streamutil.Provider
	networkConfig  protector.NetworkConfig
	proposalStore  proposal.Store
}

// NewCoordinatedHandler returns a Handler for a non-coordinator node.
func NewCoordinatedHandler(
	host ihost.Host,
	streamProvider streamutil.Provider,
	networkMode *protector.NetworkMode,
	networkConfig protector.NetworkConfig,
	proposalStore proposal.Store,
) Handler {
	handler := CoordinatedHandler{
		coordinatorID:  networkMode.CoordinatorID,
		host:           host,
		streamProvider: streamProvider,
		networkConfig:  networkConfig,
		proposalStore:  proposalStore,
	}

	host.SetStreamHandler(
		PrivateCoordinatedConfigPID,
		streamutil.WithAutoClose("bootstrap", "HandleConfigUpdate", handler.HandleConfigUpdate),
	)

	host.SetStreamHandler(
		PrivateCoordinatedProposePID,
		streamutil.WithAutoClose("bootstrap", "HandlePropose", handler.HandlePropose),
	)

	return &handler
}

// Close removes the protocol handlers.
func (h *CoordinatedHandler) Close(ctx context.Context) {
	_, span := monitoring.StartSpan(ctx, "bootstrap", "Close")
	defer span.End()

	h.host.RemoveStreamHandler(PrivateCoordinatedConfigPID)
	h.host.RemoveStreamHandler(PrivateCoordinatedProposePID)
}

// HandleConfigUpdate receives updates to the network configuration.
func (h *CoordinatedHandler) HandleConfigUpdate(
	ctx context.Context,
	span *monitoring.Span,
	stream inet.Stream,
	codec streamutil.Codec,
) error {
	if stream.Conn().RemotePeer() != h.coordinatorID {
		return protector.ErrConnectionRefused
	}

	var networkConfig protectorpb.NetworkConfig
	if err := codec.Decode(&networkConfig); err != nil {
		return errors.WithStack(err)
	}

	if !networkConfig.ValidateSignature(ctx, h.coordinatorID) {
		return protectorpb.ErrInvalidSignature
	}

	err := h.networkConfig.Reset(ctx, &networkConfig)
	if err != nil {
		return err
	}

	participants.Record(ctx, int64(len(h.networkConfig.AllowedPeers(ctx))))

	DisconnectUnauthorized(ctx, h.host, h.networkConfig)

	return nil
}

// HandlePropose handles an incoming network update proposal.
func (h *CoordinatedHandler) HandlePropose(
	ctx context.Context,
	span *monitoring.Span,
	stream inet.Stream,
	codec streamutil.Codec,
) error {
	if stream.Conn().RemotePeer() != h.coordinatorID {
		return protector.ErrConnectionRefused
	}

	var updateReq pb.UpdateProposal
	if err := codec.Decode(&updateReq); err != nil {
		return errors.WithStack(err)
	}

	req := &proposal.Request{}
	err := req.FromUpdateProposal(&updateReq)
	if err != nil {
		return err
	}

	return h.proposalStore.AddRequest(ctx, req)
}

// Handshake connects to the coordinator for the initial handshake.
// The node is expected to receive the network configuration during
// that handshake.
func (h *CoordinatedHandler) Handshake(ctx context.Context) error {
	ctx, span := monitoring.StartSpan(ctx, "bootstrap", "Handshake")
	defer span.End()

	err := h.host.Connect(ctx, h.host.Peerstore().PeerInfo(h.coordinatorID))
	if err != nil {
		span.SetUnknownError(err)
		return protector.ErrConnectionRefused
	}

	stream, err := h.streamProvider.NewStream(
		ctx,
		h.host,
		streamutil.OptPeerID(h.coordinatorID),
		streamutil.OptProtocolIDs(PrivateCoordinatorHandshakePID),
	)
	if err != nil {
		return protector.ErrConnectionRefused
	}

	defer stream.Close()

	if err := stream.Codec().Encode(&pb.Hello{}); err != nil {
		span.SetUnknownError(err)
		return protector.ErrConnectionRefused
	}

	var networkConfig protectorpb.NetworkConfig
	if err := stream.Codec().Decode(&networkConfig); err != nil {
		span.SetUnknownError(err)
		return protector.ErrConnectionRefused
	}

	span.AddIntAttribute("participants_count", int64(len(networkConfig.Participants)))

	// The network is still bootstrapping and we're not whitelisted yet.
	if len(networkConfig.Participants) == 0 {
		return h.AddNode(ctx, h.host.ID(), h.host.Addrs()[0], nil)
	}

	if !networkConfig.ValidateSignature(ctx, h.coordinatorID) {
		span.SetStatus(monitoring.NewStatus(monitoring.StatusCodeInvalidArgument, protectorpb.ErrInvalidSignature.Error()))
		return protectorpb.ErrInvalidSignature
	}

	return h.networkConfig.Reset(ctx, &networkConfig)
}

// AddNode sends a proposal to add the node to the coordinator.
// Only the coordinator is allowed to make changes to the network config.
func (h *CoordinatedHandler) AddNode(ctx context.Context, peerID peer.ID, addr multiaddr.Multiaddr, info []byte) error {
	ctx, span := monitoring.StartSpan(ctx, "bootstrap", "AddNode", monitoring.SpanOptionPeerID(peerID))
	defer span.End()

	stream, err := h.streamProvider.NewStream(
		ctx,
		h.host,
		streamutil.OptPeerID(h.coordinatorID),
		streamutil.OptProtocolIDs(PrivateCoordinatorProposePID),
	)
	if err != nil {
		return err
	}

	defer stream.Close()

	req := &pb.NodeIdentity{
		PeerId:        []byte(peerID),
		IdentityProof: info,
	}
	if addr != nil {
		req.PeerAddr = addr.Bytes()
	}

	err = stream.Codec().Encode(req)
	if err != nil {
		span.SetUnknownError(err)
		return err
	}

	return nil
}

// RemoveNode sends a proposal to remove the node to the coordinator.
// The coordinator will notify each node that needs to sign their agreement.
// The node will then eventually be removed if enough participants agree.
func (h *CoordinatedHandler) RemoveNode(ctx context.Context, peerID peer.ID) error {
	ctx, span := monitoring.StartSpan(ctx, "bootstrap", "RemoveNode", monitoring.SpanOptionPeerID(peerID))
	defer span.End()

	stream, err := h.streamProvider.NewStream(
		ctx,
		h.host,
		streamutil.OptPeerID(h.coordinatorID),
		streamutil.OptProtocolIDs(PrivateCoordinatorProposePID),
	)
	if err != nil {
		return err
	}

	defer stream.Close()

	req := &pb.NodeIdentity{PeerId: []byte(peerID)}
	err = stream.Codec().Encode(req)
	if err != nil {
		span.SetUnknownError(err)
		return err
	}

	return nil
}

// Accept broadcasts a signed message to accept a proposal to add
// or remove a node.
func (h *CoordinatedHandler) Accept(ctx context.Context, peerID peer.ID) error {
	ctx, span := monitoring.StartSpan(ctx, "bootstrap", "Accept", monitoring.SpanOptionPeerID(peerID))
	defer span.End()

	r, err := h.proposalStore.Get(ctx, peerID)
	if err != nil {
		span.SetUnknownError(err)
		return err
	}

	if r == nil {
		span.SetStatus(monitoring.NewStatus(
			monitoring.StatusCodeFailedPrecondition,
			proposal.ErrMissingRequest.Error()))
		return proposal.ErrMissingRequest
	}

	if r.Type != proposal.RemoveNode {
		span.SetStatus(monitoring.NewStatus(
			monitoring.StatusCodeInvalidArgument,
			ErrInvalidOperation.Error()))
		return ErrInvalidOperation
	}

	v, err := proposal.NewVote(ctx, h.host.Peerstore().PrivKey(h.host.ID()), r)
	if err != nil {
		span.SetUnknownError(err)
		return err
	}

	stream, err := h.streamProvider.NewStream(
		ctx,
		h.host,
		streamutil.OptPeerID(h.coordinatorID),
		streamutil.OptProtocolIDs(PrivateCoordinatorVotePID),
	)
	if err != nil {
		return err
	}

	defer stream.Close()

	err = stream.Codec().Encode(v.ToProtoVote())
	if err != nil {
		span.SetUnknownError(err)
		return err
	}

	err = h.proposalStore.Remove(ctx, peerID)
	if err != nil {
		span.SetUnknownError(err)
		return err
	}

	return nil
}

// Reject ignores a proposal to add or remove a node.
func (h *CoordinatedHandler) Reject(ctx context.Context, peerID peer.ID) error {
	ctx, span := monitoring.StartSpan(ctx, "bootstrap", "Reject", monitoring.SpanOptionPeerID(peerID))
	defer span.End()

	err := h.proposalStore.Remove(ctx, peerID)
	if err != nil {
		span.SetUnknownError(err)
		return err
	}

	return nil
}

// CompleteBootstrap can't be used by a coordinated node.
// Only the coordinator can complete the bootstrap phase.
func (h *CoordinatedHandler) CompleteBootstrap(context.Context) error {
	return ErrInvalidOperation
}
