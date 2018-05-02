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

// Package dummyauditstore implements the audit.Store interface.
// It stores links in memory with no persistence.
package dummyauditstore

import (
	"context"
	"sync"

	logging "gx/ipfs/QmSpJByNKFX1sCsHBEp3R73FL4NF6FnQTEGyNAXHm2GS52/go-log"
	peer "gx/ipfs/QmZoWKhxUmZ2seW4BzX6fJkNR8hh9PsGModr7q171yq2SS/go-libp2p-peer"

	"github.com/stratumn/alice/core/protocol/indigo/store/audit"
	"github.com/stratumn/alice/core/protocol/indigo/store/constants"
	"github.com/stratumn/go-indigocore/cs"
)

var (
	log = logging.Logger("indigo.store.audit.dummy")
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
	for _, e := range evidences {
		if segment.Meta.AddEvidence(*e) != nil {
			log.Event(ctx, "Trying to add existing evidence", logging.Metadata{
				"linkHash": segment.GetLinkHashString(),
			})
		}
	}
}

// AddSegment stores a segment in memory.
func (s *DummyAuditStore) AddSegment(ctx context.Context, segment *cs.Segment) error {
	s.auditMapLock.Lock()
	defer s.auditMapLock.Unlock()

	e := log.EventBegin(ctx, "AddSegment", logging.Metadata{
		"linkHash": segment.GetLinkHashString(),
	})
	defer e.Done()

	peerID, err := constants.GetLinkNodeID(&segment.Link)
	if err != nil {
		return err
	}

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
func (s *DummyAuditStore) GetByPeer(ctx context.Context, peerID peer.ID, p *audit.Pagination) (cs.SegmentSlice, error) {
	s.auditMapLock.RLock()
	defer s.auditMapLock.RUnlock()

	e := log.EventBegin(ctx, "GetByPeer", logging.Metadata{
		"peerID": peerID,
	})
	defer e.Done()

	if p == nil {
		p = &audit.Pagination{
			Skip: 0,
			Top:  audit.DefaultLimit,
		}
	}

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
