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

	pb "github.com/stratumn/alice/pb/indigo/store"
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
