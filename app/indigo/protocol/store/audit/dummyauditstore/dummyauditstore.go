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

	"github.com/stratumn/alice/app/indigo/protocol/store/audit"
	"github.com/stratumn/alice/app/indigo/protocol/store/constants"
	"github.com/stratumn/go-indigocore/cs"

	logging "gx/ipfs/QmSpJByNKFX1sCsHBEp3R73FL4NF6FnQTEGyNAXHm2GS52/go-log"
	peer "gx/ipfs/QmcJukH2sAFjY3HdBKq35WDzWoL3UUu2gt9wdfqZTUyM74/go-libp2p-peer"
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
		if err := segment.Meta.AddEvidence(*e); err != nil {
			log.Event(ctx, "AddEvidenceError", logging.Metadata{
				"linkHash": segment.GetLinkHashString(),
				"err":      err.Error(),
			})
		}
	}
}

// AddSegment stores a segment in memory.
func (s *DummyAuditStore) AddSegment(ctx context.Context, segment *cs.Segment) error {
	e := log.EventBegin(ctx, "AddSegment", logging.Metadata{
		"linkHash": segment.GetLinkHashString(),
	})
	defer e.Done()

	s.auditMapLock.Lock()
	defer s.auditMapLock.Unlock()

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
func (s *DummyAuditStore) GetByPeer(ctx context.Context, peerID peer.ID, p audit.Pagination) (cs.SegmentSlice, error) {
	e := log.EventBegin(ctx, "GetByPeer", logging.Metadata{"peerID": peerID})
	defer e.Done()

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
