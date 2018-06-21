// Copyright © 2017-2018  Stratumn SAS
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

package service

import (
	"encoding/binary"
	peer "gx/ipfs/QmcJukH2sAFjY3HdBKq35WDzWoL3UUu2gt9wdfqZTUyM74/go-libp2p-peer"
	"time"

	"github.com/pkg/errors"
	pb "github.com/stratumn/alice/app/chat/grpc"
	"github.com/stratumn/alice/core/db"
)

// PeerHistory represents a list of messages
type PeerHistory []pb.DatedMessage

// Manager manages a chat history.
type Manager struct {
	hdb db.DB
}

// NewManager creates a new chat manager.
func NewManager(dbPath string) (*Manager, error) {
	hdb, err := db.NewFileDB(dbPath, nil)
	if err != nil {
		return nil, err
	}

	return &Manager{hdb: hdb}, nil
}

// Get returns the chat history with a peer.
func (m *Manager) Get(peerID peer.ID) (PeerHistory, error) {
	var result PeerHistory

	iter := m.hdb.IteratePrefix([]byte(peerID))

	for {
		next, err := iter.Next()
		if err != nil {
			return nil, err
		}
		if !next {
			break
		}
		msg := pb.DatedMessage{}
		err = msg.Unmarshal(iter.Value())
		if err != nil {
			return nil, err
		}

		result = append(result, msg)
	}

	return result, nil
}

// Add adds a message to the history of a peer.
func (m *Manager) Add(peerID peer.ID, message *pb.DatedMessage) error {
	buf, err := message.Marshal()
	if err != nil {
		return errors.WithStack(err)
	}

	return m.hdb.Put(msgKey(peerID, message), buf)
}

// msgKey builds the key for a message. Adding the time to the key allows us
// to prevent duplicated keys and to retrieve messages in the right order.
func msgKey(peerID peer.ID, message *pb.DatedMessage) []byte {
	t := time.Unix(message.Time.Seconds, int64(message.Time.Nanos))
	return append([]byte(peerID), encodeUint64(uint64(t.UnixNano()))...)
}

// encodeUint64 encodes an uint64 to a BigEndian buffer.
func encodeUint64(value uint64) []byte {
	v := make([]byte, 8)
	binary.BigEndian.PutUint64(v, value)

	return v
}
