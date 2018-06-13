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

	var requests []*Request
	err = json.Unmarshal(b, &requests)
	if err != nil {
		event.SetError(errors.WithStack(err))
		return err
	}

	for _, r := range requests {
		err = s.Store.AddRequest(ctx, r)
		if err != nil {
			event.SetError(err)
			return err
		}
	}

	return nil
}

// Save saves the data to disk.
func (s *FileSaver) Save(ctx context.Context) error {
	event := log.EventBegin(ctx, "FileSaver.Save")
	defer event.Done()

	rr, err := s.Store.List(ctx)
	if err != nil {
		event.SetError(err)
		return err
	}

	b, err := json.Marshal(rr)
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
