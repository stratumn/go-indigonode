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

// Package protocol implements the network bootstrapping protocols.
// It contains the node-to-node communication layer to administer
// private networks.
package protocol

import (
	"context"

	"github.com/pkg/errors"
	"github.com/stratumn/go-indigonode/core/app/bootstrap/protocol/proposal"
	"github.com/stratumn/go-indigonode/core/protector"
	"github.com/stratumn/go-indigonode/core/streamutil"

	"gx/ipfs/QmWWQ2Txc2c6tqjsBpzg5Ar652cHPGNsQQp2SejkNmkUMb/go-multiaddr"
	"gx/ipfs/QmcJukH2sAFjY3HdBKq35WDzWoL3UUu2gt9wdfqZTUyM74/go-libp2p-peer"
	ihost "gx/ipfs/QmfZTdmunzKzAGJrSvXXQbQ5kLLUiEMX5vdwux7iXkdk7D/go-libp2p-host"
)

// Errors returned by bootstrap.
var (
	ErrInvalidProtectionMode = errors.New("invalid protection mode")
	ErrInvalidOperation      = errors.New("invalid operation")
)

// Handler defines the methods to bootstrap and administer a network.
type Handler interface {
	// Handshake performs a network handshake.
	// This should be done when the node starts.
	Handshake(context.Context) error

	// AddNode adds a node to the network. Depending on the underlying
	// protocol, adding the node might require other node's approval
	// or even be rejected.
	AddNode(context.Context, peer.ID, multiaddr.Multiaddr, []byte) error

	// RemoveNode removes a node from the network.
	// Depending on the underlying protocol, removing the node might require
	// other node's approval or even be rejected.
	RemoveNode(context.Context, peer.ID) error

	// Accept accepts a proposal to add or remove a node (identified
	// by its PeerID).
	Accept(context.Context, peer.ID) error

	// Reject rejects a proposal to add or remove a node (identified
	// by its PeerID).
	Reject(context.Context, peer.ID) error

	// CompleteBootstrap completes the network bootstrapping phase.
	CompleteBootstrap(context.Context) error

	// Close closes all resources used by the protocol handler.
	Close(context.Context)
}

// New creates the right instance of the Handler interface
// depending on the network parameters.
// It will register protocols to handle network requests.
func New(
	host ihost.Host,
	streamProvider streamutil.Provider,
	networkMode *protector.NetworkMode,
	networkConfig protector.NetworkConfig,
	store proposal.Store,
) (Handler, error) {
	if networkMode == nil {
		return &PublicNetworkHandler{}, nil
	}

	switch networkMode.ProtectionMode {
	case "":
		return &PublicNetworkHandler{}, nil
	case protector.PrivateWithCoordinatorMode:
		if networkMode.IsCoordinator {
			h := NewCoordinatorHandler(host, streamProvider, networkConfig, store)
			return h, nil
		}

		h := NewCoordinatedHandler(host, streamProvider, networkMode, networkConfig, store)
		return h, nil
	default:
		return nil, ErrInvalidProtectionMode
	}
}
