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

//go:generate mockgen -package mockstore -destination mockstore/mockstore.go github.com/stratumn/go-indigocore/store Adapter

package store

import (
	"context"

	json "github.com/gibson042/canonicaljson-go"
	"github.com/pkg/errors"
	"github.com/stratumn/alice/core/protocol/indigo/store/audit"
	pb "github.com/stratumn/alice/pb/indigo/store"
	"github.com/stratumn/go-indigocore/cs"
	"github.com/stratumn/go-indigocore/store"
	"github.com/stratumn/go-indigocore/types"

	logging "gx/ipfs/QmSpJByNKFX1sCsHBEp3R73FL4NF6FnQTEGyNAXHm2GS52/go-log"
)

var log = logging.Logger("indigo.store")

// Store implements github.com/stratumn/go-indigocore/store.Adapter.
type Store struct {
	store      store.Adapter
	auditStore audit.Store
	sync       SyncEngine
	networkMgr NetworkManager
	linksChan  <-chan *pb.SignedLink
}

// New creates a new Indigo store.
// It expects a NetworkManager connected to a PoP network.
func New(
	networkMgr NetworkManager,
	sync SyncEngine,
	adapter store.Adapter,
	auditStore audit.Store,
) *Store {
	store := &Store{
		store:      adapter,
		auditStore: auditStore,
		networkMgr: networkMgr,
		sync:       sync,
	}

	store.linksChan = networkMgr.AddListener()
	go store.listenNetwork()

	return store
}

// listenNetwork listens to incoming links and stores them.
// If invalid links are received, they are kept for auditing.
func (s *Store) listenNetwork() {
	for {
		ctx := context.Background()
		remoteLink, ok := <-s.linksChan
		if !ok {
			log.Event(ctx, "ListenChanClosed")
			return
		}

		s.storeNetworkLink(ctx, remoteLink)
	}
}

func (s *Store) addAuditTrail(ctx context.Context, remoteLink *pb.SignedLink) {
	event := log.EventBegin(ctx, "AddAuditTrail", logging.Metadata{"sender": remoteLink.From})
	defer event.Done()

	if err := s.auditStore.AddLink(ctx, remoteLink); err != nil {
		event.SetError(err)
	}
}

func (s *Store) storeNetworkLink(ctx context.Context, remoteLink *pb.SignedLink) {
	event := log.EventBegin(ctx, "NetworkNewLink")
	defer event.Done()

	if !remoteLink.VerifySignature() {
		event.SetError(errors.New("invalid link signature"))
		return
	}

	var link cs.Link
	err := json.Unmarshal(remoteLink.Link, &link)
	if err != nil {
		event.SetError(errors.Wrap(err, "malformed link"))
		s.addAuditTrail(ctx, remoteLink)
		return
	}

	lh, _ := link.Hash()
	seg, _ := s.store.GetSegment(ctx, lh)
	if seg != nil {
		event.Append(logging.Metadata{"already_stored": true})
		return
	}

	if err = s.syncMissingLinks(ctx, &link, remoteLink); err != nil {
		event.SetError(errors.Wrap(err, "could not sync missing links"))
		return
	}

	if err = link.Validate(ctx, s.GetSegment); err != nil {
		event.SetError(errors.Wrap(err, "invalid link"))
		s.addAuditTrail(ctx, remoteLink)
		return
	}

	_, err = s.store.CreateLink(ctx, &link)
	if err != nil {
		event.SetError(errors.Wrap(err, "could not add to store"))
		return
	}
}

func (s *Store) syncMissingLinks(ctx context.Context, link *cs.Link, remoteLink *pb.SignedLink) (err error) {
	event := log.EventBegin(ctx, "SyncMissingLinks")
	defer func() {
		if err != nil {
			event.SetError(err)
		}

		event.Done()
	}()

	missedLinks, err := s.sync.GetMissingLinks(ctx, link, s.store)
	if err != nil {
		return err
	}

	if len(missedLinks) != 0 {
		b, err := s.store.NewBatch(ctx)
		if err != nil {
			return err
		}

		for _, l := range missedLinks {
			_, err := b.CreateLink(ctx, l)
			if err != nil {
				return err
			}
		}

		for _, l := range missedLinks {
			if err = l.Validate(ctx, b.GetSegment); err != nil {
				s.addAuditTrail(ctx, remoteLink)
				return err
			}
		}

		if err = b.Write(ctx); err != nil {
			return err
		}
	}

	return nil
}

// Close cleans up the store and stops it.
func (s *Store) Close(ctx context.Context) {
	log.Event(ctx, "Close")
	s.networkMgr.RemoveListener(s.linksChan)
}

// GetInfo returns information about the underlying store.
func (s *Store) GetInfo(ctx context.Context) (interface{}, error) {
	log.Event(ctx, "GetInfo")
	return s.store.GetInfo(ctx)
}

// CreateLink forwards the request to the underlying store.
func (s *Store) CreateLink(ctx context.Context, link *cs.Link) (lh *types.Bytes32, err error) {
	event := log.EventBegin(ctx, "CreateLink")
	defer func() {
		if err != nil {
			event.SetError(err)
		} else {
			event.Append(logging.Metadata{"link_hash": lh.String()})
		}

		event.Done()
	}()

	err = link.Validate(ctx, s.GetSegment)
	if err != nil {
		return
	}

	lh, err = s.store.CreateLink(ctx, link)
	if err != nil {
		return
	}

	if err = s.networkMgr.Publish(ctx, link); err != nil {
		return
	}

	return
}

// GetSegment forwards the request to the underlying store.
func (s *Store) GetSegment(ctx context.Context, linkHash *types.Bytes32) (*cs.Segment, error) {
	log.Event(ctx, "GetSegment", logging.Metadata{
		"link_hash": linkHash.String(),
	})

	return s.store.GetSegment(ctx, linkHash)
}

// FindSegments forwards the request to the underlying store.
func (s *Store) FindSegments(ctx context.Context, filter *store.SegmentFilter) (cs.SegmentSlice, error) {
	log.Event(ctx, "FindSegments")
	return s.store.FindSegments(ctx, filter)
}

// GetMapIDs forwards the request to the underlying store.
func (s *Store) GetMapIDs(ctx context.Context, filter *store.MapFilter) ([]string, error) {
	log.Event(ctx, "GetMapIDs")
	return s.store.GetMapIDs(ctx, filter)
}

// AddEvidence forwards the request to the underlying store.
func (s *Store) AddEvidence(ctx context.Context, linkHash *types.Bytes32, evidence *cs.Evidence) error {
	log.Event(ctx, "AddEvidence", logging.Metadata{
		"link_hash": linkHash.String(),
	})

	return s.store.AddEvidence(ctx, linkHash, evidence)
}

// GetEvidences forwards the request to the underlying store.
func (s *Store) GetEvidences(ctx context.Context, linkHash *types.Bytes32) (*cs.Evidences, error) {
	log.Event(ctx, "GetEvidences", logging.Metadata{
		"link_hash": linkHash.String(),
	})

	return s.store.GetEvidences(ctx, linkHash)
}

// AddStoreEventChannel forwards the request to the underlying store.
func (s *Store) AddStoreEventChannel(c chan *store.Event) {
	log.Event(context.Background(), "AddStoreEventChannel")
	s.store.AddStoreEventChannel(c)
}

// NewBatch forwards the request to the underlying store.
func (s *Store) NewBatch(ctx context.Context) (store.Batch, error) {
	log.Event(ctx, "NewBatch")
	return s.store.NewBatch(ctx)
}
