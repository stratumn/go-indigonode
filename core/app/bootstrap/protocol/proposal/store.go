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
	"sync"
	"time"

	"github.com/pkg/errors"

	logging "gx/ipfs/QmSpJByNKFX1sCsHBEp3R73FL4NF6FnQTEGyNAXHm2GS52/go-log"
	"gx/ipfs/QmcJukH2sAFjY3HdBKq35WDzWoL3UUu2gt9wdfqZTUyM74/go-libp2p-peer"
	"gx/ipfs/Qme1knMqwt1hKZbc1BmQFmnm9f36nyQGwXxPGVpVJ9rMK5/go-libp2p-crypto"
)

// InMemoryStore stores the requests in memory only.
type InMemoryStore struct {
	storeLock sync.RWMutex
	requests  map[peer.ID]*Request
	votes     map[peer.ID]map[peer.ID]*Vote
}

// NewInMemoryStore returns a new store without disk backup.
func NewInMemoryStore() Store {
	return &InMemoryStore{
		requests: make(map[peer.ID]*Request),
		votes:    make(map[peer.ID]map[peer.ID]*Vote),
	}
}

// AddRequest adds a new request.
func (s *InMemoryStore) AddRequest(ctx context.Context, r *Request) error {
	event := log.EventBegin(ctx, "AddRequest", r.PeerID)
	defer event.Done()

	if r.PeerID == "" {
		return ErrInvalidPeerID
	}

	if r.Type == AddNode && r.PeerAddr == nil {
		return ErrMissingPeerAddr
	}

	s.storeLock.Lock()
	defer s.storeLock.Unlock()

	s.requests[r.PeerID] = r

	return nil
}

// AddVote adds a vote to a request.
func (s *InMemoryStore) AddVote(ctx context.Context, v *Vote) error {
	event := log.EventBegin(ctx, "AddVote", v.PeerID)
	defer event.Done()

	r, err := s.Get(ctx, v.PeerID)
	if err != nil {
		return err
	}

	if r == nil {
		return ErrMissingRequest
	}

	err = v.Verify(ctx, r)
	if err != nil {
		event.SetError(err)
		return err
	}

	pk, _ := crypto.UnmarshalPublicKey(v.Signature.PublicKey)
	votingPeer, _ := peer.IDFromPublicKey(pk)

	s.storeLock.Lock()
	defer s.storeLock.Unlock()

	votes, ok := s.votes[v.PeerID]
	if !ok {
		votes = make(map[peer.ID]*Vote)
	}

	votes[votingPeer] = v
	s.votes[v.PeerID] = votes

	return nil
}

// Remove removes a request and its votes.
func (s *InMemoryStore) Remove(ctx context.Context, peerID peer.ID) error {
	defer log.EventBegin(ctx, "Remove", peerID).Done()

	s.storeLock.Lock()
	defer s.storeLock.Unlock()

	delete(s.requests, peerID)
	delete(s.votes, peerID)

	return nil
}

// Get a request for a given PeerID.
func (s *InMemoryStore) Get(ctx context.Context, peerID peer.ID) (*Request, error) {
	event := log.EventBegin(ctx, "Get", peerID)
	defer event.Done()

	now := time.Now().UTC()

	s.storeLock.RLock()
	r, ok := s.requests[peerID]
	event.Append(logging.Metadata{"found": ok})
	s.storeLock.RUnlock()

	if !ok {
		return nil, nil
	}

	if !r.Expires.IsZero() && r.Expires.Before(now) {
		event.Append(logging.Metadata{"expired": true})

		s.storeLock.Lock()
		delete(s.requests, peerID)
		delete(s.votes, peerID)
		s.storeLock.Unlock()

		r = nil
	}

	return r, nil
}

// GetVotes gets the votes for a given PeerID.
func (s *InMemoryStore) GetVotes(ctx context.Context, peerID peer.ID) ([]*Vote, error) {
	event := log.EventBegin(ctx, "GetVotes", peerID)
	defer event.Done()

	_, err := s.Get(ctx, peerID)
	if err != nil {
		return nil, err
	}

	s.storeLock.RLock()
	defer s.storeLock.RUnlock()

	votes, ok := s.votes[peerID]
	event.Append(logging.Metadata{"found": ok})

	var results []*Vote
	for _, vote := range votes {
		results = append(results, vote)
	}

	return results, nil
}

// List all the pending requests.
func (s *InMemoryStore) List(ctx context.Context) ([]*Request, error) {
	event := log.EventBegin(ctx, "List")
	defer event.Done()

	now := time.Now().UTC()
	var results []*Request

	s.storeLock.Lock()
	defer s.storeLock.Unlock()

	for peerID, peerRequest := range s.requests {
		if !peerRequest.Expires.IsZero() && peerRequest.Expires.Before(now) {
			delete(s.requests, peerID)
			delete(s.votes, peerID)
		} else {
			results = append(results, peerRequest)
		}
	}

	event.Append(logging.Metadata{"count": len(results)})
	return results, nil
}

// MarshalJSON marshals the store's content to JSON.
func (s *InMemoryStore) MarshalJSON() ([]byte, error) {
	toSerialize := make(map[string]struct {
		Request *Request
		Votes   []*Vote
	})

	s.storeLock.RLock()
	defer s.storeLock.RUnlock()

	for peerID, req := range s.requests {
		values := struct {
			Request *Request
			Votes   []*Vote
		}{}

		values.Request = req

		for _, vote := range s.votes[peerID] {
			values.Votes = append(values.Votes, vote)
		}

		toSerialize[peerID.Pretty()] = values
	}

	return json.Marshal(toSerialize)
}

// UnmarshalJSON unmarshals JSON content to the store.
func (s *InMemoryStore) UnmarshalJSON(data []byte) error {
	ctx := context.Background()

	deserialized := map[string]struct {
		Request *Request
		Votes   []*Vote
	}{}

	err := json.Unmarshal(data, &deserialized)
	if err != nil {
		return errors.WithStack(err)
	}

	if s.requests == nil {
		s.requests = make(map[peer.ID]*Request)
	}

	if s.votes == nil {
		s.votes = make(map[peer.ID]map[peer.ID]*Vote)
	}

	for _, values := range deserialized {
		err = s.AddRequest(ctx, values.Request)
		if err != nil {
			return err
		}

		for _, vote := range values.Votes {
			err = s.AddVote(ctx, vote)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
