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

	"github.com/stratumn/alice/core/protector"

	logging "gx/ipfs/QmSpJByNKFX1sCsHBEp3R73FL4NF6FnQTEGyNAXHm2GS52/go-log"
	"gx/ipfs/QmcJukH2sAFjY3HdBKq35WDzWoL3UUu2gt9wdfqZTUyM74/go-libp2p-peer"
	ihost "gx/ipfs/QmfZTdmunzKzAGJrSvXXQbQ5kLLUiEMX5vdwux7iXkdk7D/go-libp2p-host"
)

// Disconnect disconnects from the given peer.
func Disconnect(
	host ihost.Host,
	peerID peer.ID,
	event *logging.EventInProgress,
) {
	for _, c := range host.Network().Conns() {
		if c.RemotePeer() == peerID {
			err := c.Close()
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
}

// DisconnectUnauthorized disconnects from all unauthorized peers.
func DisconnectUnauthorized(
	ctx context.Context,
	host ihost.Host,
	networkConfig protector.NetworkConfig,
	event *logging.EventInProgress,
) {
	for _, c := range host.Network().Conns() {
		peerID := c.RemotePeer()
		if !networkConfig.IsAllowed(ctx, peerID) {
			err := c.Close()
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
}
