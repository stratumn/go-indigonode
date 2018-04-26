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
	"testing"

	"github.com/stratumn/alice/pb/indigo/store"
	"github.com/stratumn/go-indigocore/cs"
	"github.com/stratumn/go-indigocore/cs/cstesting"
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

func TestEvidences(t *testing.T) {
	e1 := cstesting.RandomEvidence()
	e2 := cstesting.RandomEvidence()

	e, err := store.FromEvidences(cs.Evidences{e1, e2})
	assert.NoError(t, err)

	assert.Len(t, e.Evidences, 2)
	assert.NotNil(t, e.Evidences[0].Data)
	assert.NotNil(t, e.Evidences[1].Data)
}
