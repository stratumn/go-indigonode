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

//go:generate mockgen -package mockaudit -destination mockaudit/mockaudit.go github.com/stratumn/alice/core/protocol/indigo/store/audit Store

package audit

import (
	"context"

	pb "github.com/stratumn/alice/pb/indigo/store"

	peer "gx/ipfs/QmZoWKhxUmZ2seW4BzX6fJkNR8hh9PsGModr7q171yq2SS/go-libp2p-peer"
)

// Pagination defines pagination options.
type Pagination struct {
	Top  uint
	Skip uint
}

// Reader defines operations to read from the store.
type Reader interface {
	GetByPeer(context.Context, peer.ID, Pagination) ([]pb.SignedLink, error)
}

// Writer defines operations to add to the store.
type Writer interface {
	AddLink(context.Context, *pb.SignedLink) error
}

// Store defines an audit store. This is where we store invalid
// links received from peers that need auditing.
// Since these links are signed by the sender, it is non-repudiable.
// The sender will have to explain why it produced invalid links.
type Store interface {
	Reader
	Writer
}
