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

	inet "gx/ipfs/QmXoz9o2PT3tEzf7hicegwex5UgVP54n3k82K7jrWFyN86/go-libp2p-net"
	"gx/ipfs/QmZNkThpqfVXs9GNbexPrfBbXSLNYeKrE7jwFM2oqHbyqN/go-libp2p-protocol"
	ihost "gx/ipfs/QmfZTdmunzKzAGJrSvXXQbQ5kLLUiEMX5vdwux7iXkdk7D/go-libp2p-host"
)

var (
	// PrivateWithCoordinatorProtocolID is the protocol for managing a private
	// network when a special node acts as coordinator.
	PrivateWithCoordinatorProtocolID = protocol.ID("/alice/indigo/bootstrap/private/coordinator/v1.0.0")
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
	networkMode *protector.NetworkMode,
	networkConfig protector.NetworkConfig,
) (Handler, error) {
	handler := CoordinatorHandler{host: host, networkConfig: networkConfig}

	host.SetStreamHandler(PrivateWithCoordinatorProtocolID, handler.Handle)

	return &handler, nil
}

// Handle handles an incoming stream.
func (h *CoordinatorHandler) Handle(stream inet.Stream) {
	ctx := context.Background()
	defer log.EventBegin(ctx, "Coordinator.Handle").Done()
}

// Close removes the protocol handlers.
func (h *CoordinatorHandler) Close(ctx context.Context) {
	log.Event(ctx, "Coordinator.Close")
	h.host.RemoveStreamHandler(PrivateWithCoordinatorProtocolID)
}
