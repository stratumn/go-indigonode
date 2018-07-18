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
