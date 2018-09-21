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

package service

import (
	"encoding/binary"
	peer "gx/ipfs/QmQsErDt8Qgw1XrsXf2BpEzDgGWtB1YLsTAARBup5b6B9W/go-libp2p-peer"
	"time"

	"github.com/pkg/errors"
	pb "github.com/stratumn/go-indigonode/app/chat/grpc"
	"github.com/stratumn/go-indigonode/core/db"
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
