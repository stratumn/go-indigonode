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
	"encoding/json"

	pb "github.com/stratumn/go-indigonode/app/indigo/pb/store"
	"github.com/stratumn/go-indigocore/cs"
	indigostore "github.com/stratumn/go-indigocore/store"
	"github.com/stratumn/go-indigocore/types"
)

// ToSegmentFilter converts to the Indigo-core type.
func (f *SegmentFilter) ToSegmentFilter() (*indigostore.SegmentFilter, error) {
	if f == nil {
		return nil, pb.ErrInvalidArgument
	}

	var filter indigostore.SegmentFilter
	err := json.Unmarshal(f.Data, &filter)
	if err != nil {
		return nil, pb.ErrInvalidArgument
	}
	if filter.Pagination.Limit == 0 {
		filter.Pagination.Limit = indigostore.DefaultLimit
	}

	return &filter, nil
}

// ToMapFilter converts to the Indigo-core type.
func (f *MapFilter) ToMapFilter() (*indigostore.MapFilter, error) {
	if f == nil {
		return nil, pb.ErrInvalidArgument
	}

	var filter indigostore.MapFilter
	err := json.Unmarshal(f.Data, &filter)
	if err != nil {
		return nil, pb.ErrInvalidArgument
	}
	if filter.Pagination.Limit == 0 {
		filter.Pagination.Limit = indigostore.DefaultLimit
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
		return nil, nil, pb.ErrInvalidArgument
	}

	lh, err := r.LinkHash.ToLinkHash()
	if err != nil {
		return nil, nil, pb.ErrInvalidArgument
	}

	var e cs.Evidence
	err = json.Unmarshal(r.Evidence.Data, &e)
	if err != nil {
		return nil, nil, pb.ErrInvalidArgument
	}

	return lh, &e, nil
}
