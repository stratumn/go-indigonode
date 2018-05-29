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

package protector

import (
	"github.com/pkg/errors"

	"gx/ipfs/QmWWQ2Txc2c6tqjsBpzg5Ar652cHPGNsQQp2SejkNmkUMb/go-multiaddr"
	"gx/ipfs/QmcJukH2sAFjY3HdBKq35WDzWoL3UUu2gt9wdfqZTUyM74/go-libp2p-peer"
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
