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

	"github.com/stratumn/go-indigonode/core/monitoring"
	"github.com/stratumn/go-indigonode/core/protector"

	"gx/ipfs/QmcJukH2sAFjY3HdBKq35WDzWoL3UUu2gt9wdfqZTUyM74/go-libp2p-peer"
	ihost "gx/ipfs/QmfZTdmunzKzAGJrSvXXQbQ5kLLUiEMX5vdwux7iXkdk7D/go-libp2p-host"
)

// Disconnect disconnects from the given peer.
func Disconnect(
	ctx context.Context,
	host ihost.Host,
	peerID peer.ID,
) {
	_, span := monitoring.StartSpan(ctx, "bootstrap", "Disconnect", monitoring.SpanOptionPeerID(peerID))
	defer span.End()

	for _, c := range host.Network().Conns() {
		if c.RemotePeer() == peerID {
			err := c.Close()
			if err != nil {
				span.SetUnknownError(err)
			}
		}
	}
}

// DisconnectUnauthorized disconnects from all unauthorized peers.
func DisconnectUnauthorized(
	ctx context.Context,
	host ihost.Host,
	networkConfig protector.NetworkConfig,
) {
	ctx, span := monitoring.StartSpan(ctx, "bootstrap", "DisconnectUnauthorized")
	defer span.End()

	for _, c := range host.Network().Conns() {
		peerID := c.RemotePeer()
		if !networkConfig.IsAllowed(ctx, peerID) {
			err := c.Close()
			if err != nil {
				span.Annotate(ctx, peerID.Pretty(), err.Error())
			} else {
				span.Annotate(ctx, peerID.Pretty(), "disconnected")
			}
		}
	}
}
