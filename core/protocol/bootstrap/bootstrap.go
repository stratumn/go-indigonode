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

// Package bootstrap implements the network bootstrapping protocols.
// It contains the node-to-node communication layer to administer
// private networks.
package bootstrap

import (
	"context"

	"github.com/pkg/errors"
	"github.com/stratumn/alice/core/protector"
	"github.com/stratumn/alice/core/protocol/bootstrap/proposal"

	logging "gx/ipfs/QmSpJByNKFX1sCsHBEp3R73FL4NF6FnQTEGyNAXHm2GS52/go-log"
	"gx/ipfs/QmWWQ2Txc2c6tqjsBpzg5Ar652cHPGNsQQp2SejkNmkUMb/go-multiaddr"
	"gx/ipfs/QmcJukH2sAFjY3HdBKq35WDzWoL3UUu2gt9wdfqZTUyM74/go-libp2p-peer"
	ihost "gx/ipfs/QmfZTdmunzKzAGJrSvXXQbQ5kLLUiEMX5vdwux7iXkdk7D/go-libp2p-host"
)

var log = logging.Logger("bootstrap")

// Errors returned by bootstrap.
var (
	ErrInvalidProtectionMode = errors.New("invalid protection mode")
	ErrInvalidOperation      = errors.New("invalid operation")
)

// Handler defines the methods to bootstrap and administer a network.
type Handler interface {
	// AddNode adds a node to the network. Depending on the underlying
	// protocol, adding the node might require other node's approval
	// or even be rejected.
	AddNode(context.Context, peer.ID, multiaddr.Multiaddr, []byte) error

	// Accept accepts a proposal to add or remove a node (identified
	// by its PeerID).
	Accept(context.Context, peer.ID) error

	// Reject rejects a proposal to add or remove a node (identified
	// by its PeerID).
	Reject(context.Context, peer.ID) error

	// Close closes all resources used by the protocol handler.
	Close(context.Context)
}

// New creates the right instance of the Handler interface
// depending on the network parameters.
// It will register protocols to handle network requests.
func New(
	ctx context.Context,
	host ihost.Host,
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
			return NewCoordinatorHandler(host, networkConfig, store)
		}

		return NewCoordinatedHandler(ctx, host, networkMode, networkConfig)
	default:
		return nil, ErrInvalidProtectionMode
	}
}
