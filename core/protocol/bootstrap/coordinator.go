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
	"fmt"
	"sync"

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
	"gx/ipfs/QmdeiKhUy1TVGBaKxt7y1QmBDLBdisSrLJ1x58Eoj4PXUh/go-libp2p-peerstore"
	"gx/ipfs/Qme1knMqwt1hKZbc1BmQFmnm9f36nyQGwXxPGVpVJ9rMK5/go-libp2p-crypto"
	ihost "gx/ipfs/QmfZTdmunzKzAGJrSvXXQbQ5kLLUiEMX5vdwux7iXkdk7D/go-libp2p-host"
)

var (
	// PrivateCoordinatorHandshakePID is the protocol for handling handshake
	// messages and sending the network participants list.
	// Only the network coordinator should implement this protocol.
	PrivateCoordinatorHandshakePID = protocol.ID("/alice/indigo/bootstrap/private/coordinator/handshake/v1.0.0")

	// PrivateCoordinatorProposePID is the protocol for receiving network update
	// proposals from peers.
	// Only the network coordinator should implement this protocol.
	PrivateCoordinatorProposePID = protocol.ID("/alice/indigo/bootstrap/private/coordinator/propose/v1.0.0")

	// PrivateCoordinatorVotePID is the protocol for receiving votes
	// from network participants.
	// Only the network coordinator should implement this protocol.
	PrivateCoordinatorVotePID = protocol.ID("/alice/indigo/bootstrap/private/coordinator/vote/v1.0.0")
)

// Errors used by the coordinator.
var (
	ErrUnknownNode = errors.New("unknown node: no addresses available")
)

// CoordinatorHandler is the handler for the coordinator
// of a private network.
type CoordinatorHandler struct {
	host          ihost.Host
	networkConfig protector.NetworkConfig
	proposalStore proposal.Store
}

// NewCoordinatorHandler returns a Handler for a coordinator node.
func NewCoordinatorHandler(
	host ihost.Host,
	networkConfig protector.NetworkConfig,
	proposalStore proposal.Store,
) (Handler, error) {
	handler := CoordinatorHandler{
		host:          host,
		networkConfig: networkConfig,
		proposalStore: proposalStore,
	}

	host.SetStreamHandler(
		PrivateCoordinatorHandshakePID,
		streamutil.WithAutoClose(log, "Coordinator.HandleHandshake", handler.HandleHandshake),
	)

	host.SetStreamHandler(
		PrivateCoordinatorProposePID,
		streamutil.WithAutoClose(log, "Coordinator.HandlePropose", handler.HandlePropose),
	)

	host.SetStreamHandler(
		PrivateCoordinatorVotePID,
		streamutil.WithAutoClose(log, "Coordinator.HandleVote", handler.HandleVote),
	)

	return &handler, nil
}

// validateSender rejects unauthorized requests.
// This should already be done at the connection level by our protector
// component, but it's always better to have multi-level security.
func (h *CoordinatorHandler) validateSender(ctx context.Context, stream inet.Stream, event *logging.EventInProgress) error {
	networkState := h.networkConfig.NetworkState(ctx)
	allowed := h.networkConfig.IsAllowed(ctx, stream.Conn().RemotePeer())

	// Once the bootstrap is complete, we reject non-white-listed peers.
	if !allowed && networkState == protectorpb.NetworkState_PROTECTED {
		if err := stream.Reset(); err != nil {
			event.Append(logging.Metadata{"reset_err": err.Error()})
		}

		return protector.ErrConnectionRefused
	}

	return nil
}

// HandleHandshake handles an incoming handshake and responds with the network
// configuration if handshake succeeds.
func (h *CoordinatorHandler) HandleHandshake(
	ctx context.Context,
	event *logging.EventInProgress,
	stream inet.Stream,
	codec streamutil.Codec,
) error {
	err := h.validateSender(ctx, stream, event)
	if err != nil {
		return err
	}

	var hello pb.Hello
	if err := codec.Decode(&hello); err != nil {
		return protector.ErrConnectionRefused
	}

	allowed := h.networkConfig.IsAllowed(ctx, stream.Conn().RemotePeer())

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
	event *logging.EventInProgress,
	stream inet.Stream,
	codec streamutil.Codec,
) error {
	err := h.validateSender(ctx, stream, event)
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

		if r.PeerID != stream.Conn().RemotePeer() {
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
			go h.SendProposal(ctx, req)
		}
	}

	return codec.Encode(&pb.Ack{})
}

// HandleVote handles an incoming vote.
func (h *CoordinatorHandler) HandleVote(
	ctx context.Context,
	event *logging.EventInProgress,
	stream inet.Stream,
	codec streamutil.Codec,
) error {
	err := h.validateSender(ctx, stream, event)
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

// AddNode adds the node to the network configuration
// and notifies network participants.
func (h *CoordinatorHandler) AddNode(ctx context.Context, peerID peer.ID, addr multiaddr.Multiaddr, info []byte) error {
	event := log.EventBegin(ctx, "Coordinator.AddNode", peerID)
	defer event.Done()

	if h.networkConfig.IsAllowed(ctx, peerID) {
		return nil
	}

	pstore := h.host.Peerstore()
	pi := pstore.PeerInfo(peerID)
	if len(pi.Addrs) == 0 {
		if addr == nil {
			event.SetError(ErrUnknownNode)
			return ErrUnknownNode
		}

		pstore.AddAddr(peerID, addr, peerstore.PermanentAddrTTL)
		pi = pstore.PeerInfo(peerID)
	}

	err := h.networkConfig.AddPeer(ctx, peerID, pi.Addrs)
	if err != nil {
		event.SetError(err)
		return err
	}

	h.SendNetworkConfig(ctx)

	return nil
}

// RemoveNode removes a node from the network configuration
// and notifies network participants.
func (h *CoordinatorHandler) RemoveNode(ctx context.Context, peerID peer.ID) error {
	event := log.EventBegin(ctx, "Coordinator.RemoveNode", peerID)
	defer event.Done()

	if peerID == h.host.ID() {
		event.SetError(ErrInvalidOperation)
		return ErrInvalidOperation
	}

	if !h.networkConfig.IsAllowed(ctx, peerID) {
		return nil
	}

	err := h.networkConfig.RemovePeer(ctx, peerID)
	if err != nil {
		event.SetError(err)
		return err
	}

	for _, c := range h.host.Network().Conns() {
		if c.RemotePeer() == peerID {
			err = c.Close()
			if err != nil {
				event.Append(logging.Metadata{"close_err": err.Error()})
			}
		}
	}

	h.SendNetworkConfig(ctx)

	return nil
}

// Accept accepts a proposal to add or remove a node
// and notifies network participants.
func (h *CoordinatorHandler) Accept(ctx context.Context, peerID peer.ID) error {
	event := log.EventBegin(ctx, "Coordinator.Accept", peerID)
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

	err = h.proposalStore.Remove(ctx, peerID)
	if err != nil {
		event.SetError(err)
		return err
	}

	if r.Type == proposal.RemoveNode {
		return h.RemoveNode(ctx, peerID)
	}

	if r.PeerAddr == nil {
		event.SetError(proposal.ErrMissingPeerAddr)
		return proposal.ErrMissingPeerAddr
	}

	if h.networkConfig.IsAllowed(ctx, peerID) {
		// Nothing to do, peer was already added.
		return nil
	}

	err = h.networkConfig.AddPeer(ctx, peerID, []multiaddr.Multiaddr{r.PeerAddr})
	if err != nil {
		event.SetError(err)
		return err
	}

	h.SendNetworkConfig(ctx)

	return nil
}

// Reject ignores a proposal to add or remove a node.
func (h *CoordinatorHandler) Reject(ctx context.Context, peerID peer.ID) error {
	defer log.EventBegin(ctx, "Coordinator.Reject", peerID).Done()
	return h.proposalStore.Remove(ctx, peerID)
}

// CompleteBootstrap completes the bootstrap phase and notifies
// white-listed network participants.
func (h *CoordinatorHandler) CompleteBootstrap(ctx context.Context) error {
	event := log.EventBegin(ctx, "Coordinator.CompleteBootstrap")
	defer event.Done()

	if h.networkConfig.NetworkState(ctx) == protectorpb.NetworkState_PROTECTED {
		return nil
	}

	err := h.networkConfig.SetNetworkState(ctx, protectorpb.NetworkState_PROTECTED)
	if err != nil {
		event.SetError(err)
		return err
	}

	h.SendNetworkConfig(ctx)

	// Disconnect from unauthorized nodes.
	for _, c := range h.host.Network().Conns() {
		peerID := c.RemotePeer()
		if !h.networkConfig.IsAllowed(ctx, peerID) {
			err = c.Close()
			if err != nil {
				event.Append(logging.Metadata{
					peerID.Pretty(): err.Error(),
				})
			} else {
				event.Append(logging.Metadata{
					peerID.Pretty(): "disconnected",
				})
			}
		}
	}

	return nil
}

// SendProposal sends a network update proposal to all participants.
// The proposal contains a random challenge and needs to be signed by
// participants to confirm their agreement.
func (h *CoordinatorHandler) SendProposal(ctx context.Context, req *proposal.Request) {
	eventLock := &sync.Mutex{}
	event := log.EventBegin(ctx, "Coordinator.SendProposal")
	defer event.Done()

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

			stream, err := h.host.NewStream(ctx, peerID, PrivateCoordinatedProposePID)
			if err != nil {
				eventLock.Lock()
				event.Append(logging.Metadata{peerID.Pretty(): err.Error()})
				eventLock.Unlock()
				return
			}

			defer func() {
				if err = stream.Close(); err != nil {
					eventLock.Lock()
					event.Append(logging.Metadata{
						fmt.Sprintf("%s-close-err", peerID.Pretty()): err.Error(),
					})
					eventLock.Unlock()
				}
			}()

			enc := protobuf.Multicodec(nil).Encoder(stream)
			err = enc.Encode(updateReq)
			if err != nil {
				eventLock.Lock()
				event.Append(logging.Metadata{peerID.Pretty(): err.Error()})
				eventLock.Unlock()
				return
			}

			eventLock.Lock()
			event.Append(logging.Metadata{peerID.Pretty(): "ok"})
			eventLock.Unlock()
		}(peerID)
	}

	wg.Wait()
}

// SendNetworkConfig sends the current network configuration to all
// white-listed participants. It logs errors but swallows them.
func (h *CoordinatorHandler) SendNetworkConfig(ctx context.Context) {
	eventLock := &sync.Mutex{}
	event := log.EventBegin(ctx, "Coordinator.SendNetworkConfig")
	defer event.Done()

	networkConfig := h.networkConfig.Copy(ctx)
	allowedPeers := h.networkConfig.AllowedPeers(ctx)

	wg := &sync.WaitGroup{}

	for _, peerID := range allowedPeers {
		if peerID == h.host.ID() {
			continue
		}

		wg.Add(1)

		go func(peerID peer.ID) {
			defer wg.Done()

			stream, err := h.host.NewStream(ctx, peerID, PrivateCoordinatedConfigPID)
			if err != nil {
				eventLock.Lock()
				event.Append(logging.Metadata{peerID.Pretty(): err.Error()})
				eventLock.Unlock()
				return
			}

			defer func() {
				if err = stream.Close(); err != nil {
					eventLock.Lock()
					event.Append(logging.Metadata{
						fmt.Sprintf("%s-close-err", peerID.Pretty()): err.Error(),
					})
					eventLock.Unlock()
				}
			}()

			enc := protobuf.Multicodec(nil).Encoder(stream)
			err = enc.Encode(&networkConfig)
			if err != nil {
				eventLock.Lock()
				event.Append(logging.Metadata{peerID.Pretty(): err.Error()})
				eventLock.Unlock()
				return
			}

			eventLock.Lock()
			event.Append(logging.Metadata{peerID.Pretty(): "ok"})
			eventLock.Unlock()
		}(peerID)
	}

	wg.Wait()
}

// Close removes the protocol handlers.
func (h *CoordinatorHandler) Close(ctx context.Context) {
	log.Event(ctx, "Coordinator.Close")
	h.host.RemoveStreamHandler(PrivateCoordinatorHandshakePID)
	h.host.RemoveStreamHandler(PrivateCoordinatorProposePID)
	h.host.RemoveStreamHandler(PrivateCoordinatorVotePID)
}
