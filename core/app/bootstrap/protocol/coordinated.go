// Copyright © 2017-2018 Stratumn SAS
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package protocol

import (
	"context"

	"github.com/pkg/errors"
	"github.com/stratumn/go-indigonode/core/app/bootstrap/pb"
	"github.com/stratumn/go-indigonode/core/app/bootstrap/protocol/proposal"
	"github.com/stratumn/go-indigonode/core/protector"
	protectorpb "github.com/stratumn/go-indigonode/core/protector/pb"
	"github.com/stratumn/go-indigonode/core/streamutil"

	logging "gx/ipfs/QmSpJByNKFX1sCsHBEp3R73FL4NF6FnQTEGyNAXHm2GS52/go-log"
	"gx/ipfs/QmWWQ2Txc2c6tqjsBpzg5Ar652cHPGNsQQp2SejkNmkUMb/go-multiaddr"
	inet "gx/ipfs/QmXoz9o2PT3tEzf7hicegwex5UgVP54n3k82K7jrWFyN86/go-libp2p-net"
	"gx/ipfs/QmZNkThpqfVXs9GNbexPrfBbXSLNYeKrE7jwFM2oqHbyqN/go-libp2p-protocol"
	"gx/ipfs/QmcJukH2sAFjY3HdBKq35WDzWoL3UUu2gt9wdfqZTUyM74/go-libp2p-peer"
	ihost "gx/ipfs/QmfZTdmunzKzAGJrSvXXQbQ5kLLUiEMX5vdwux7iXkdk7D/go-libp2p-host"
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
		streamutil.WithAutoClose(log, "Coordinated.HandleConfigUpdate", handler.HandleConfigUpdate),
	)

	host.SetStreamHandler(
		PrivateCoordinatedProposePID,
		streamutil.WithAutoClose(log, "Coordinated.HandlePropose", handler.HandlePropose),
	)

	return &handler
}

// Close removes the protocol handlers.
func (h *CoordinatedHandler) Close(ctx context.Context) {
	log.Event(ctx, "Coordinated.Close")

	h.host.RemoveStreamHandler(PrivateCoordinatedConfigPID)
	h.host.RemoveStreamHandler(PrivateCoordinatedProposePID)
}

// HandleConfigUpdate receives updates to the network configuration.
func (h *CoordinatedHandler) HandleConfigUpdate(
	ctx context.Context,
	event *logging.EventInProgress,
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

	DisconnectUnauthorized(ctx, h.host, h.networkConfig, event)

	return nil
}

// HandlePropose handles an incoming network update proposal.
func (h *CoordinatedHandler) HandlePropose(
	ctx context.Context,
	event *logging.EventInProgress,
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
	event := log.EventBegin(ctx, "Coordinated.handshake")
	defer event.Done()

	err := h.host.Connect(ctx, h.host.Peerstore().PeerInfo(h.coordinatorID))
	if err != nil {
		event.SetError(err)
		return protector.ErrConnectionRefused
	}

	stream, err := h.streamProvider.NewStream(
		ctx,
		h.host,
		streamutil.OptPeerID(h.coordinatorID),
		streamutil.OptProtocolIDs(PrivateCoordinatorHandshakePID),
		streamutil.OptLog(event),
	)
	if err != nil {
		return protector.ErrConnectionRefused
	}

	defer stream.Close()

	if err := stream.Codec().Encode(&pb.Hello{}); err != nil {
		event.SetError(errors.WithStack(err))
		return protector.ErrConnectionRefused
	}

	var networkConfig protectorpb.NetworkConfig
	if err := stream.Codec().Decode(&networkConfig); err != nil {
		event.SetError(errors.WithStack(err))
		return protector.ErrConnectionRefused
	}

	// The network is still bootstrapping and we're not whitelisted yet.
	if len(networkConfig.Participants) == 0 {
		event.Append(logging.Metadata{"participants_count": 0})
		return h.AddNode(ctx, h.host.ID(), h.host.Addrs()[0], nil)
	}

	if !networkConfig.ValidateSignature(ctx, h.coordinatorID) {
		event.SetError(protectorpb.ErrInvalidSignature)
		return protectorpb.ErrInvalidSignature
	}

	return h.networkConfig.Reset(ctx, &networkConfig)
}

// AddNode sends a proposal to add the node to the coordinator.
// Only the coordinator is allowed to make changes to the network config.
func (h *CoordinatedHandler) AddNode(ctx context.Context, peerID peer.ID, addr multiaddr.Multiaddr, info []byte) error {
	event := log.EventBegin(ctx, "Coordinated.AddNode", peerID)
	defer event.Done()

	stream, err := h.streamProvider.NewStream(
		ctx,
		h.host,
		streamutil.OptPeerID(h.coordinatorID),
		streamutil.OptProtocolIDs(PrivateCoordinatorProposePID),
		streamutil.OptLog(event),
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
		event.SetError(errors.WithStack(err))
		return err
	}

	return nil
}

// RemoveNode sends a proposal to remove the node to the coordinator.
// The coordinator will notify each node that needs to sign their agreement.
// The node will then eventually be removed if enough participants agree.
func (h *CoordinatedHandler) RemoveNode(ctx context.Context, peerID peer.ID) error {
	event := log.EventBegin(ctx, "Coordinated.RemoveNode", peerID)
	defer event.Done()

	stream, err := h.streamProvider.NewStream(
		ctx,
		h.host,
		streamutil.OptPeerID(h.coordinatorID),
		streamutil.OptProtocolIDs(PrivateCoordinatorProposePID),
		streamutil.OptLog(event),
	)
	if err != nil {
		return err
	}

	defer stream.Close()

	req := &pb.NodeIdentity{PeerId: []byte(peerID)}
	err = stream.Codec().Encode(req)
	if err != nil {
		event.SetError(errors.WithStack(err))
		return err
	}

	return nil
}

// Accept broadcasts a signed message to accept a proposal to add
// or remove a node.
func (h *CoordinatedHandler) Accept(ctx context.Context, peerID peer.ID) error {
	event := log.EventBegin(ctx, "Coordinated.Accept", peerID)
	defer event.Done()

	r, err := h.proposalStore.Get(ctx, peerID)
	if err != nil {
		event.SetError(err)
		return err
	}

	if r == nil {
		event.SetError(proposal.ErrMissingRequest)
		return proposal.ErrMissingRequest
	}

	if r.Type != proposal.RemoveNode {
		event.SetError(ErrInvalidOperation)
		return ErrInvalidOperation
	}

	v, err := proposal.NewVote(ctx, h.host.Peerstore().PrivKey(h.host.ID()), r)
	if err != nil {
		event.SetError(err)
		return err
	}

	stream, err := h.streamProvider.NewStream(
		ctx,
		h.host,
		streamutil.OptPeerID(h.coordinatorID),
		streamutil.OptProtocolIDs(PrivateCoordinatorVotePID),
		streamutil.OptLog(event),
	)
	if err != nil {
		return err
	}

	defer stream.Close()

	err = stream.Codec().Encode(v.ToProtoVote())
	if err != nil {
		event.SetError(errors.WithStack(err))
		return err
	}

	return h.proposalStore.Remove(ctx, peerID)
}

// Reject ignores a proposal to add or remove a node.
func (h *CoordinatedHandler) Reject(ctx context.Context, peerID peer.ID) error {
	defer log.EventBegin(ctx, "Coordinated.Reject", peerID).Done()
	return h.proposalStore.Remove(ctx, peerID)
}

// CompleteBootstrap can't be used by a coordinated node.
// Only the coordinator can complete the bootstrap phase.
func (h *CoordinatedHandler) CompleteBootstrap(context.Context) error {
	return ErrInvalidOperation
}
