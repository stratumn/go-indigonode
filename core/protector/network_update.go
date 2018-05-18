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

import "gx/ipfs/QmcJukH2sAFjY3HdBKq35WDzWoL3UUu2gt9wdfqZTUyM74/go-libp2p-peer"

// NetworkUpdateType defines the operations supported by a private network.
type NetworkUpdateType int

// Operations supported by a private network.
const (
	// Add a peer.
	Add NetworkUpdateType = 1
	// Remove a peer
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
