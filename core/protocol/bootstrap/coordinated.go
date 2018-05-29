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
	ihost "gx/ipfs/QmfZTdmunzKzAGJrSvXXQbQ5kLLUiEMX5vdwux7iXkdk7D/go-libp2p-host"
)

// CoordinatedHandler is the handler for a non-coordinator node
// in a private network that has a coordinator.
type CoordinatedHandler struct {
	host          ihost.Host
	networkConfig protector.NetworkConfig
}

// NewCoordinatedHandler returns a Handler for a non-coordinator node.
func NewCoordinatedHandler(
	host ihost.Host,
	networkMode *protector.NetworkMode,
	networkConfig protector.NetworkConfig,
) (Handler, error) {
	handler := CoordinatedHandler{host: host, networkConfig: networkConfig}

	host.SetStreamHandler(PrivateWithCoordinatorProtocolID, handler.handle)

	return &handler, nil
}

func (h *CoordinatedHandler) handle(stream inet.Stream) {
	ctx := context.Background()
	defer log.EventBegin(ctx, "Coordinated.Handle").Done()
}

// Close removes the protocol handlers.
func (h *CoordinatedHandler) Close(ctx context.Context) {
	log.Event(ctx, "Coordinated.Close")
	h.host.RemoveStreamHandler(PrivateWithCoordinatorProtocolID)
}
