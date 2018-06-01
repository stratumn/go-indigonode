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
	pb "github.com/stratumn/alice/pb/bootstrap"
	protectorpb "github.com/stratumn/alice/pb/protector"

	protobuf "gx/ipfs/QmRDePEiL4Yupq5EkcK3L3ko3iMgYaqUdLu7xc1kqs7dnV/go-multicodec/protobuf"
	logging "gx/ipfs/QmSpJByNKFX1sCsHBEp3R73FL4NF6FnQTEGyNAXHm2GS52/go-log"
	inet "gx/ipfs/QmXoz9o2PT3tEzf7hicegwex5UgVP54n3k82K7jrWFyN86/go-libp2p-net"
	"gx/ipfs/QmZNkThpqfVXs9GNbexPrfBbXSLNYeKrE7jwFM2oqHbyqN/go-libp2p-protocol"
	"gx/ipfs/QmcJukH2sAFjY3HdBKq35WDzWoL3UUu2gt9wdfqZTUyM74/go-libp2p-peer"
	ihost "gx/ipfs/QmfZTdmunzKzAGJrSvXXQbQ5kLLUiEMX5vdwux7iXkdk7D/go-libp2p-host"
)

var (
	// PrivateCoordinatedProtocolID is the protocol for receiving network
	// updates in a private network. Network participants should implement this protocol.
	PrivateCoordinatedProtocolID = protocol.ID("/alice/indigo/bootstrap/private/coordinated/v1.0.0")
)

// CoordinatedHandler is the handler for a non-coordinator node
// in a private network that has a coordinator.
type CoordinatedHandler struct {
	coordinatorID peer.ID
	host          ihost.Host
	networkConfig protector.NetworkConfig
}

// NewCoordinatedHandler returns a Handler for a non-coordinator node.
func NewCoordinatedHandler(
	ctx context.Context,
	host ihost.Host,
	networkMode *protector.NetworkMode,
	networkConfig protector.NetworkConfig,
) (Handler, error) {
	event := log.EventBegin(ctx, "Coordinated.New", logging.Metadata{
		"coordinatorID": networkMode.CoordinatorID.Pretty(),
	})
	defer event.Done()

	handler := CoordinatedHandler{
		coordinatorID: networkMode.CoordinatorID,
		host:          host,
		networkConfig: networkConfig,
	}

	err := handler.handshake(ctx)
	if err != nil {
		return nil, err
	}

	host.SetStreamHandler(PrivateCoordinatedProtocolID, handler.Handle)

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

	stream, err := h.host.NewStream(ctx, h.coordinatorID, PrivateCoordinatorProtocolID)
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

	if len(networkConfig.Participants) == 0 {
		event.SetError(protector.ErrConnectionRefused)
		return protector.ErrConnectionRefused
	}

	if !networkConfig.ValidateSignature(ctx, h.coordinatorID) {
		event.SetError(protectorpb.ErrInvalidSignature)
		return protectorpb.ErrInvalidSignature
	}

	return h.networkConfig.Reset(ctx, &networkConfig)
}

// Handle handles an incoming stream.
func (h *CoordinatedHandler) Handle(stream inet.Stream) {
	ctx := context.Background()
	defer log.EventBegin(ctx, "Coordinated.Handle").Done()
}

// Close removes the protocol handlers.
func (h *CoordinatedHandler) Close(ctx context.Context) {
	log.Event(ctx, "Coordinated.Close")
	h.host.RemoveStreamHandler(PrivateCoordinatedProtocolID)
}
