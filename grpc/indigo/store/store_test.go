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

package store_test

import (
	"encoding/json"
	"testing"

	"github.com/stratumn/alice/grpc/indigo/store"
	pb "github.com/stratumn/alice/pb/indigo/store"
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
