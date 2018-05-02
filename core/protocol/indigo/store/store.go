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

	"github.com/pkg/errors"
	"github.com/stratumn/alice/core/protocol/indigo/store/audit"
	"github.com/stratumn/alice/core/protocol/indigo/store/constants"
	"github.com/stratumn/alice/core/protocol/indigo/store/sync"
	"github.com/stratumn/go-indigocore/cs"
	"github.com/stratumn/go-indigocore/postgresstore"
	"github.com/stratumn/go-indigocore/store"
	"github.com/stratumn/go-indigocore/types"

	logging "gx/ipfs/QmSpJByNKFX1sCsHBEp3R73FL4NF6FnQTEGyNAXHm2GS52/go-log"
)

var log = logging.Logger("indigo.store")

// Store implements github.com/stratumn/go-indigocore/store.Adapter.
type Store struct {
	store        store.Adapter
	auditStore   audit.Store
	sync         sync.Engine
	networkMgr   NetworkManager
	segmentsChan <-chan *cs.Segment
}

// New creates a new Indigo store.
// It expects a NetworkManager connected to a PoP network.
func New(
	networkMgr NetworkManager,
	sync sync.Engine,
	adapter store.Adapter,
	auditStore audit.Store,
) *Store {
	store := &Store{
		store:      adapter,
		auditStore: auditStore,
		networkMgr: networkMgr,
		sync:       sync,
	}

	store.segmentsChan = networkMgr.AddListener()
	go store.listenNetwork()

	return store
}

// listenNetwork listens to incoming segments and stores them.
// If invalid segments are received, they are kept for auditing.
func (s *Store) listenNetwork() {
	for {
		ctx := context.Background()
		segment, ok := <-s.segmentsChan
		if !ok {
			log.Event(ctx, "ListenChanClosed")
			return
		}

		s.storeNetworkSegment(ctx, segment)
	}
}

func (s *Store) addAuditTrail(ctx context.Context, segment *cs.Segment) {
	event := log.EventBegin(ctx, "AddAuditTrail")
	defer event.Done()

	if err := s.auditStore.AddSegment(ctx, segment); err != nil {
		event.SetError(err)
	}
}

func (s *Store) storeNetworkSegment(ctx context.Context, segment *cs.Segment) {
	event := log.EventBegin(ctx, "NetworkNewSegment")
	defer event.Done()

	if segment == nil {
		event.SetError(errors.New("nil segment"))
		return
	}

	if err := s.verifySegmentEvidence(ctx, segment); err != nil {
		event.SetError(err)
		return
	}

	seg, _ := s.store.GetSegment(ctx, segment.GetLinkHash())
	if seg != nil {
		event.Append(logging.Metadata{"already_stored": true})
		return
	}

	if err := s.syncMissingLinks(ctx, segment); err != nil {
		event.SetError(errors.Wrap(err, "could not sync missing links"))
		return
	}

	if err := segment.Link.Validate(ctx, s.GetSegment); err != nil {
		event.SetError(errors.Wrap(err, "invalid link"))
		s.addAuditTrail(ctx, segment)
		return
	}

	_, err := s.store.CreateLink(ctx, &segment.Link)
	if err != nil {
		event.SetError(errors.Wrap(err, "could not add to store"))
		return
	}
}

func (s *Store) verifySegmentEvidence(ctx context.Context, segment *cs.Segment) (err error) {
	event := log.EventBegin(ctx, "verifySegmentEvidence")
	defer func() {
		if err != nil {
			event.SetError(err)
		}

		event.Done()
	}()

	networkEvidences := segment.Meta.FindEvidences(audit.PeerSignatureBackend)
	event.Append(logging.Metadata{"evidence_count": len(networkEvidences)})

	if len(networkEvidences) == 0 {
		return audit.ErrMissingPeerSignature
	}

	defer func() {
		// If we have at least one valid signature on an invalid segment,
		// it's interesting to store for auditing.
		if err != nil {
			auditable := false
			for _, evidence := range networkEvidences {
				if evidence.Proof.Verify(segment.GetLinkHash()[:]) {
					auditable = true
					break
				}
			}

			if auditable {
				s.addAuditTrail(ctx, segment)
			}
		}
	}()

	peerID, err := constants.GetLinkNodeID(&segment.Link)
	if err != nil {
		return err
	}

	event.Append(logging.Metadata{"sender": peerID.Pretty()})

	networkEvidence := segment.Meta.GetEvidence(peerID.Pretty())
	if networkEvidence == nil || networkEvidence.Backend != audit.PeerSignatureBackend {
		return audit.ErrMissingPeerSignature
	}

	if !networkEvidence.Proof.Verify(segment.GetLinkHash()[:]) {
		return audit.ErrInvalidPeerSignature
	}

	return nil
}

// syncMissingLinks assumes that the incoming segment's evidence
// has been validated.
func (s *Store) syncMissingLinks(ctx context.Context, segment *cs.Segment) (err error) {
	event := log.EventBegin(ctx, "SyncMissingLinks")
	defer func() {
		if err != nil {
			event.SetError(err)
		}

		event.Done()
	}()

	missedLinks, err := s.sync.GetMissingLinks(ctx, &segment.Link, s.store)
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
				s.addAuditTrail(ctx, segment)
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
func (s *Store) Close(ctx context.Context) error {
	log.Event(ctx, "Close")

	switch a := s.store.(type) {
	case *postgresstore.Store:
		return a.Close()
	}
	s.networkMgr.RemoveListener(s.segmentsChan)
	return nil
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

	constants.SetLinkNodeID(link, s.networkMgr.NodeID())

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
