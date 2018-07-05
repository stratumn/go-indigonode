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

package store

import (
	"context"

	"github.com/stratumn/go-indigocore/cs"

	peer "gx/ipfs/QmcJukH2sAFjY3HdBKq35WDzWoL3UUu2gt9wdfqZTUyM74/go-libp2p-peer"
	ihost "gx/ipfs/QmfZTdmunzKzAGJrSvXXQbQ5kLLUiEMX5vdwux7iXkdk7D/go-libp2p-host"
)

// Host represents an Indigo Node host.
type Host = ihost.Host

// NetworkManager provides methods to manage and join PoP networks.
type NetworkManager interface {
	// NodeID returns the ID of the node in the network.
	NodeID() peer.ID

	// Join joins a PoP network.
	Join(ctx context.Context, networkID string, host Host) error
	// Leave leaves a PoP network.
	Leave(ctx context.Context, networkID string) error

	// Publish sends a message to all the network.
	Publish(ctx context.Context, link *cs.Link) error
	// Listen to messages from the network. Cancel the context
	// to stop listening.
	Listen(ctx context.Context) error

	// AddListener adds a listeners for incoming segments.
	AddListener() <-chan *cs.Segment
	// RemoveListener removes a listener.
	RemoveListener(<-chan *cs.Segment)
}
