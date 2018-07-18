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

package store

import (
	"errors"

	json "github.com/gibson042/canonicaljson-go"
	"github.com/stratumn/go-indigocore/cs"
	"github.com/stratumn/go-indigocore/types"
)

var (
	// ErrInvalidArgument is returned when the input is invalid.
	ErrInvalidArgument = errors.New("invalid argument")
)

// ToLinkHash converts to the Indigo-core type.
func (lh *LinkHash) ToLinkHash() (*types.Bytes32, error) {
	if lh == nil {
		return nil, ErrInvalidArgument
	}

	hash := types.NewBytes32FromBytes(lh.Data)
	return hash, nil
}

// FromLinkHash converts from the Indigo-core type.
func FromLinkHash(lh *types.Bytes32) *LinkHash {
	return &LinkHash{Data: lh[:]}
}

// ToLinkHashes converts to the Indigo-core type.
func (lhs *LinkHashes) ToLinkHashes() ([]string, error) {
	if lhs == nil {
		return nil, ErrInvalidArgument
	}

	res := make([]string, len(lhs.LinkHashes))
	for i, lh := range lhs.LinkHashes {
		parsed, err := lh.ToLinkHash()
		if err != nil {
			return nil, err
		}

		res[i] = parsed.String()
	}

	return res, nil
}

// FromLinkHashes converts from the Indigo-core type.
func FromLinkHashes(lhs []string) *LinkHashes {
	res := &LinkHashes{}

	for _, lh := range lhs {
		linkHash, _ := types.NewBytes32FromString(lh)
		res.LinkHashes = append(res.LinkHashes, FromLinkHash(linkHash))
	}

	return res
}

// ToLink converts to the Indigo-core type.
func (l *Link) ToLink() (*cs.Link, error) {
	if l == nil {
		return nil, ErrInvalidArgument
	}

	var link cs.Link
	err := json.Unmarshal(l.Data, &link)
	if err != nil {
		return nil, ErrInvalidArgument
	}

	return &link, nil
}

// FromLink converts from the Indigo-core type.
func FromLink(l *cs.Link) (*Link, error) {
	linkBytes, err := json.Marshal(l)
	if err != nil {
		return nil, err
	}

	return &Link{Data: linkBytes}, nil
}

// ToSegment converts to the Indigo-core type.
func (s *Segment) ToSegment() (*cs.Segment, error) {
	if s == nil {
		return nil, ErrInvalidArgument
	}

	var segment cs.Segment
	err := json.Unmarshal(s.Data, &segment)
	if err != nil {
		return nil, ErrInvalidArgument
	}

	return &segment, nil
}

// FromSegment converts from the Indigo-core type.
func FromSegment(s *cs.Segment) (*Segment, error) {
	segmentBytes, err := json.Marshal(s)
	if err != nil {
		return nil, err
	}

	return &Segment{Data: segmentBytes}, nil
}

// FromSegments converts from the Indigo-core type.
func FromSegments(segments cs.SegmentSlice) (*Segments, error) {
	res := &Segments{}
	res.Segments = make([]*Segment, len(segments))

	for i := 0; i < len(segments); i++ {
		s, err := FromSegment(segments[i])
		if err != nil {
			return nil, err
		}

		res.Segments[i] = s
	}

	return res, nil
}

// FromEvidence converts from the Indigo-core type.
func FromEvidence(e *cs.Evidence) (*Evidence, error) {
	evidenceBytes, err := json.Marshal(e)
	if err != nil {
		return nil, ErrInvalidArgument
	}

	return &Evidence{Data: evidenceBytes}, nil
}

// FromEvidences converts from the Indigo-core type.
func FromEvidences(evidences cs.Evidences) (*Evidences, error) {
	res := &Evidences{}
	res.Evidences = make([]*Evidence, len(evidences))

	for i := 0; i < len(evidences); i++ {
		e, err := FromEvidence(evidences[i])
		if err != nil {
			return nil, err
		}

		res.Evidences[i] = e
	}

	return res, nil
}
