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

import "gx/ipfs/QmQsErDt8Qgw1XrsXf2BpEzDgGWtB1YLsTAARBup5b6B9W/go-libp2p-peer"

// NetworkUpdateType defines the operations supported by a private network.
type NetworkUpdateType byte

// Operations supported by a private network.
const (
	// Add a peer.
	Add NetworkUpdateType = 1
	// Remove a peer.
	Remove NetworkUpdateType = 2
)

// NetworkUpdate describes a change in a private network.
type NetworkUpdate struct {
	Type   NetworkUpdateType
	PeerID peer.ID
}

// CreateAddNetworkUpdate creates an update to add a peer to the network.
func CreateAddNetworkUpdate(peerID peer.ID) NetworkUpdate {
	return NetworkUpdate{
		Type:   Add,
		PeerID: peerID,
	}
}

// CreateRemoveNetworkUpdate creates an update to remove a peer from the network.
func CreateRemoveNetworkUpdate(peerID peer.ID) NetworkUpdate {
	return NetworkUpdate{
		Type:   Remove,
		PeerID: peerID,
	}
}
