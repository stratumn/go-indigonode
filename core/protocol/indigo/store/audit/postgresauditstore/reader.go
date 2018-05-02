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

	json "github.com/gibson042/canonicaljson-go"

	"github.com/stratumn/alice/core/protocol/indigo/store/audit"
	"github.com/stratumn/go-indigocore/cs"
	"github.com/stratumn/go-indigocore/types"
)

type reader struct {
	stmts readStmts
}

// GetSegment implements github.com/stratumn/go-indigocore/store.SegmentReader.GetSegment.
func (a *reader) GetSegment(ctx context.Context, linkHash *types.Bytes32) (*cs.Segment, error) {
	var segments = make(cs.SegmentSlice, 0, 1)

	rows, err := a.stmts.GetSegment.Query(linkHash[:])
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	if err = scanLinkAndEvidences(rows, &segments); err != nil {
		return nil, err
	}

	if len(segments) == 0 {
		return nil, nil
	}
	return segments[0], nil
}

func (a *reader) FindSegments(ctx context.Context, filter *audit.SegmentFilter) (cs.SegmentSlice, error) {
	var (
		rows *sql.Rows
		err  error

		segments = make(cs.SegmentSlice, 0, 1)
		limit    = filter.Top
		offset   = filter.Skip
	)

	if filter.PeerID != nil {
		rows, err = a.stmts.FindSegmentsByPeer.Query([]byte(*filter.PeerID), offset, limit)
	} else {
		rows, err = a.stmts.FindSegments.Query(offset, limit)
	}
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	err = scanLinkAndEvidences(rows, &segments)

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
			peerID       []byte
			evidenceData sql.NullString
			evidence     cs.Evidence
		)

		if err := rows.Scan(&linkHash, &linkData, &peerID, &evidenceData); err != nil {
			return err
		}

		if bytes.Compare(currentHash, linkHash) != 0 {
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
