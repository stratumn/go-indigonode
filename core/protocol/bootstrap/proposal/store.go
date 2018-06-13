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
	"sync"
	"time"

	logging "gx/ipfs/QmSpJByNKFX1sCsHBEp3R73FL4NF6FnQTEGyNAXHm2GS52/go-log"
	"gx/ipfs/QmcJukH2sAFjY3HdBKq35WDzWoL3UUu2gt9wdfqZTUyM74/go-libp2p-peer"
	"gx/ipfs/Qme1knMqwt1hKZbc1BmQFmnm9f36nyQGwXxPGVpVJ9rMK5/go-libp2p-crypto"
)

// InMemoryStore stores the requests in memory only.
type InMemoryStore struct {
	requestsLock sync.RWMutex
	requests     map[peer.ID]*Request

	votesLock sync.RWMutex
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

	s.requestsLock.Lock()
	defer s.requestsLock.Unlock()

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

	err = v.Verify(r)
	if err != nil {
		event.SetError(err)
		return err
	}

	pk, _ := crypto.UnmarshalPublicKey(v.Signature.PublicKey)
	votingPeer, _ := peer.IDFromPublicKey(pk)

	s.votesLock.Lock()
	defer s.votesLock.Unlock()

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

	s.requestsLock.Lock()
	delete(s.requests, peerID)
	s.requestsLock.Unlock()

	s.votesLock.Lock()
	delete(s.votes, peerID)
	s.votesLock.Unlock()

	return nil
}

// Get a request for a given PeerID.
func (s *InMemoryStore) Get(ctx context.Context, peerID peer.ID) (*Request, error) {
	event := log.EventBegin(ctx, "Get", peerID)
	defer event.Done()

	now := time.Now().UTC()

	s.requestsLock.RLock()
	r, ok := s.requests[peerID]
	event.Append(logging.Metadata{"found": ok})
	s.requestsLock.RUnlock()

	if !ok {
		return nil, nil
	}

	if !r.Expires.IsZero() && r.Expires.Before(now) {
		event.Append(logging.Metadata{"expired": true})

		s.requestsLock.Lock()
		delete(s.requests, peerID)
		s.requestsLock.Unlock()

		s.votesLock.Lock()
		delete(s.votes, peerID)
		s.votesLock.Unlock()

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

	s.votesLock.RLock()
	defer s.votesLock.RUnlock()

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

	s.requestsLock.Lock()
	defer s.requestsLock.Unlock()

	for peerID, peerRequest := range s.requests {
		if !peerRequest.Expires.IsZero() && peerRequest.Expires.Before(now) {
			delete(s.requests, peerID)

			s.votesLock.Lock()
			delete(s.votes, peerID)
			s.votesLock.Unlock()
		} else {
			results = append(results, peerRequest)
		}
	}

	event.Append(logging.Metadata{"count": len(results)})
	return results, nil
}
