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

// Package postgresauditstore implements the audit.Store interface.
// It stores links and their evidences in a PostgreSQL database.
package postgresauditstore

import (
	"context"
	"database/sql"

	"github.com/pkg/errors"
	"github.com/stratumn/go-indigocore/cs"
	"github.com/stratumn/go-indigocore/postgresstore"
	"github.com/stratumn/go-indigonode/app/indigo/protocol/store/audit"
	"github.com/stratumn/go-indigonode/app/indigo/protocol/store/constants"
	"github.com/stratumn/go-indigonode/core/monitoring"

	peer "gx/ipfs/QmQsErDt8Qgw1XrsXf2BpEzDgGWtB1YLsTAARBup5b6B9W/go-libp2p-peer"
)

// PostgresAuditStore implements the audit.Store interface.
// It stores segments in a PostgreSQL database.
type PostgresAuditStore struct {
	*reader
	*writer

	db    *sql.DB
	stmts *stmts
}

// New creates a new PostgresAuditStore.
func New(config *postgresstore.Config) (*PostgresAuditStore, error) {
	db, err := sql.Open("postgres", config.URL)
	if err != nil {
		return nil, err
	}

	return &PostgresAuditStore{
		db: db,
	}, nil
}

// AddSegment stores a segment and its signature in the DB.
func (s *PostgresAuditStore) AddSegment(ctx context.Context, segment *cs.Segment) error {
	ctx, span := monitoring.StartSpan(ctx, "indigo.store.audit", "AddSegment")
	defer span.End()

	span.AddStringAttribute("link_hash", segment.GetLinkHashString())

	peerID, err := constants.GetLinkNodeID(&segment.Link)
	if err != nil {
		span.SetUnknownError(err)
		return err
	}

	span.SetPeerID(peerID)

	lh, err := s.createLink(ctx, &segment.Link, peerID)
	if err != nil {
		span.SetUnknownError(err)
		return errors.WithStack(err)
	}

	for _, evidence := range segment.Meta.Evidences {
		if err := s.addEvidence(ctx, lh, evidence); err != nil {
			span.SetUnknownError(err)
			return err
		}
	}

	return nil
}

// GetByPeer returns segments saved in the database.
func (s *PostgresAuditStore) GetByPeer(ctx context.Context, peerID peer.ID, p audit.Pagination) (cs.SegmentSlice, error) {
	ctx, span := monitoring.StartSpan(ctx, "indigo.store.audit", "GetByPeer", monitoring.SpanOptionPeerID(peerID))
	defer span.End()

	p.InitIfInvalid()

	res, err := s.FindSegments(ctx, &audit.SegmentFilter{
		PeerID:     peerID,
		Pagination: p,
	})
	if err != nil {
		span.SetUnknownError(err)
	}

	return res, err
}

// Create creates the database tables and indexes.
// Note that the actual database itself needs to be created before calling Create().
func (s *PostgresAuditStore) Create() error {
	for _, query := range sqlCreate {
		if _, err := s.db.Exec(query); err != nil {
			return err
		}
	}
	return nil
}

// Prepare prepares the database statements.
// It should be called once before interacting with segments.
// It assumes the tables have been created using Create().
func (s *PostgresAuditStore) Prepare() error {
	stmts, err := newStmts(s.db)
	if err != nil {
		return err
	}
	s.stmts = stmts
	s.reader = &reader{stmts: s.stmts.readStmts}
	s.writer = &writer{stmts: s.stmts.writeStmts}
	return nil
}

// Drop drops the database tables and indexes.
func (s *PostgresAuditStore) Drop() error {
	for _, query := range sqlDrop {
		if _, err := s.db.Exec(query); err != nil {
			return err
		}
	}
	return nil
}

// Close closes the database connection.
func (s *PostgresAuditStore) Close() error {
	return s.db.Close()
}
