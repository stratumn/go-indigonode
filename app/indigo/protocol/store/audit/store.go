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

package audit

import (
	"context"

	"github.com/stratumn/go-indigocore/cs"
	// Blank import to register fossilizer concrete evidence types.
	_ "github.com/stratumn/go-indigocore/fossilizer/evidences"

	peer "gx/ipfs/QmcJukH2sAFjY3HdBKq35WDzWoL3UUu2gt9wdfqZTUyM74/go-libp2p-peer"
)

const (
	// DefaultLimit is the default pagination limit.
	DefaultLimit = 20
)

// SegmentFilter contains filtering options for segments.
type SegmentFilter struct {
	Pagination

	PeerID peer.ID
}

// Pagination defines pagination options.
type Pagination struct {
	Top  uint
	Skip uint
}

// NewDefaultPagination returns the default pagination.
func NewDefaultPagination() Pagination {
	return Pagination{
		Top:  DefaultLimit,
		Skip: 0,
	}
}

// InitIfInvalid sets pagination default values if invalid
// (for example requesting 0 results).
func (p *Pagination) InitIfInvalid() {
	if p.Top == 0 {
		p.Top = DefaultLimit
	}
}

// Reader defines operations to read from the store.
type Reader interface {
	GetByPeer(context.Context, peer.ID, Pagination) (cs.SegmentSlice, error)
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
