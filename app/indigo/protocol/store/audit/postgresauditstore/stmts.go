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

package postgresauditstore

import (
	"database/sql"
)

const (
	sqlCreateLink = `
		INSERT INTO audited_links (
			link_hash,
			data,
			peer_id
		)
		VALUES ($1, $2, $3)
		ON CONFLICT (link_hash)
		DO UPDATE SET
			data = $2,
			peer_id = $3
	`

	sqlFindSegments = `
		SELECT l.link_hash, l.data, l.peer_id, s.data FROM audited_links l
		LEFT JOIN audited_evidences s ON l.link_hash = s.link_hash
		ORDER BY l.created_at ASC
		OFFSET $1 LIMIT $2
	`

	sqlFindSegmentsByPeer = `
		SELECT l.link_hash, l.data, l.peer_id, s.data FROM audited_links l
		LEFT JOIN audited_evidences s ON l.link_hash = s.link_hash
		WHERE l.peer_id = $1
		ORDER BY l.created_at ASC
		OFFSET $2 LIMIT $3
	`

	sqlAddEvidence = `
		INSERT INTO audited_evidences (
			link_hash,
			data,
			peer_id
		)
		VALUES ($1, $2, $3)
		ON CONFLICT (link_hash, peer_id)
		DO NOTHING
`
)

var sqlCreate = []string{
	`
		CREATE TABLE audited_links (
			id BIGSERIAL PRIMARY KEY,
			link_hash bytea NOT NULL,
			data jsonb NOT NULL,
			peer_id text NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		)
	`,
	`
		CREATE UNIQUE INDEX audited_links_link_hash_idx
		ON audited_links (link_hash)
	`,
	`
		CREATE INDEX audited_links_peer_id_idx
		ON audited_links (peer_id)
	`,
	`
		CREATE TABLE audited_evidences (
			id BIGSERIAL PRIMARY KEY,
			link_hash bytea NOT NULL,
			data bytea NOT NULL,
			peer_id text NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
	)
	`,
	`
		CREATE INDEX audited_evidences_link_hash_idx
		ON audited_evidences (link_hash)
	`,
	`
		CREATE UNIQUE INDEX audited_evidences_link_hash_provider_idx
		ON audited_evidences (link_hash, peer_id)
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

	s.FindSegments = prepare(sqlFindSegments)
	s.FindSegmentsByPeer = prepare(sqlFindSegmentsByPeer)

	s.CreateLink = prepare(sqlCreateLink)
	s.AddEvidence = prepare(sqlAddEvidence)

	if err != nil {
		return nil, err
	}

	return &s, nil
}
