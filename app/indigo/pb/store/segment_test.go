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
	"testing"

	"github.com/stratumn/go-indigonode/app/indigo/pb/store"
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
