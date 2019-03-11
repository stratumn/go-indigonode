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

// Package proposal implements a store for network update proposals.
package proposal

import (
	"context"

	"github.com/libp2p/go-libp2p-peer"
	logging "github.com/ipfs/go-log"
)

var log = logging.Logger("bootstrap.proposal")

// Store is used to store pending network updates
// until they have been approved.
type Store interface {
	// AddRequest adds a pending request.
	// If overwrites a previous request for that PeerID if there is one.
	AddRequest(context.Context, *Request) error

	// AddVote adds a vote to a request.
	AddVote(context.Context, *Vote) error

	// Remove removes a request (and partial votes).
	Remove(context.Context, peer.ID) error

	// Get a request for a given PeerID.
	Get(context.Context, peer.ID) (*Request, error)

	// GetVotes gets the votes for a given PeerID.
	GetVotes(context.Context, peer.ID) ([]*Vote, error)

	// List all the pending requests.
	List(context.Context) ([]*Request, error)
}
