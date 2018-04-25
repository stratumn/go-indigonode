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
	"github.com/stratumn/go-indigocore/cs"
	"github.com/stratumn/go-indigocore/cs/cstesting"
	indigostore "github.com/stratumn/go-indigocore/store"
	"github.com/stratumn/go-indigocore/types"
	"github.com/stretchr/testify/assert"
)

func TestLinkHash(t *testing.T) {
	var lh *store.LinkHash
	_, err := lh.ToLinkHash()
	assert.EqualError(t, err, store.ErrInvalidArgument.Error())

	linkHash := types.NewBytes32FromBytes([]byte{0x24, 0x42})
	lh = store.FromLinkHash(linkHash)
	assert.Equal(t, linkHash[:], lh.Data)

	linkHash2, err := lh.ToLinkHash()
	assert.NoError(t, err)
	assert.Equal(t, linkHash, linkHash2)
}

func TestLink(t *testing.T) {
	var l *store.Link
	_, err := l.ToLink()
	assert.EqualError(t, err, store.ErrInvalidArgument.Error())

	link := cstesting.RandomLink()
	l, err = store.FromLink(link)
	assert.NoError(t, err)

	link2, err := l.ToLink()
	assert.NoError(t, err)
	assert.Equal(t, link, link2)
}

func TestSegments(t *testing.T) {
	s1 := cstesting.RandomSegment()
	s2 := cstesting.RandomSegment()

	s, err := store.FromSegments(cs.SegmentSlice{s1, s2})
	assert.NoError(t, err)

	assert.Len(t, s.Segments, 2)
	assert.NotNil(t, s.Segments[0].Data)
	assert.NotNil(t, s.Segments[1].Data)

	ss1, err := s.Segments[0].ToSegment()
	assert.NoError(t, err)
	assert.Equal(t, s1, ss1)

	ss2, err := s.Segments[1].ToSegment()
	assert.NoError(t, err)
	assert.Equal(t, s2, ss2)
}

func TestSegmentFilter(t *testing.T) {
	var f *store.SegmentFilter
	_, err := f.ToSegmentFilter()
	assert.EqualError(t, err, store.ErrInvalidArgument.Error())

	filter := &indigostore.SegmentFilter{Process: "p"}
	filterBytes, _ := json.Marshal(filter)
	f = &store.SegmentFilter{Data: filterBytes}

	filter2, err := f.ToSegmentFilter()
	assert.NoError(t, err)
	assert.Equal(t, filter, filter2)
}

func TestMapFilter(t *testing.T) {
	var f *store.MapFilter
	_, err := f.ToMapFilter()
	assert.EqualError(t, err, store.ErrInvalidArgument.Error())

	filter := &indigostore.MapFilter{Process: "p"}
	filterBytes, _ := json.Marshal(filter)
	f = &store.MapFilter{Data: filterBytes}

	filter2, err := f.ToMapFilter()
	assert.NoError(t, err)
	assert.Equal(t, filter, filter2)
}

func TestAddEvidence(t *testing.T) {
	var req *store.AddEvidenceReq
	_, _, err := req.ToAddEvidenceParams()
	assert.EqualError(t, err, store.ErrInvalidArgument.Error())

	req = &store.AddEvidenceReq{}
	_, _, err = req.ToAddEvidenceParams()
	assert.EqualError(t, err, store.ErrInvalidArgument.Error())

	e := cstesting.RandomEvidence()
	eBytes, _ := json.Marshal(e)

	req = &store.AddEvidenceReq{
		LinkHash: &store.LinkHash{Data: []byte{0x24, 0x42}},
		Evidence: &store.Evidence{Data: eBytes},
	}

	lh, e2, err := req.ToAddEvidenceParams()
	assert.NoError(t, err)
	assert.Equal(t, byte(0x24), lh[0])
	assert.Equal(t, byte(0x42), lh[1])
	assert.Equal(t, e.Provider, e2.Provider)
}

func TestEvidences(t *testing.T) {
	e1 := cstesting.RandomEvidence()
	e2 := cstesting.RandomEvidence()

	e, err := store.FromEvidences(cs.Evidences{e1, e2})
	assert.NoError(t, err)

	assert.Len(t, e.Evidences, 2)
	assert.NotNil(t, e.Evidences[0].Data)
	assert.NotNil(t, e.Evidences[1].Data)
}
