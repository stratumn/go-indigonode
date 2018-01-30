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
	assert := assert.New(t)

	tests := []struct {
		name string
		run  func(DB)
	}{{
		"get-existing-value",
		func(db DB) {
			assert.NoError(db.Put([]byte("keyA"), []byte("valA")), "db.Put(A)")
			v, err := db.Get([]byte("keyA"))
			assert.NoError(err, "db.Get(A)")
			assert.EqualValues("valA", v)
		},
	}, {
		"get-inexisting-value",
		func(db DB) {
			_, err := db.Get([]byte("keyA"))
			assert.EqualError(err, ErrNotFound.Error())
		},
	}, {
		"delete-existing-value",
		func(db DB) {
			assert.NoError(db.Put([]byte("keyA"), []byte("valA")), "db.Put(A)")
			err := db.Delete([]byte("keyA"))
			assert.NoError(err, "db.Delete(A)")
			_, err = db.Get([]byte("keyA"))
			assert.EqualError(err, ErrNotFound.Error(), "db.Get(A)")
		},
	}, {
		"delete-inexisting-value",
		func(db DB) {
			err := db.Delete([]byte("keyA"))
			assert.EqualError(err, ErrNotFound.Error())
		},
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := create(t)
			defer db.Close()

			tt.run(db)
		})
	}
}
