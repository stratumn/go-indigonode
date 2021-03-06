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

// testImplementation runs implementation-agnostic tests.
func testImplementation(t *testing.T, create func(*testing.T) DB) {
	tests := []struct {
		name string
		run  func(*testing.T, DB)
	}{{
		"get-existing-value",
		func(t *testing.T, db DB) {
			assert.NoError(t, db.Put([]byte("keyA"), []byte("valA")), "db.Put(A)")

			v, err := db.Get([]byte("keyA"))
			assert.NoError(t, err, "db.Get(A)")
			assert.EqualValues(t, "valA", v)
		},
	}, {
		"get-inexisting-value",
		func(t *testing.T, db DB) {
			_, err := db.Get([]byte("keyA"))
			assert.EqualError(t, err, ErrNotFound.Error())
		},
	}, {
		"put-overwrite",
		func(t *testing.T, db DB) {
			assert.NoError(t, db.Put([]byte("keyA"), []byte("val1")), "db.Put(A)")
			assert.NoError(t, db.Put([]byte("keyA"), []byte("val2")), "db.Put(A)")

			v, err := db.Get([]byte("keyA"))
			assert.NoError(t, err, "db.Get(A)")
			assert.EqualValues(t, "val2", v)
		},
	}, {
		"delete-existing-value",
		func(t *testing.T, db DB) {
			assert.NoError(t, db.Put([]byte("keyA"), []byte("valA")), "db.Put(A)")

			err := db.Delete([]byte("keyA"))
			assert.NoError(t, err, "db.Delete(A)")

			_, err = db.Get([]byte("keyA"))
			assert.EqualError(t, err, ErrNotFound.Error(), "db.Get(A)")
		},
	}, {
		"delete-inexisting-value",
		func(t *testing.T, db DB) {
			err := db.Delete([]byte("keyA"))
			assert.NoError(t, err, "db.Delete(A)")
		},
	}, {
		"iterate-range",
		func(t *testing.T, db DB) {
			assert.NoError(t, db.Put([]byte("keyA"), []byte("valA")), "db.Put(A)")
			assert.NoError(t, db.Put([]byte("keyBA"), []byte("valBA")), "db.Put(BA)")
			assert.NoError(t, db.Put([]byte("keyBB"), []byte("valBB")), "db.Put(BB)")
			assert.NoError(t, db.Put([]byte("keyC"), []byte("valC")), "db.Put(C)")

			iter := db.IterateRange([]byte("keyBA"), []byte("keyC"))

			next, err := iter.Next()
			require.NoError(t, err, "iter.Next()#1")
			assert.True(t, next, "iter.Next()#1")
			assert.EqualValues(t, "keyBA", iter.Key(), "iter.Key()#1")
			assert.EqualValues(t, "valBA", iter.Value(), "iter.Value()#1")

			next, err = iter.Next()
			require.NoError(t, err, "iter.Next()#2")
			assert.True(t, next, "iter.Next()#2")
			assert.EqualValues(t, "keyBB", iter.Key(), "iter.Key()#2")
			assert.EqualValues(t, "valBB", iter.Value(), "iter.Value()#2")

			next, err = iter.Next()
			require.NoError(t, err, "iter.Next()#3")
			assert.False(t, next, "iter.Next()#3")
		},
	}, {
		"iterate-prefix",
		func(t *testing.T, db DB) {
			assert.NoError(t, db.Put([]byte("keyA"), []byte("valA")), "db.Put(A)")
			assert.NoError(t, db.Put([]byte("keyBA"), []byte("valBA")), "db.Put(BA)")
			assert.NoError(t, db.Put([]byte("keyBB"), []byte("valBB")), "db.Put(BB)")
			assert.NoError(t, db.Put([]byte("keyC"), []byte("valC")), "db.Put(C)")

			iter := db.IteratePrefix([]byte("keyB"))

			next, err := iter.Next()
			require.NoError(t, err, "iter.Next()#1")
			assert.True(t, next, "iter.Next()#1")
			assert.EqualValues(t, "keyBA", iter.Key(), "iter.Key()#1")
			assert.EqualValues(t, "valBA", iter.Value(), "iter.Value()#1")

			next, err = iter.Next()
			require.NoError(t, err, "iter.Next()#2")
			assert.True(t, next, "iter.Next()#2")
			assert.EqualValues(t, "keyBB", iter.Key(), "iter.Key()#2")
			assert.EqualValues(t, "valBB", iter.Value(), "iter.Value()#2")

			next, err = iter.Next()
			require.NoError(t, err, "iter.Next()#3")
			assert.False(t, next, "iter.Next()#3")
		},
	}, {
		"write",
		func(t *testing.T, db DB) {
			b := db.Batch()
			b.Put([]byte("keyA"), []byte("valA"))
			b.Put([]byte("keyB"), []byte("valB"))

			// Make sure DB is not updated yet.
			_, err := db.Get([]byte("keyA"))
			assert.EqualError(t, err, ErrNotFound.Error(), "db.Get(A)")
			_, err = db.Get([]byte("keyB"))
			assert.EqualError(t, err, ErrNotFound.Error(), "db.Get(B)")

			assert.NoError(t, db.Write(b))

			// Make sure DB was updated.
			v, err := db.Get([]byte("keyA"))
			assert.NoError(t, err, "db.Get(A)")
			assert.EqualValues(t, "valA", v, "db.Get(A)")

			v, err = db.Get([]byte("keyB"))
			assert.NoError(t, err, "db.Get(B)")
			assert.EqualValues(t, "valB", v, "db.Get(B)")
		},
	}, {
		"transaction-commit",
		func(t *testing.T, db DB) {
			tx, err := db.Transaction()
			assert.NoError(t, err, "db.Transaction()")
			assert.NoError(t, tx.Put([]byte("keyA"), []byte("valA")), "tx.Put(A)")
			assert.NoError(t, tx.Put([]byte("keyB"), []byte("valB")), "tx.Put(B)")

			// Make sure transaction is up to date.
			v, err := tx.Get([]byte("keyA"))
			assert.NoError(t, err, "tx.Get(A)")
			assert.EqualValues(t, "valA", v, "tx.Get(A)")

			assert.NoError(t, tx.Delete([]byte("keyA")), "tx.Delete(A)")

			// Make sure DB is not updated yet.
			_, err = db.Get([]byte("keyA"))
			assert.EqualError(t, err, ErrNotFound.Error(), "db.Get(A)")
			_, err = db.Get([]byte("keyB"))
			assert.EqualError(t, err, ErrNotFound.Error(), "db.Get(B)")

			assert.NoError(t, tx.Commit())

			// Make sure DB was updated propertly.
			_, err = db.Get([]byte("keyA"))
			assert.EqualError(t, err, ErrNotFound.Error(), "db.Get(A)")

			v, err = db.Get([]byte("keyB"))
			assert.NoError(t, err, "db.Get(B)")
			assert.EqualValues(t, "valB", v, "db.Get(B)")
		},
	}, {
		"transaction-discard",
		func(t *testing.T, db DB) {
			tx, err := db.Transaction()
			assert.NoError(t, err, "db.Transaction()")
			assert.NoError(t, tx.Put([]byte("keyA"), []byte("valA")), "tx.Put(A)")
			tx.Discard()

			// Make sure DB wasn't updated.
			_, err = db.Get([]byte("keyA"))
			assert.EqualError(t, err, ErrNotFound.Error(), "db.Get(A)")

			_, err = db.Get([]byte("keyB"))
			assert.EqualError(t, err, ErrNotFound.Error(), "db.Get(B)")
		},
	}, {
		"transaction-iterate-range",
		func(t *testing.T, db DB) {
			tx, err := db.Transaction()
			assert.NoError(t, err, "db.Transaction()")
			defer tx.Discard()

			assert.NoError(t, tx.Put([]byte("keyA"), []byte("valA")), "tx.Put(A)")
			assert.NoError(t, tx.Put([]byte("keyBA"), []byte("valBA")), "tx.Put(BA)")
			assert.NoError(t, tx.Put([]byte("keyBB"), []byte("valBB")), "tx.Put(BB)")
			assert.NoError(t, tx.Put([]byte("keyC"), []byte("valC")), "tx.Put(C)")

			iter := tx.IterateRange([]byte("keyB"), []byte("keyC"))

			next, err := iter.Next()
			require.NoError(t, err, "iter.Next()#1")
			assert.True(t, next, "iter.Next()#1")
			assert.EqualValues(t, "keyBA", iter.Key(), "iter.Key()#1")
			assert.EqualValues(t, "valBA", iter.Value(), "iter.Value()#1")

			next, err = iter.Next()
			require.NoError(t, err, "iter.Next()#2")
			assert.True(t, next, "iter.Next()#2")
			assert.EqualValues(t, "keyBB", iter.Key(), "iter.Key()#2")
			assert.EqualValues(t, "valBB", iter.Value(), "iter.Value()#2")

			next, err = iter.Next()
			require.NoError(t, err, "iter.Next()#3")
			assert.False(t, next, "iter.Next()#3")
		},
	}, {
		"transaction-iterate-prefix",
		func(t *testing.T, db DB) {
			tx, err := db.Transaction()
			assert.NoError(t, err, "db.Transaction()")
			defer tx.Discard()

			assert.NoError(t, tx.Put([]byte("keyA"), []byte("valA")), "tx.Put(A)")
			assert.NoError(t, tx.Put([]byte("keyBA"), []byte("valBA")), "tx.Put(BA)")
			assert.NoError(t, tx.Put([]byte("keyBB"), []byte("valBB")), "tx.Put(BB)")
			assert.NoError(t, tx.Put([]byte("keyC"), []byte("valC")), "tx.Put(C)")

			iter := tx.IteratePrefix([]byte("keyB"))

			next, err := iter.Next()
			require.NoError(t, err, "iter.Next()#1")
			assert.True(t, next, "iter.Next()#1")
			assert.EqualValues(t, "keyBA", iter.Key(), "iter.Key()#1")
			assert.EqualValues(t, "valBA", iter.Value(), "iter.Value()#1")

			next, err = iter.Next()
			require.NoError(t, err, "iter.Next()#2")
			assert.True(t, next, "iter.Next()#2")
			assert.EqualValues(t, "keyBB", iter.Key(), "iter.Key()#2")
			assert.EqualValues(t, "valBB", iter.Value(), "iter.Value()#2")

			next, err = iter.Next()
			require.NoError(t, err, "iter.Next()#3")
			assert.False(t, next, "iter.Next()#3")
		},
	}, {
		"transaction-batch",
		func(t *testing.T, db DB) {
			tx, err := db.Transaction()
			assert.NoError(t, err, "db.Transaction()")

			b := tx.Batch()
			b.Put([]byte("keyA"), []byte("valA"))
			b.Put([]byte("keyB"), []byte("valB"))
			assert.NoError(t, tx.Write(b), "tx.Write()")

			assert.NoError(t, tx.Commit(), "tx.Commit()")

			// Make sure DB was updated.
			v, err := db.Get([]byte("keyA"))
			assert.NoError(t, err, "db.Get(A)")
			assert.EqualValues(t, "valA", v, "db.Get(A)")

			v, err = db.Get([]byte("keyB"))
			assert.NoError(t, err, "db.Get(B)")
			assert.EqualValues(t, "valB", v, "db.Get(B)")
		},
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := create(t)
			defer db.Close()

			tt.run(t, db)
		})
	}
}
