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

package store

import (
	"context"

	"github.com/pkg/errors"
	"github.com/stratumn/go-indigocore/cs"

	peer "gx/ipfs/QmQsErDt8Qgw1XrsXf2BpEzDgGWtB1YLsTAARBup5b6B9W/go-libp2p-peer"
	ihost "gx/ipfs/QmeMYW7Nj8jnnEfs9qhm7SxKkoDPUWXu3MsxX6BFwz34tf/go-libp2p-host"
)

// Host represents an Indigo Node host.
type Host = ihost.Host

// Errors used by the NetworkManager.
var (
	ErrNetworkNotReady = errors.New("network not ready to receive messages")
)

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
