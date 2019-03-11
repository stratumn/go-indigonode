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
	"context"
	"os"
	"sync"

	"github.com/pelletier/go-toml"
	"github.com/pkg/errors"

	peer "github.com/libp2p/go-libp2p-peer"
	logging "github.com/ipfs/go-log"
)

var (
	// ErrNotFound is returned when a contact is not found.
	ErrNotFound = errors.New("contact not found")
)

// Contact represents a contact.
type Contact struct {
	PeerID peer.ID
}

// record is the format of a contact stored in a TOML file.
//
// In the TOML file, records are stored as an array rather than a map because
// it is easier to read, so we have to add the name to each entry.
type record struct {
	Name   string `toml:"name" comment:"The name of the contact."`
	PeerID string `toml:"peer_id" comment:"The peer ID of the contact."`
}

// collection is a slice of records in a TOML file.
type collection struct {
	Records []record `toml:"contact"`
}

// Manager manages a contact list.
type Manager struct {
	filename string

	mu       sync.RWMutex
	contacts map[string]Contact
}

// NewManager creates a new contact manager.
func NewManager(filename string) (*Manager, error) {
	contacts, err := load(filename)
	if err != nil {
		if !os.IsNotExist(errors.Cause(err)) {
			return nil, err
		}

		contacts = map[string]Contact{}
	}

	return &Manager{
		filename: filename,
		contacts: contacts,
	}, nil
}

// List returns the list of contacts.
func (m *Manager) List() map[string]Contact {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Clone values.
	contacts := map[string]Contact{}
	for k, v := range m.contacts {
		contacts[k] = v
	}

	return contacts
}

// Get finds a contact by name.
//
// It returns a copy of the value.
func (m *Manager) Get(name string) (*Contact, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	contact, ok := m.contacts[name]
	if !ok {
		return nil, errors.WithStack(ErrNotFound)
	}

	return &contact, nil
}

// Set sets or adds a contact.
//
// A name is required because the name of the contact could change.
func (m *Manager) Set(name string, contact *Contact) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Clone values in case saving fails.
	contacts := map[string]Contact{}
	for k, v := range m.contacts {
		contacts[k] = v
	}

	contacts[name] = *contact

	if err := save(contacts, m.filename); err != nil {
		return err
	}

	m.contacts = contacts

	return nil
}

// Delete deletes a contact by name.
func (m *Manager) Delete(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.contacts[name]; !ok {
		return errors.WithStack(ErrNotFound)
	}

	// Clone values in case saving fails.
	contacts := map[string]Contact{}
	for k, v := range m.contacts {
		if k != name {
			contacts[k] = v
		}
	}

	if err := save(contacts, m.filename); err != nil {
		return err
	}

	m.contacts = contacts

	return nil
}

// load loads contacts from a file.
func load(filename string) (map[string]Contact, error) {
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

	var col collection
	if err := toml.NewDecoder(f).Decode(&col); err != nil {
		return nil, errors.WithStack(err)
	}

	contacts := map[string]Contact{}

	for _, rec := range col.Records {
		pid, err := peer.IDB58Decode(rec.PeerID)
		if err != nil {
			return nil, errors.WithStack(err)
		}

		contacts[rec.Name] = Contact{PeerID: pid}
	}

	return contacts, nil
}

// save saves contacts to a file.
func save(contacts map[string]Contact, filename string) error {
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

	col := collection{}
	for name, contact := range contacts {
		col.Records = append(col.Records, record{
			Name:   name,
			PeerID: contact.PeerID.Pretty(),
		})
	}

	return errors.WithStack(toml.NewEncoder(f).Encode(col))
}
