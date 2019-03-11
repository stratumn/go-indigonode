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

	"github.com/libp2p/go-libp2p-peer"
	"github.com/multiformats/go-multiaddr"
)

// PublicNetworkHandler is the handler for public networks.
// Public networks are completely open, so this handler doesn't do anything.
type PublicNetworkHandler struct{}

// Handshake can't be used in a public network.
// You can directly connect to any node freely.
func (h *PublicNetworkHandler) Handshake(context.Context) error {
	return nil
}

// AddNode can't be used in a public network.
// You can directly connect to any node freely.
func (h *PublicNetworkHandler) AddNode(context.Context, peer.ID, multiaddr.Multiaddr, []byte) error {
	return ErrInvalidOperation
}

// RemoveNode can't be used in a public network.
// You can directly connect to any node freely.
func (h *PublicNetworkHandler) RemoveNode(context.Context, peer.ID) error {
	return ErrInvalidOperation
}

// Accept can't be used in a public network.
// You can directly connect to any node freely.
func (h *PublicNetworkHandler) Accept(context.Context, peer.ID) error {
	return ErrInvalidOperation
}

// Reject can't be used in a public network.
// You can directly connect to any node freely.
func (h *PublicNetworkHandler) Reject(context.Context, peer.ID) error {
	return ErrInvalidOperation
}

// CompleteBootstrap can't be used in a public network.
// There is no bootstrapping phase in such networks.
func (h *PublicNetworkHandler) CompleteBootstrap(context.Context) error {
	return ErrInvalidOperation
}

// Close doesn't do anything.
func (h *PublicNetworkHandler) Close(context.Context) {}
