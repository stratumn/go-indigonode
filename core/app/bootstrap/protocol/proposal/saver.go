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

package proposal

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/pkg/errors"

	"gx/ipfs/QmcJukH2sAFjY3HdBKq35WDzWoL3UUu2gt9wdfqZTUyM74/go-libp2p-peer"
)

// FileSaver saves the store data to a file
// when it changes.
type FileSaver struct {
	Store
	path string
}

// WrapWithSaver wraps an existing Store and saves changes to a file.
// It loads existing data if present.
func WrapWithSaver(ctx context.Context, s Store, path string) (Store, error) {
	ss := &FileSaver{
		Store: s,
		path:  path,
	}

	err := ss.Load(ctx)
	if err != nil {
		return nil, err
	}

	return ss, nil
}

// Load loads previous data.
func (s *FileSaver) Load(ctx context.Context) error {
	event := log.EventBegin(ctx, "FileSaver.Load")
	defer event.Done()

	// Create the directory if it doesn't exist.
	configDir, _ := filepath.Split(s.path)
	if err := os.MkdirAll(configDir, 0744); err != nil {
		event.SetError(errors.WithStack(err))
		return err
	}

	_, err := os.Stat(s.path)
	if os.IsNotExist(err) {
		return nil
	}

	b, err := ioutil.ReadFile(s.path)
	if err != nil {
		event.SetError(errors.WithStack(err))
		return err
	}

	var store InMemoryStore
	err = json.Unmarshal(b, &store)
	if err != nil {
		event.SetError(errors.WithStack(err))
		return err
	}

	reqs, err := store.List(ctx)
	if err != nil {
		event.SetError(err)
		return err
	}

	for _, req := range reqs {
		err = s.Store.AddRequest(ctx, req)
		if err != nil {
			event.SetError(err)
			return err
		}

		votes, err := store.GetVotes(ctx, req.PeerID)
		if err != nil {
			event.SetError(err)
			return err
		}

		for _, vote := range votes {
			err = s.Store.AddVote(ctx, vote)
			if err != nil {
				event.SetError(err)
				return err
			}
		}
	}

	return nil
}

// Save saves the data to disk.
func (s *FileSaver) Save(ctx context.Context) error {
	event := log.EventBegin(ctx, "FileSaver.Save")
	defer event.Done()

	b, err := json.Marshal(s.Store)
	if err != nil {
		event.SetError(err)
		return err
	}

	err = ioutil.WriteFile(s.path, b, 0644)
	if err != nil {
		event.SetError(err)
		return err
	}

	return nil
}

// AddRequest adds a new request and saves to disk.
func (s *FileSaver) AddRequest(ctx context.Context, r *Request) error {
	err := s.Store.AddRequest(ctx, r)
	if err != nil {
		return err
	}

	return s.Save(ctx)
}

// AddVote adds a new vote and saves to disk.
func (s *FileSaver) AddVote(ctx context.Context, v *Vote) error {
	err := s.Store.AddVote(ctx, v)
	if err != nil {
		return err
	}

	return s.Save(ctx)
}

// Remove a request and saves to disk.
func (s *FileSaver) Remove(ctx context.Context, peerID peer.ID) error {
	err := s.Store.Remove(ctx, peerID)
	if err != nil {
		return err
	}

	return s.Save(ctx)
}

// List all the pending requests and saves to disk (because expired requests
// need to be removed).
func (s *FileSaver) List(ctx context.Context) ([]*Request, error) {
	results, err := s.Store.List(ctx)
	if err != nil {
		return nil, err
	}

	return results, s.Save(ctx)
}
