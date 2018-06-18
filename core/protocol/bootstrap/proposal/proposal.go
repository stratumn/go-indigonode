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

// Package proposal implements a store for network update proposals.
package proposal

import (
	"context"

	logging "gx/ipfs/QmSpJByNKFX1sCsHBEp3R73FL4NF6FnQTEGyNAXHm2GS52/go-log"
	"gx/ipfs/QmcJukH2sAFjY3HdBKq35WDzWoL3UUu2gt9wdfqZTUyM74/go-libp2p-peer"
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
