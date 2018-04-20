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

package store

import (
	"encoding/json"
	"errors"

	"github.com/stratumn/go-indigocore/cs"
	indigostore "github.com/stratumn/go-indigocore/store"
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

// FromSegment converts from the Indigo-core type.
func FromSegment(s *cs.Segment) (*Segment, error) {
	segmentBytes, err := json.Marshal(s)
	if err != nil {
		return nil, err
	}

	return &Segment{Data: segmentBytes}, nil
}

// ToSegmentFilter converts to the Indigo-core type.
func (f *SegmentFilter) ToSegmentFilter() (*indigostore.SegmentFilter, error) {
	if f == nil {
		return nil, ErrInvalidArgument
	}

	var filter indigostore.SegmentFilter
	err := json.Unmarshal(f.Data, &filter)
	if err != nil {
		return nil, ErrInvalidArgument
	}

	return &filter, nil
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

// ToMapFilter converts to the Indigo-core type.
func (f *MapFilter) ToMapFilter() (*indigostore.MapFilter, error) {
	if f == nil {
		return nil, ErrInvalidArgument
	}

	var filter indigostore.MapFilter
	err := json.Unmarshal(f.Data, &filter)
	if err != nil {
		return nil, ErrInvalidArgument
	}

	return &filter, nil
}

// FromMapIDs converts from the Indigo-core type.
func FromMapIDs(mapIDs []string) (*MapIDs, error) {
	return &MapIDs{MapIds: mapIDs}, nil
}

// ToAddEvidenceParams converts to the Indigo-core type.
func (r *AddEvidenceReq) ToAddEvidenceParams() (*types.Bytes32, *cs.Evidence, error) {
	if r == nil || r.Evidence == nil {
		return nil, nil, ErrInvalidArgument
	}

	lh, err := r.LinkHash.ToLinkHash()
	if err != nil {
		return nil, nil, ErrInvalidArgument
	}

	var e cs.Evidence
	err = json.Unmarshal(r.Evidence.Data, &e)
	if err != nil {
		return nil, nil, ErrInvalidArgument
	}

	return lh, &e, nil
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
