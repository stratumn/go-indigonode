// Copyright 2017 Stratumn SAS. All rights reserved.
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

package postgresauditstore

import (
	"database/sql"
)

const (
	sqlCreateLink = `
		INSERT INTO audited_links (
			link_hash,
			data,
			peerID
		)
		VALUES ($1, $2, $3)
		ON CONFLICT (link_hash)
		DO UPDATE SET
			data = $2,
			peerID = $3
	`

	sqlGetSegment = `
		SELECT l.link_hash, l.data, s.data FROM audited_links l
		LEFT JOIN audited_evidences s ON l.link_hash = s.link_hash
		WHERE l.link_hash = $1
	`

	sqlFindSegments = `
		SELECT l.link_hash, l.data, l.peerID, s.data FROM audited_links l
		LEFT JOIN audited_evidences s ON l.link_hash = s.link_hash
		ORDER BY l.created_at ASC
		OFFSET $1 LIMIT $2
	`

	sqlFindSegmentsByPeer = `
		SELECT l.link_hash, l.data, l.peerID, s.data FROM audited_links l
		LEFT JOIN audited_evidences s ON l.link_hash = s.link_hash
		WHERE l.peerID = $1
		ORDER BY l.created_at ASC
		OFFSET $2 LIMIT $3
	`

	sqlGetEvidences = `
		SELECT data FROM audited_evidences
		WHERE link_hash = $1
	`

	sqlAddEvidence = `
		INSERT INTO audited_evidences (
			link_hash,
			data
		)
		VALUES ($1, $2)
		ON CONFLICT (link_hash)
		DO NOTHING
`
)

var sqlCreate = []string{
	`
		CREATE TABLE audited_links (
			id BIGSERIAL PRIMARY KEY,
			link_hash bytea NOT NULL,
			data jsonb NOT NULL,
			peerID bytea NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		)
	`,
	`
		CREATE UNIQUE INDEX audited_links_link_hash_idx
		ON audited_links (link_hash)
	`,
	`
		CREATE UNIQUE INDEX audited_links_peerID_idx
		ON audited_links (link_hash)
	`,
	`
		CREATE TABLE audited_evidences (
			id BIGSERIAL PRIMARY KEY,
			link_hash bytea NOT NULL,
			data bytea NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
	)
	`,
	`
		CREATE UNIQUE INDEX audited_evidences_link_hash_idx
		ON audited_evidences (link_hash)
	`,
}

var sqlDrop = []string{
	"DROP TABLE audited_links, audited_evidences",
}

type writeStmts struct {
	CreateLink  *sql.Stmt
	AddEvidence *sql.Stmt
}

type readStmts struct {
	GetSegment         *sql.Stmt
	FindSegments       *sql.Stmt
	FindSegmentsByPeer *sql.Stmt
}

type stmts struct {
	readStmts
	writeStmts
}

func newStmts(db *sql.DB) (*stmts, error) {
	var (
		s   stmts
		err error
	)

	prepare := func(str string) (stmt *sql.Stmt) {
		if err == nil {
			stmt, err = db.Prepare(str)
		}
		return
	}

	s.GetSegment = prepare(sqlGetSegment)
	s.FindSegments = prepare(sqlFindSegments)
	s.FindSegmentsByPeer = prepare(sqlFindSegmentsByPeer)

	s.CreateLink = prepare(sqlCreateLink)
	s.AddEvidence = prepare(sqlAddEvidence)

	if err != nil {
		return nil, err
	}

	return &s, nil
}
