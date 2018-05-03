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

	"github.com/stratumn/go-indigocore/cs"
	// Blank import to register fossilizer concrete evidence types.
	_ "github.com/stratumn/go-indigocore/fossilizer/evidences"

	logging "gx/ipfs/QmSpJByNKFX1sCsHBEp3R73FL4NF6FnQTEGyNAXHm2GS52/go-log"
	peer "gx/ipfs/QmZoWKhxUmZ2seW4BzX6fJkNR8hh9PsGModr7q171yq2SS/go-libp2p-peer"
)

var (
	log = logging.Logger("indigo.store.audit")
)

const (
	// DefaultLimit is the default pagination limit.
	DefaultLimit = 20
)

// SegmentFilter contains filtering options for segments.
type SegmentFilter struct {
	Pagination

	PeerID *peer.ID
}

// Pagination defines pagination options.
type Pagination struct {
	Top  uint
	Skip uint
}

// Reader defines operations to read from the store.
type Reader interface {
	GetByPeer(context.Context, peer.ID, *Pagination) (cs.SegmentSlice, error)
}

// Writer defines operations to add to the store.
type Writer interface {
	AddSegment(context.Context, *cs.Segment) error
}

// Store defines an audit store. This is where we store invalid
// segments received from peers that need auditing.
// Since these segments are signed by the sender, it is non-repudiable.
// The sender will have to explain why it produced invalid segments.
type Store interface {
	Reader
	Writer
}
