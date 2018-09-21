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

package grpc

import (
	"time"

	peer "gx/ipfs/QmQsErDt8Qgw1XrsXf2BpEzDgGWtB1YLsTAARBup5b6B9W/go-libp2p-peer"

	"github.com/gogo/protobuf/types"
)

// NewDatedMessageReceived creates a new DatedMessage with a From field
func NewDatedMessageReceived(from peer.ID, content string) *DatedMessage {
	t := time.Now()
	return &DatedMessage{
		From:    []byte(from),
		Content: content,
		Time: &types.Timestamp{
			Seconds: t.Unix(),
			Nanos:   int32(t.Nanosecond()),
		},
	}
}

// NewDatedMessageSent creates a new DatedMessage with a to field
func NewDatedMessageSent(to peer.ID, content string) *DatedMessage {
	t := time.Now()
	return &DatedMessage{
		To:      []byte(to),
		Content: content,
		Time: &types.Timestamp{
			Seconds: t.Unix(),
			Nanos:   int32(t.Nanosecond()),
		},
	}
}

// FromPeer returns the sender peerID as go-libp2p-peer.ID
func (d *DatedMessage) FromPeer() (peer.ID, error) {
	return peer.IDFromBytes(d.From)
}
