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

package grpc

import (
	"time"

	peer "gx/ipfs/QmcJukH2sAFjY3HdBKq35WDzWoL3UUu2gt9wdfqZTUyM74/go-libp2p-peer"

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
