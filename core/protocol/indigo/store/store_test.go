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
	"context"
	"testing"

	"github.com/stratumn/alice/core/protocol/indigo/store"
	"github.com/stratumn/go-indigocore/cs/cstesting"
	"github.com/stratumn/go-indigocore/types"
	"github.com/stretchr/testify/assert"
)

func TestCreateLink(t *testing.T) {
	tests := []struct {
		name string
		run  func(*testing.T, *store.Store)
	}{{
		"invalid-link",
		func(t *testing.T, s *store.Store) {
			link := cstesting.RandomLink()
			link.Meta.MapID = ""

			_, err := s.CreateLink(context.Background(), link)
			assert.Error(t, err, "s.CreateLink()")
		},
	}, {
		"valid-link",
		func(t *testing.T, s *store.Store) {
			link := cstesting.RandomLink()
			lh, err := s.CreateLink(context.Background(), link)
			assert.NoError(t, err, "s.CreateLink()")
			assert.NotNil(t, lh, "s.CreateLink()")
		},
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := store.New()
			tt.run(t, s)
		})
	}
}

func TestGetSegment(t *testing.T) {
	linkHashNotFound, _ := types.NewBytes32FromString("4242424242424242424242424242424242424242424242424242424242424242")

	tests := []struct {
		name string
		run  func(*testing.T, *store.Store)
	}{{
		"missing-link",
		func(t *testing.T, s *store.Store) {
			seg, err := s.GetSegment(context.Background(), linkHashNotFound)
			assert.NoError(t, err, "s.GetSegment()")
			assert.Nil(t, seg, "s.GetSegment()")
		},
	}, {
		"valid-link",
		func(t *testing.T, s *store.Store) {
			ctx := context.Background()

			link := cstesting.RandomLink()
			lh, err := s.CreateLink(ctx, link)
			assert.NoError(t, err, "s.CreateLink()")

			seg, err := s.GetSegment(ctx, lh)
			assert.NoError(t, err, "s.GetSegment()")
			assert.NotNil(t, seg, "s.GetSegment()")

			assert.Equal(t, link, &seg.Link, "seg.Link")
		},
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := store.New()
			tt.run(t, s)
		})
	}
}
