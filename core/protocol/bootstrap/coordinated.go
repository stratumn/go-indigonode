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

package bootstrap

import (
	"context"

	"github.com/pkg/errors"
	"github.com/stratumn/alice/core/protector"
	"github.com/stratumn/alice/core/protocol/bootstrap/proposal"
	"github.com/stratumn/alice/core/streamutil"
	pb "github.com/stratumn/alice/pb/bootstrap"
	protectorpb "github.com/stratumn/alice/pb/protector"

	protobuf "gx/ipfs/QmRDePEiL4Yupq5EkcK3L3ko3iMgYaqUdLu7xc1kqs7dnV/go-multicodec/protobuf"
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
	PrivateCoordinatedConfigPID = protocol.ID("/alice/indigo/bootstrap/private/coordinated/config/v1.0.0")

	// PrivateCoordinatedProposePID is the protocol for receiving network update
	// proposals from peers.
	// Network participants should implement this protocol.
	PrivateCoordinatedProposePID = protocol.ID("/alice/indigo/bootstrap/private/coordinated/propose/v1.0.0")
)

// CoordinatedHandler is the handler for a non-coordinator node
// in a private network that has a coordinator.
type CoordinatedHandler struct {
	coordinatorID peer.ID
	host          ihost.Host
	networkConfig protector.NetworkConfig
	proposalStore proposal.Store
}

// NewCoordinatedHandler returns a Handler for a non-coordinator node.
func NewCoordinatedHandler(
	ctx context.Context,
	host ihost.Host,
	networkMode *protector.NetworkMode,
	networkConfig protector.NetworkConfig,
	proposalStore proposal.Store,
) (Handler, error) {
	event := log.EventBegin(ctx, "Coordinated.New", logging.Metadata{
		"coordinatorID": networkMode.CoordinatorID.Pretty(),
	})
	defer event.Done()

	handler := CoordinatedHandler{
		coordinatorID: networkMode.CoordinatorID,
		host:          host,
		networkConfig: networkConfig,
		proposalStore: proposalStore,
	}

	err := handler.handshake(ctx)
	if err != nil {
		return nil, err
	}

	host.SetStreamHandler(
		PrivateCoordinatedConfigPID,
		streamutil.WithAutoClose(log, "Coordinated.HandleConfigUpdate", handler.HandleConfigUpdate),
	)

	host.SetStreamHandler(
		PrivateCoordinatedProposePID,
		streamutil.WithAutoClose(log, "Coordinated.HandlePropose", handler.HandlePropose),
	)

	return &handler, nil
}

// handshake connects to the coordinator for the initial handshake.
// The node is expected to receive the network configuration during
// that handshake.
func (h *CoordinatedHandler) handshake(ctx context.Context) error {
	event := log.EventBegin(ctx, "Coordinated.handshake")
	defer event.Done()

	err := h.host.Connect(ctx, h.host.Peerstore().PeerInfo(h.coordinatorID))
	if err != nil {
		return protector.ErrConnectionRefused
	}

	stream, err := h.host.NewStream(ctx, h.coordinatorID, PrivateCoordinatorHandshakePID)
	if err != nil {
		return protector.ErrConnectionRefused
	}

	defer func() {
		err := stream.Close()
		if err != nil {
			event.Append(logging.Metadata{"close_err": err.Error()})
		}
	}()

	enc := protobuf.Multicodec(nil).Encoder(stream)
	if err := enc.Encode(&pb.Hello{}); err != nil {
		event.SetError(errors.WithStack(err))
		return protector.ErrConnectionRefused
	}

	dec := protobuf.Multicodec(nil).Decoder(stream)
	var networkConfig protectorpb.NetworkConfig
	if err := dec.Decode(&networkConfig); err != nil {
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

// HandleConfigUpdate receives updates to the network configuration.
func (h *CoordinatedHandler) HandleConfigUpdate(ctx context.Context, stream inet.Stream, event *logging.EventInProgress) error {
	dec := protobuf.Multicodec(nil).Decoder(stream)
	var networkConfig protectorpb.NetworkConfig
	if err := dec.Decode(&networkConfig); err != nil {
		return errors.WithStack(err)
	}

	if !networkConfig.ValidateSignature(ctx, h.coordinatorID) {
		return protectorpb.ErrInvalidSignature
	}

	err := h.networkConfig.Reset(ctx, &networkConfig)
	if err != nil {
		return err
	}

	go func() {
		for _, c := range h.host.Network().Conns() {
			if !h.networkConfig.IsAllowed(ctx, c.RemotePeer()) {
				err = c.Close()
				if err != nil {
					log.Event(ctx, "CloseErr", c.RemotePeer(), logging.Metadata{"err": err.Error()})
				}
			}
		}
	}()

	return nil
}

// HandlePropose handles an incoming network update proposal.
func (h *CoordinatedHandler) HandlePropose(ctx context.Context, stream inet.Stream, event *logging.EventInProgress) error {
	dec := protobuf.Multicodec(nil).Decoder(stream)
	var updateReq pb.UpdateProposal
	if err := dec.Decode(&updateReq); err != nil {
		return errors.WithStack(err)
	}

	req := &proposal.Request{}
	err := req.FromUpdateProposal(&updateReq)
	if err != nil {
		return err
	}

	return h.proposalStore.AddRequest(ctx, req)
}

// AddNode sends a proposal to add the node to the coordinator.
// Only the coordinator is allowed to make changes to the network config.
func (h *CoordinatedHandler) AddNode(ctx context.Context, peerID peer.ID, addr multiaddr.Multiaddr, info []byte) error {
	event := log.EventBegin(ctx, "Coordinated.AddNode", peerID)
	defer event.Done()

	stream, err := h.host.NewStream(ctx, h.coordinatorID, PrivateCoordinatorProposePID)
	if err != nil {
		event.SetError(errors.WithStack(err))
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

	enc := protobuf.Multicodec(nil).Encoder(stream)
	err = enc.Encode(req)
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

	stream, err := h.host.NewStream(ctx, h.coordinatorID, PrivateCoordinatorProposePID)
	if err != nil {
		event.SetError(errors.WithStack(err))
		return err
	}

	defer stream.Close()

	req := &pb.NodeIdentity{PeerId: []byte(peerID)}

	enc := protobuf.Multicodec(nil).Encoder(stream)
	err = enc.Encode(req)
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

	v, err := proposal.NewVote(h.host.Peerstore().PrivKey(h.host.ID()), r)
	if err != nil {
		event.SetError(err)
		return err
	}

	stream, err := h.host.NewStream(ctx, h.coordinatorID, PrivateCoordinatorVotePID)
	if err != nil {
		event.SetError(errors.WithStack(err))
		return err
	}

	defer stream.Close()

	enc := protobuf.Multicodec(nil).Encoder(stream)
	err = enc.Encode(v.ToProtoVote())
	if err != nil {
		event.SetError(errors.WithStack(err))
		return err
	}

	return h.proposalStore.Remove(ctx, peerID)
}

// Reject ignores a proposal to add or remove a node.
func (h *CoordinatedHandler) Reject(context.Context, peer.ID) error {
	return nil
}

// CompleteBootstrap can't be used by a coordinated node.
// Only the coordinator can complete the bootstrap phase.
func (h *CoordinatedHandler) CompleteBootstrap(context.Context) error {
	return ErrInvalidOperation
}

// Close removes the protocol handlers.
func (h *CoordinatedHandler) Close(ctx context.Context) {
	log.Event(ctx, "Coordinated.Close")
	h.host.RemoveStreamHandler(PrivateCoordinatedConfigPID)
	h.host.RemoveStreamHandler(PrivateCoordinatedProposePID)
}
