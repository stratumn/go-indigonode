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

package db

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDiff(t *testing.T) {
	tests := []struct {
		name string
		run  func(*testing.T, DB, *Diff)
	}{{
		"put",
		func(t *testing.T, db DB, diff *Diff) {
			err := db.Put([]byte{0, 1, 2}, []byte("original"))
			assert.NoError(t, err, "db.Put(012, original)")

			err = diff.Put([]byte{0, 1, 2}, []byte("valA"))
			assert.NoError(t, err, "diff.Put(012, valA)")

			v, err := diff.Get([]byte{0, 1, 2})
			assert.NoError(t, err, "diff.Get(012)#1")
			assert.EqualValues(t, v, []byte("valA"), "diff.Get(012)#1")

			v, err = db.Get([]byte{0, 1, 2})
			assert.NoError(t, err, "db.Get(012)#1")
			assert.EqualValues(t, v, []byte("original"), "db.Get(012)#1")

			_, err = diff.Get([]byte{0})
			assert.EqualError(t, err, ErrNotFound.Error(), "diff.Get(0)")

			err = diff.Put([]byte{0, 1, 2}, []byte("valB"))
			assert.NoError(t, err, "diff.Put(012, valB)")

			v, err = diff.Get([]byte{0, 1, 2})
			assert.NoError(t, err, "diff.Get(012)#2")
			assert.EqualValues(t, v, []byte("valB"), "diff.Get(012)#2")

			v, err = db.Get([]byte{0, 1, 2})
			assert.NoError(t, err, "db.Get(012)#2")
			assert.EqualValues(t, v, []byte("original"), "db.Get(012)#2")

			assert.NoError(t, diff.Apply(), "diff.Apply()")

			v, err = db.Get([]byte{0, 1, 2})
			assert.NoError(t, err, "db.Get(012)#3")
			assert.EqualValues(t, v, []byte("valB"), "db.Get(012)#3")
		},
	}, {
		"put-delete",
		func(t *testing.T, db DB, diff *Diff) {
			alice, bob, charlie := []byte("alice"), []byte("bob"), []byte("charlie")

			err := db.Put(alice, charlie)
			assert.NoError(t, err, "db.Put(alice, charlie)")

			err = diff.Put(alice, alice)
			assert.NoError(t, err, "diff.Put(alice, alice)")

			v, err := diff.Get(alice)
			assert.NoError(t, err, "diff.Get(alice)#1")
			assert.Equal(t, v, alice, "diff.Get(alice)#1")

			err = diff.Delete(alice)
			assert.NoError(t, err, "diff.Delete(alice)#2")

			v, err = db.Get(alice)
			assert.NoError(t, err, "db.Get(alice)#1")
			assert.Equal(t, v, charlie, "db.Get(alice)#1")

			_, err = diff.Get(alice)
			assert.EqualError(t, err, ErrNotFound.Error(), "diff.Get(alice)#2")

			err = diff.Put(alice, bob)
			assert.NoError(t, err, "diff.Put(alice, bob)")

			v, err = diff.Get(alice)
			assert.NoError(t, err, "diff.Get(alice)#3")
			assert.Equal(t, v, bob, "diff.Get(alice)#3")

			assert.NoError(t, diff.Apply(), "diff.Apply()#1")

			v, err = db.Get(alice)
			assert.NoError(t, err, "db.Get(alice)#2")
			assert.Equal(t, v, bob, "db.Get(alice)#2")

			err = diff.Delete(alice)
			assert.NoError(t, err, "diff.Delete(alice)#2")

			assert.NoError(t, diff.Apply(), "diff.Apply()#2")

			_, err = db.Get(alice)
			assert.EqualError(t, err, ErrNotFound.Error(), "db.Get(alice)#3")
		},
	}, {
		"reset",
		func(t *testing.T, db DB, diff *Diff) {
			alice := []byte("alice")

			err := diff.Put(alice, alice)
			assert.NoError(t, err, "diff.Put(alice, alice)")

			v, err := diff.Get(alice)
			assert.NoError(t, err, "diff.Get(alice)#1")
			assert.Equal(t, v, alice, "diff.Get(alice)#1")

			diff.Reset()

			_, err = diff.Get(alice)
			assert.EqualError(t, err, ErrNotFound.Error(), "diff.Get(alice)#2")

			assert.NoError(t, diff.Apply(), "diff.Apply()")

			_, err = db.Get(alice)
			assert.EqualError(t, err, ErrNotFound.Error(), "db.Get(alice)")
		},
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, err := NewMemDB(nil)
			require.NoError(t, err, "NewMemDB()")
			defer db.Close()

			tt.run(t, db, NewDiff(db))
		})
	}
}
