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

package store

import (
	"context"

	"github.com/pkg/errors"
	"github.com/stratumn/go-indigocore/cs"
	"github.com/stratumn/go-indigocore/postgresstore"
	"github.com/stratumn/go-indigocore/store"
	"github.com/stratumn/go-indigocore/types"
	"github.com/stratumn/go-indigocore/validator"
	"github.com/stratumn/go-indigonode/app/indigo/protocol/store/audit"
	"github.com/stratumn/go-indigonode/app/indigo/protocol/store/constants"
	"github.com/stratumn/go-indigonode/app/indigo/protocol/store/sync"
	"github.com/stratumn/go-indigonode/core/monitoring"
)

// Store implements github.com/stratumn/go-indigocore/store.Adapter.
type Store struct {
	store        store.Adapter
	govMgr       validator.GovernanceManager
	auditStore   audit.Store
	sync         sync.Engine
	networkMgr   NetworkManager
	segmentsChan <-chan *cs.Segment
}

// New creates a new Indigo store.
// It expects a NetworkManager connected to a PoP network.
func New(
	ctx context.Context,
	networkMgr NetworkManager,
	sync sync.Engine,
	adapter store.Adapter,
	auditStore audit.Store,
	governanceManager validator.GovernanceManager,
) *Store {
	store := &Store{
		store:      adapter,
		auditStore: auditStore,
		govMgr:     governanceManager,
		networkMgr: networkMgr,
		sync:       sync,
	}

	store.segmentsChan = networkMgr.AddListener()

	go store.listenNetwork()
	go func() {
		ctx, span := monitoring.StartSpan(ctx, "indigo.store", "GovernanceManager")
		defer span.End()

		if err := store.govMgr.ListenAndUpdate(ctx); err != nil {
			span.SetUnknownError(err)
		}
	}()

	return store
}

// listenNetwork listens to incoming segments and stores them.
// If invalid segments are received, they are kept for auditing.
func (s *Store) listenNetwork() {
	for {
		ctx := context.Background()
		segment, ok := <-s.segmentsChan
		if !ok {
			return
		}

		if err := s.storeNetworkSegment(ctx, segment); err != nil {
			continue
		}
	}
}

func (s *Store) addAuditTrail(ctx context.Context, segment *cs.Segment) {
	ctx, span := monitoring.StartSpan(ctx, "indigo.store", "addAuditTrail")
	defer span.End()

	invalidSegments.Record(ctx, 1)

	if err := s.auditStore.AddSegment(ctx, segment); err != nil {
		span.SetUnknownError(err)
	}
}

func (s *Store) storeNetworkSegment(ctx context.Context, segment *cs.Segment) (err error) {
	ctx, span := monitoring.StartSpan(ctx, "indigo.store", "storeNetworkSegment")
	ctx = monitoring.NewTaggedContext(ctx).Tag(monitoring.ErrorTag, "success").Build()
	defer func() {
		if err != nil {
			span.SetUnknownError(err)
			ctx = monitoring.NewTaggedContext(ctx).Tag(monitoring.ErrorTag, err.Error()).Build()
		}

		segmentsReceived.Record(ctx, 1)
		span.End()
	}()

	if segment == nil {
		return errors.New("nil segment")
	}

	if err := s.verifySegmentEvidence(ctx, segment); err != nil {
		return err
	}

	seg, _ := s.store.GetSegment(ctx, segment.GetLinkHash())
	if seg != nil {
		span.AddBoolAttribute("already_stored", true)
		return nil
	}

	if err := s.syncMissingLinks(ctx, segment); err != nil {
		return errors.Wrap(err, "could not sync missing links")
	}

	if err := segment.Link.Validate(ctx, s.GetSegment); err != nil {
		s.addAuditTrail(ctx, segment)
		return errors.Wrap(err, "invalid link")
	}

	if rulesValidator := s.govMgr.Current(); rulesValidator != nil {
		incomingValidatorHash, err := constants.GetValidatorHash(&segment.Link)
		if err != nil {
			s.addAuditTrail(ctx, segment)
			return errors.Wrap(err, "missing validator hash")
		}
		validatorHash, err := rulesValidator.Hash()
		if err != nil {
			return errors.Wrap(err, "could not get current validator hash")
		}
		if !validatorHash.Equals(incomingValidatorHash) {
			s.addAuditTrail(ctx, segment)
			return errors.New("validator hash does not match")
		}
		if err := rulesValidator.Validate(ctx, s.store, &segment.Link); err != nil {
			s.addAuditTrail(ctx, segment)
			return errors.Wrap(err, "link validation failed")
		}
	}

	_, err = s.store.CreateLink(ctx, &segment.Link)
	if err != nil {
		return errors.Wrap(err, "could not add to store")
	}

	return nil
}

func (s *Store) verifySegmentEvidence(ctx context.Context, segment *cs.Segment) (err error) {
	ctx, span := monitoring.StartSpan(ctx, "indigo.store", "verifySegmentEvidence")
	defer func() {
		if err != nil {
			span.SetUnknownError(err)
		}

		span.End()
	}()

	networkEvidences := segment.Meta.FindEvidences(audit.PeerSignatureBackend)
	span.AddIntAttribute("evidence_count", int64(len(networkEvidences)))

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

	span.SetPeerID(peerID)

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
	ctx, span := monitoring.StartSpan(ctx, "indigo.store", "syncMissingLinks")
	defer func() {
		if err != nil {
			span.SetUnknownError(err)
		}

		span.End()
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
func (s *Store) Close(ctx context.Context) (err error) {
	_, span := monitoring.StartSpan(ctx, "indigo.store", "Close")
	defer span.End()

	switch a := s.store.(type) {
	case *postgresstore.Store:
		err = a.Close()
		if err != nil {
			span.SetUnknownError(err)
		}
	}

	s.networkMgr.RemoveListener(s.segmentsChan)
	return
}

// GetInfo returns information about the underlying store.
func (s *Store) GetInfo(ctx context.Context) (interface{}, error) {
	ctx, span := monitoring.StartSpan(ctx, "indigo.store", "GetInfo")
	defer span.End()

	return s.store.GetInfo(ctx)
}

// CreateLink forwards the request to the underlying store.
func (s *Store) CreateLink(ctx context.Context, link *cs.Link) (lh *types.Bytes32, err error) {
	ctx, span := monitoring.StartSpan(ctx, "indigo.store", "CreateLink")
	ctx = monitoring.NewTaggedContext(ctx).Tag(monitoring.ErrorTag, "success").Build()
	defer func() {
		if err != nil {
			span.SetUnknownError(err)
			ctx = monitoring.NewTaggedContext(ctx).Tag(monitoring.ErrorTag, err.Error()).Build()
		} else {
			span.AddStringAttribute("link_hash", lh.String())
		}

		segmentsCreated.Record(ctx, 1)
		span.End()
	}()

	err = link.Validate(ctx, s.GetSegment)
	if err != nil {
		return
	}

	if rulesValidator := s.govMgr.Current(); rulesValidator != nil {
		if err = rulesValidator.Validate(ctx, s.store, link); err != nil {
			return nil, err
		}
		rulesHash, err := rulesValidator.Hash()
		if err != nil {
			return nil, err
		}
		constants.SetValidatorHash(link, rulesHash)
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
	ctx, span := monitoring.StartSpan(ctx, "indigo.store", "GetSegment")
	defer span.End()

	span.AddStringAttribute("link_hash", linkHash.String())

	return s.store.GetSegment(ctx, linkHash)
}

// FindSegments forwards the request to the underlying store.
func (s *Store) FindSegments(ctx context.Context, filter *store.SegmentFilter) (cs.SegmentSlice, error) {
	ctx, span := monitoring.StartSpan(ctx, "indigo.store", "FindSegments")
	defer span.End()

	return s.store.FindSegments(ctx, filter)
}

// GetMapIDs forwards the request to the underlying store.
func (s *Store) GetMapIDs(ctx context.Context, filter *store.MapFilter) ([]string, error) {
	ctx, span := monitoring.StartSpan(ctx, "indigo.store", "GetMapIDs")
	defer span.End()

	return s.store.GetMapIDs(ctx, filter)
}

// AddEvidence forwards the request to the underlying store.
func (s *Store) AddEvidence(ctx context.Context, linkHash *types.Bytes32, evidence *cs.Evidence) error {
	ctx, span := monitoring.StartSpan(ctx, "indigo.store", "AddEvidence")
	defer span.End()

	span.AddStringAttribute("link_hash", linkHash.String())

	return s.store.AddEvidence(ctx, linkHash, evidence)
}

// GetEvidences forwards the request to the underlying store.
func (s *Store) GetEvidences(ctx context.Context, linkHash *types.Bytes32) (*cs.Evidences, error) {
	ctx, span := monitoring.StartSpan(ctx, "indigo.store", "GetEvidences")
	defer span.End()

	span.AddStringAttribute("link_hash", linkHash.String())

	return s.store.GetEvidences(ctx, linkHash)
}

// AddStoreEventChannel forwards the request to the underlying store.
func (s *Store) AddStoreEventChannel(c chan *store.Event) {
	s.store.AddStoreEventChannel(c)
}

// NewBatch forwards the request to the underlying store.
func (s *Store) NewBatch(ctx context.Context) (store.Batch, error) {
	ctx, span := monitoring.StartSpan(ctx, "indigo.store", "NewBatch")
	defer span.End()

	return s.store.NewBatch(ctx)
}
