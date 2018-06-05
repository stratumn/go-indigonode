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

	"github.com/stratumn/alice/core/protector"
	pb "github.com/stratumn/alice/pb/bootstrap"
	protectorpb "github.com/stratumn/alice/pb/protector"

	protobuf "gx/ipfs/QmRDePEiL4Yupq5EkcK3L3ko3iMgYaqUdLu7xc1kqs7dnV/go-multicodec/protobuf"
	logging "gx/ipfs/QmSpJByNKFX1sCsHBEp3R73FL4NF6FnQTEGyNAXHm2GS52/go-log"
	inet "gx/ipfs/QmXoz9o2PT3tEzf7hicegwex5UgVP54n3k82K7jrWFyN86/go-libp2p-net"
	"gx/ipfs/QmZNkThpqfVXs9GNbexPrfBbXSLNYeKrE7jwFM2oqHbyqN/go-libp2p-protocol"
	ihost "gx/ipfs/QmfZTdmunzKzAGJrSvXXQbQ5kLLUiEMX5vdwux7iXkdk7D/go-libp2p-host"
)

var (
	// PrivateCoordinatorProtocolID is the protocol for managing a private
	// network. Only the network coordinator should implement this protocol.
	PrivateCoordinatorProtocolID = protocol.ID("/alice/indigo/bootstrap/private/coordinator/v1.0.0")
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

// Close removes the protocol handlers.
func (h *CoordinatorHandler) Close(ctx context.Context) {
	log.Event(ctx, "Coordinator.Close")
	h.host.RemoveStreamHandler(PrivateCoordinatorProtocolID)
}
