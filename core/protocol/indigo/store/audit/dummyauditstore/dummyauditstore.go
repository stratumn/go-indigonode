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

	"github.com/pkg/errors"
	"github.com/stratumn/alice/core/protocol/indigo/store/audit"
	pb "github.com/stratumn/alice/pb/indigo/store"

	peer "gx/ipfs/QmZoWKhxUmZ2seW4BzX6fJkNR8hh9PsGModr7q171yq2SS/go-libp2p-peer"
)

// DummyAuditStore implements the audit.Store interface.
// It stores links in memory with no persistence.
type DummyAuditStore struct {
	auditMapLock sync.RWMutex
	auditMap     map[peer.ID][]pb.SignedLink
}

// NewDummyAuditStore creates a new DummyAuditStore.
func NewDummyAuditStore() audit.Store {
	return &DummyAuditStore{
		auditMap: make(map[peer.ID][]pb.SignedLink),
	}
}

// AddLink stores a link in memory.
func (s *DummyAuditStore) AddLink(_ context.Context, l *pb.SignedLink) error {
	s.auditMapLock.Lock()
	defer s.auditMapLock.Unlock()

	peerID, err := peer.IDFromBytes(l.From)
	if err != nil {
		return errors.WithStack(err)
	}

	peerLinks, ok := s.auditMap[peerID]
	if !ok {
		peerLinks = make([]pb.SignedLink, 0, 1)
	}

	peerLinks = append(peerLinks, *l)
	s.auditMap[peerID] = peerLinks

	return nil
}

// GetByPeer returns links saved in memory.
func (s *DummyAuditStore) GetByPeer(_ context.Context, peerID peer.ID, p audit.Pagination) ([]pb.SignedLink, error) {
	s.auditMapLock.RLock()
	defer s.auditMapLock.RUnlock()

	peerLinks, ok := s.auditMap[peerID]
	if !ok || uint(len(peerLinks)) <= p.Skip {
		return nil, nil
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
