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

// Package protector contains implementations of the
// github.com/libp2p/go-libp2p-interface-pnet/ipnet.Protector interface.
//
// Use these implementations in the swarm service to protect a private network.
package protector

import (
	"context"

	"github.com/pkg/errors"

	"gx/ipfs/QmQsErDt8Qgw1XrsXf2BpEzDgGWtB1YLsTAARBup5b6B9W/go-libp2p-peer"
	"gx/ipfs/QmW7Ump7YyBMr712Ta3iEVh3ZYcfVvJaPryfbCnyE826b4/go-libp2p-interface-pnet"
	"gx/ipfs/QmYmsdtJ3HsodkePE3eU3TsCaP2YvPZJ4LoXnNkDE5Tpt7/go-multiaddr"
)

var (
	// ErrConnectionRefused is returned when a connection is refused.
	ErrConnectionRefused = errors.New("connection refused")
)

// Protector protects a network against non-whitelisted peers.
type Protector interface {
	ipnet.Protector

	// ListenForUpdates listens for network updates.
	// This is a blocking call that should be made in a dedicated go routine.
	// Closing the channel will stop the listener.
	ListenForUpdates(<-chan NetworkUpdate)

	// AllowedAddrs returns the list of whitelisted addresses.
	AllowedAddrs(context.Context) []multiaddr.Multiaddr

	// AllowedPeers returns the list of whitelisted peers.
	AllowedPeers(context.Context) []peer.ID
}

// StateAwareProtector protects a network depending on its state.
type StateAwareProtector interface {
	Protector
	NetworkStateWriter
}
