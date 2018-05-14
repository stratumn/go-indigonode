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
	"bytes"
	"context"
	"database/sql"
	"encoding/json"

	"github.com/stratumn/alice/core/protocol/indigo/store/audit"
	"github.com/stratumn/go-indigocore/cs"

	peer "gx/ipfs/QmcJukH2sAFjY3HdBKq35WDzWoL3UUu2gt9wdfqZTUyM74/go-libp2p-peer"
)

type reader struct {
	stmts readStmts
}

func (a *reader) FindSegments(ctx context.Context, filter *audit.SegmentFilter) (cs.SegmentSlice, error) {
	var (
		rows *sql.Rows
		err  error

		segments = make(cs.SegmentSlice, 0, 1)
		limit    = filter.Top
		offset   = filter.Skip
	)

	if filter.PeerID != "" {
		rows, err = a.stmts.FindSegmentsByPeer.Query(filter.PeerID.Pretty(), offset, limit)
	} else {
		rows, err = a.stmts.FindSegments.Query(offset, limit)
	}
	if err != nil {
		return nil, err
	}

	defer func() { err = rows.Close() }()

	if err = scanLinkAndEvidences(rows, &segments); err != nil {
		return nil, err
	}

	return segments, err
}

func scanLinkAndEvidences(rows *sql.Rows, segments *cs.SegmentSlice) error {
	var currentSegment *cs.Segment
	var currentHash []byte

	for rows.Next() {
		var (
			linkHash     []byte
			linkData     string
			link         cs.Link
			peerID       peer.ID
			evidenceData sql.NullString
			evidence     cs.Evidence
		)

		if err := rows.Scan(&linkHash, &linkData, &peerID, &evidenceData); err != nil {
			return err
		}

		if !bytes.Equal(currentHash, linkHash) {
			if err := json.Unmarshal([]byte(linkData), &link); err != nil {
				return err
			}

			hash, err := link.Hash()
			if err != nil {
				return err
			}
			currentHash = hash[:]

			currentSegment = link.Segmentify()

			*segments = append(*segments, currentSegment)
		}

		if evidenceData.Valid {
			if err := json.Unmarshal([]byte(evidenceData.String), &evidence); err != nil {
				return err
			}

			if err := currentSegment.Meta.AddEvidence(evidence); err != nil {
				return err
			}
		}
	}
	return nil
}
