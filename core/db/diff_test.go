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
