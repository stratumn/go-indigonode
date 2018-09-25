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

// Package dummyauditstore implements the audit.Store interface.
// It stores links in memory with no persistence.
package dummyauditstore

import (
	"context"
	"sync"

	"github.com/stratumn/go-indigocore/cs"
	"github.com/stratumn/go-indigonode/app/indigo/protocol/store/audit"
	"github.com/stratumn/go-indigonode/app/indigo/protocol/store/constants"
	"github.com/stratumn/go-indigonode/core/monitoring"

	peer "gx/ipfs/QmQsErDt8Qgw1XrsXf2BpEzDgGWtB1YLsTAARBup5b6B9W/go-libp2p-peer"
)

// DummyAuditStore implements the audit.Store interface.
// It stores links in memory with no persistence.
type DummyAuditStore struct {
	auditMapLock sync.RWMutex
	auditMap     map[peer.ID]cs.SegmentSlice
}

// New creates a new DummyAuditStore.
func New() audit.Store {
	return &DummyAuditStore{
		auditMap: make(map[peer.ID]cs.SegmentSlice),
	}
}

func (s *DummyAuditStore) addEvidence(ctx context.Context, segment *cs.Segment, evidences cs.Evidences) {
	ctx, span := monitoring.StartSpan(ctx, "indigo.store.audit", "addEvidence")
	defer span.End()

	for _, e := range evidences {
		if err := segment.Meta.AddEvidence(*e); err != nil {
			span.Annotate(ctx, "add_evidence_error", err.Error())
		}
	}
}

// AddSegment stores a segment in memory.
func (s *DummyAuditStore) AddSegment(ctx context.Context, segment *cs.Segment) error {
	ctx, span := monitoring.StartSpan(ctx, "indigo.store.audit", "AddSegment")
	defer span.End()

	span.AddStringAttribute("link_hash", segment.GetLinkHashString())

	s.auditMapLock.Lock()
	defer s.auditMapLock.Unlock()

	peerID, err := constants.GetLinkNodeID(&segment.Link)
	if err != nil {
		span.SetUnknownError(err)
		return err
	}

	span.SetPeerID(peerID)

	peerSegments, ok := s.auditMap[peerID]
	if !ok {
		peerSegments = cs.SegmentSlice{}
	}

	// If the segment is already stored, update its evidences.
	for _, peerSegment := range peerSegments {
		if peerSegment.GetLinkHashString() == segment.GetLinkHashString() {
			s.addEvidence(ctx, peerSegment, segment.Meta.Evidences)
			return nil
		}
	}

	peerSegments = append(peerSegments, segment)
	s.auditMap[peerID] = peerSegments

	return nil
}

// GetByPeer returns links saved in memory.
func (s *DummyAuditStore) GetByPeer(ctx context.Context, peerID peer.ID, p audit.Pagination) (cs.SegmentSlice, error) {
	_, span := monitoring.StartSpan(ctx, "indigo.store.audit", "GetByPeer", monitoring.SpanOptionPeerID(peerID))
	defer span.End()

	s.auditMapLock.RLock()
	defer s.auditMapLock.RUnlock()

	p.InitIfInvalid()

	peerLinks, ok := s.auditMap[peerID]
	if !ok || uint(len(peerLinks)) <= p.Skip {
		return cs.SegmentSlice{}, nil
	}

	if p.Top == 0 {
		return peerLinks, nil
	}

	stop := p.Skip + p.Top
	if uint(len(peerLinks)) < stop {
		stop = uint(len(peerLinks))
	}

	return peerLinks[p.Skip:stop], nil
}
