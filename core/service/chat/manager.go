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

package chat

import (
	"context"
	"os"
	"sync"

	"github.com/pelletier/go-toml"
	"github.com/pkg/errors"

	logging "gx/ipfs/QmSpJByNKFX1sCsHBEp3R73FL4NF6FnQTEGyNAXHm2GS52/go-log"
	peer "gx/ipfs/QmZoWKhxUmZ2seW4BzX6fJkNR8hh9PsGModr7q171yq2SS/go-libp2p-peer"

	pb "github.com/stratumn/alice/grpc/chat"
)

// collection is a slice of records in a TOML file.
type collection struct {
	Records map[string]PeerHistory `toml:"chat"`
}

// PeerHistory represents a list of messages
type PeerHistory []pb.DatedMessage

// Manager manages a chat history.
type Manager struct {
	filename string

	mu      sync.RWMutex
	history map[peer.ID]PeerHistory
}

// NewManager creates a new contact manager.
func NewManager(filename string) (*Manager, error) {
	history, err := load(filename)
	if err != nil {
		if !os.IsNotExist(errors.Cause(err)) {
			return nil, err
		}

		history = map[peer.ID]PeerHistory{}
	}

	return &Manager{
		filename: filename,
		history:  history,
	}, nil
}

// Get returns the chat history with a peer.
func (m *Manager) Get(peerID peer.ID) (PeerHistory, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	chatHistory := m.history[peerID]

	return chatHistory, nil
}

// Add adds a message to the history of a peer.
func (m *Manager) Add(peerID peer.ID, message *pb.DatedMessage) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Clone values in case saving fails.
	history := map[peer.ID]PeerHistory{}
	for k, v := range m.history {
		ch := make(PeerHistory, len(v))
		copy(ch, v)
		history[k] = ch
	}

	history[peerID] = append(history[peerID], *message)

	if err := save(history, m.filename); err != nil {
		return err
	}

	m.history = history

	return nil
}

// load loads chat history from a file.
func load(filename string) (map[peer.ID]PeerHistory, error) {
	mode := os.O_RDONLY
	f, err := os.OpenFile(filename, mode, 0600)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	defer func() {
		if err := f.Close(); err != nil {
			log.Event(context.Background(), "closeError", logging.Metadata{
				"error": err.Error(),
			})
		}
	}()

	col := collection{}
	if err := toml.NewDecoder(f).Decode(&col); err != nil {
		return nil, errors.WithStack(err)
	}

	history := map[peer.ID]PeerHistory{}

	for id, records := range col.Records {
		pid, err := peer.IDB58Decode(id)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		history[pid] = records
	}

	return history, nil
}

// save saves history to a file.
func save(history map[peer.ID]PeerHistory, filename string) error {
	mode := os.O_WRONLY | os.O_TRUNC | os.O_CREATE
	f, err := os.OpenFile(filename, mode, 0600)
	if err != nil {
		return errors.WithStack(err)
	}

	defer func() {
		if err := f.Close(); err != nil {
			log.Event(context.Background(), "closeError", logging.Metadata{
				"error": err.Error(),
			})
		}
	}()

	recordsMap := make(map[string]PeerHistory)

	for peerID, peerHistory := range history {
		recordsMap[peerID.Pretty()] = peerHistory
	}

	return errors.WithStack(toml.NewEncoder(f).Encode(collection{recordsMap}))
}
