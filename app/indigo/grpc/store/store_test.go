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

package store_test

import (
	"encoding/json"
	"testing"

	"github.com/stratumn/go-node/app/indigo/grpc/store"
	pb "github.com/stratumn/go-node/app/indigo/pb/store"
	"github.com/stratumn/go-indigocore/cs/cstesting"
	indigostore "github.com/stratumn/go-indigocore/store"
	"github.com/stretchr/testify/assert"
)

func TestSegmentFilter(t *testing.T) {
	var f *store.SegmentFilter
	_, err := f.ToSegmentFilter()
	assert.EqualError(t, err, pb.ErrInvalidArgument.Error())

	filter := &indigostore.SegmentFilter{Process: "p", Pagination: indigostore.Pagination{Limit: 10}}
	filterBytes, _ := json.Marshal(filter)
	f = &store.SegmentFilter{Data: filterBytes}

	filter2, err := f.ToSegmentFilter()
	assert.NoError(t, err)
	assert.Equal(t, filter, filter2)
}

func TestMapFilter(t *testing.T) {
	var f *store.MapFilter
	_, err := f.ToMapFilter()
	assert.EqualError(t, err, pb.ErrInvalidArgument.Error())

	filter := &indigostore.MapFilter{Process: "p", Pagination: indigostore.Pagination{Limit: 10}}
	filterBytes, _ := json.Marshal(filter)
	f = &store.MapFilter{Data: filterBytes}

	filter2, err := f.ToMapFilter()
	assert.NoError(t, err)
	assert.Equal(t, filter, filter2)
}

func TestAddEvidence(t *testing.T) {
	var req *store.AddEvidenceReq
	_, _, err := req.ToAddEvidenceParams()
	assert.EqualError(t, err, pb.ErrInvalidArgument.Error())

	req = &store.AddEvidenceReq{}
	_, _, err = req.ToAddEvidenceParams()
	assert.EqualError(t, err, pb.ErrInvalidArgument.Error())

	e := cstesting.RandomEvidence()
	eBytes, _ := json.Marshal(e)

	req = &store.AddEvidenceReq{
		LinkHash: &pb.LinkHash{Data: []byte{0x24, 0x42}},
		Evidence: &pb.Evidence{Data: eBytes},
	}

	lh, e2, err := req.ToAddEvidenceParams()
	assert.NoError(t, err)
	assert.Equal(t, byte(0x24), lh[0])
	assert.Equal(t, byte(0x42), lh[1])
	assert.Equal(t, e.Provider, e2.Provider)
}
