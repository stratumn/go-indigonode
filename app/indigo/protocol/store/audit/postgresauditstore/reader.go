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

package postgresauditstore

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"

	"github.com/stratumn/go-indigonode/app/indigo/protocol/store/audit"
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
