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
	pb "github.com/stratumn/alice/pb/bootstrap"
	protectorpb "github.com/stratumn/alice/pb/protector"

	protobuf "gx/ipfs/QmRDePEiL4Yupq5EkcK3L3ko3iMgYaqUdLu7xc1kqs7dnV/go-multicodec/protobuf"
	logging "gx/ipfs/QmSpJByNKFX1sCsHBEp3R73FL4NF6FnQTEGyNAXHm2GS52/go-log"
	"gx/ipfs/QmWWQ2Txc2c6tqjsBpzg5Ar652cHPGNsQQp2SejkNmkUMb/go-multiaddr"
	inet "gx/ipfs/QmXoz9o2PT3tEzf7hicegwex5UgVP54n3k82K7jrWFyN86/go-libp2p-net"
	"gx/ipfs/QmZNkThpqfVXs9GNbexPrfBbXSLNYeKrE7jwFM2oqHbyqN/go-libp2p-protocol"
	"gx/ipfs/QmcJukH2sAFjY3HdBKq35WDzWoL3UUu2gt9wdfqZTUyM74/go-libp2p-peer"
	"gx/ipfs/QmdeiKhUy1TVGBaKxt7y1QmBDLBdisSrLJ1x58Eoj4PXUh/go-libp2p-peerstore"
	ihost "gx/ipfs/QmfZTdmunzKzAGJrSvXXQbQ5kLLUiEMX5vdwux7iXkdk7D/go-libp2p-host"
)

var (
	// PrivateCoordinatorProtocolID is the protocol for managing a private
	// network. Only the network coordinator should implement this protocol.
	PrivateCoordinatorProtocolID = protocol.ID("/alice/indigo/bootstrap/private/coordinator/v1.0.0")
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
}

// NewCoordinatorHandler returns a Handler for a coordinator node.
func NewCoordinatorHandler(
	host ihost.Host,
	networkConfig protector.NetworkConfig,
) (Handler, error) {
	handler := CoordinatorHandler{host: host, networkConfig: networkConfig}

	host.SetStreamHandler(PrivateCoordinatorProtocolID, handler.Handle)

	return &handler, nil
}

// Handle handles an incoming stream.
func (h *CoordinatorHandler) Handle(stream inet.Stream) {
	var protocolErr error
	ctx := context.Background()
	event := log.EventBegin(ctx, "Coordinator.Handle", logging.Metadata{
		"remote": stream.Conn().RemotePeer().Pretty(),
	})
	defer func() {
		if protocolErr != nil {
			event.SetError(protocolErr)
		}

		if err := stream.Close(); err != nil {
			event.Append(logging.Metadata{"close_err": err.Error()})
		}

		event.Done()
	}()

	networkState := h.networkConfig.NetworkState(ctx)
	allowed := h.networkConfig.IsAllowed(ctx, stream.Conn().RemotePeer())

	// Once the bootstrap is complete, we reject non-white-listed peers.
	if !allowed && networkState == protectorpb.NetworkState_PROTECTED {
		protocolErr = protector.ErrConnectionRefused
		if err := stream.Reset(); err != nil {
			event.Append(logging.Metadata{"reset_err": err.Error()})
		}

		return
	}

	dec := protobuf.Multicodec(nil).Decoder(stream)
	var hello pb.Hello
	if err := dec.Decode(&hello); err != nil {
		protocolErr = protector.ErrConnectionRefused
		if err := stream.Reset(); err != nil {
			event.Append(logging.Metadata{"reset_err": err.Error()})
		}

		return
	}

	enc := protobuf.Multicodec(nil).Encoder(stream)

	// We should not reveal network participants to unwanted peers.
	if !allowed {
		protocolErr = enc.Encode(&protectorpb.NetworkConfig{})
		return
	}

	networkConfig := h.networkConfig.Copy(ctx)
	protocolErr = enc.Encode(&networkConfig)
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

// Accept accepts a proposal to add or remove a node
// and notifies network participants.
func (h *CoordinatorHandler) Accept(ctx context.Context, peerID peer.ID) error {
	event := log.EventBegin(ctx, "Coordinator.Accept", peerID)
	defer event.Done()

	return nil
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
		wg.Add(1)

		go func(peerID peer.ID) {
			defer wg.Done()

			stream, err := h.host.NewStream(ctx, peerID, PrivateCoordinatedProtocolID)
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
	h.host.RemoveStreamHandler(PrivateCoordinatorProtocolID)
}
