// Copyright © 2017-2018  Stratumn SAS
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
			assert.EqualError(t, err, ErrNotFound.Error())
		},
	}, {
		"batch-commit",
		func(t *testing.T, db DB) {
			b := db.Batch()
			assert.NoError(t, b.Put([]byte("keyA"), []byte("valA")), "b.Put(A)")
			assert.NoError(t, b.Put([]byte("keyB"), []byte("valB")), "b.Put(B)")
			assert.NoError(t, b.Commit())

			// Make sure DB was updated.
			v, err := db.Get([]byte("keyA"))
			assert.NoError(t, err, "db.Get(A)")
			assert.EqualValues(t, "valA", v)
			v, err = db.Get([]byte("keyB"))
			assert.NoError(t, err, "db.Get(B)")
			assert.EqualValues(t, "valB", v)
		},
	}, {
		"batch-commit-error",
		func(t *testing.T, db DB) {
			b := db.Batch()
			assert.NoError(t, b.Put([]byte("keyA"), []byte("valA")), "b.Put(A)")
			assert.NoError(t, b.Delete([]byte("keyB")), "b.Delete(B)")
			assert.NoError(t, b.Put([]byte("keyB"), []byte("valB")), "b.Put(B)")
			assert.EqualError(t, b.Commit(), ErrNotFound.Error())

			// Make sure DB didn't changed.
			_, err := db.Get([]byte("keyA"))
			assert.EqualError(t, err, ErrNotFound.Error(), "db.Get(A)")
			_, err = db.Get([]byte("keyB"))
			assert.EqualError(t, err, ErrNotFound.Error(), "db.Get(B)")
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
