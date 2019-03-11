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

package protector

import (
	"github.com/pkg/errors"

	"github.com/libp2p/go-libp2p-peer"
	"github.com/multiformats/go-multiaddr"
)

// Supported network protection modes.
const (
	// PrivateWithCoordinatorMode uses a coordinator node
	// for network participants updates.
	PrivateWithCoordinatorMode = "private-with-coordinator"
)

// Errors encountered when creating a NetworkMode.
var (
	ErrInvalidCoordinatorID   = errors.New("invalid coordinator ID")
	ErrMissingCoordinatorAddr = errors.New("missing coordinator address")
	ErrInvalidCoordinatorAddr = errors.New("invalid coordinator address")
)

// NetworkMode describes the mode of operation of the network.
// It contains all the configured values necessary for operating
// the network correctly.
// It is used to configure the network protocols used by nodes.
type NetworkMode struct {
	ProtectionMode   string
	IsCoordinator    bool
	CoordinatorID    peer.ID
	CoordinatorAddrs []multiaddr.Multiaddr
}

// NewCoordinatorNetworkMode returns the NetworkMode for a network coordinator.
func NewCoordinatorNetworkMode() *NetworkMode {
	return &NetworkMode{
		ProtectionMode: PrivateWithCoordinatorMode,
		IsCoordinator:  true,
	}
}

// NewCoordinatedNetworkMode returns the NetworkMode for a network
// that uses a coordinator.
func NewCoordinatedNetworkMode(coordinatorID string, coordinatorAddrs []string) (*NetworkMode, error) {
	coordinatorPeerID, err := peer.IDB58Decode(coordinatorID)
	if err != nil {
		return nil, ErrInvalidCoordinatorID
	}

	if len(coordinatorAddrs) == 0 {
		return nil, ErrMissingCoordinatorAddr
	}

	var coordinatorMultiaddrs []multiaddr.Multiaddr
	for _, addr := range coordinatorAddrs {
		coordinatorAddr, err := multiaddr.NewMultiaddr(addr)
		if err != nil {
			return nil, ErrInvalidCoordinatorAddr
		}

		coordinatorMultiaddrs = append(coordinatorMultiaddrs, coordinatorAddr)
	}

	return &NetworkMode{
		ProtectionMode:   PrivateWithCoordinatorMode,
		IsCoordinator:    false,
		CoordinatorID:    coordinatorPeerID,
		CoordinatorAddrs: coordinatorMultiaddrs,
	}, nil
}
