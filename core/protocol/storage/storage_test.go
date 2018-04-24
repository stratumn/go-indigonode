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

package storage

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/stratumn/alice/core/db"
	"github.com/stretchr/testify/assert"

	multihash "gx/ipfs/QmZyZDi491cCNTLfAhwcaDii2Kg4pwKRkhqQzURGDvY6ua/go-multihash"
)

func TestStorageProtocol_IndexFile(t *testing.T) {
	db, err := db.NewMemDB(nil)
	assert.NoError(t, err, "NewMemDB()")
	storage := &Storage{
		db: db,
	}

	filepath := fmt.Sprintf("/tmp/storage_test_%d", time.Now().UnixNano())

	file, err := os.Create(filepath)
	assert.NoError(t, err, "Create()")

	content := []byte("I love storing stuff")
	hash, err := multihash.Sum(content, multihash.SHA2_256, -1)
	assert.NoError(t, err, "Sum()")
	_, err = file.Write(content)
	assert.NoError(t, err, "file.Write()")

	fh, err := storage.IndexFile(context.Background(), file)
	assert.NoError(t, err, "IndexFile")
	assert.Equal(t, []byte(hash), fh, "IndexFile")

	// Check that the file name and hash are in the db.
	got, err := db.Get(append(prefixFilesHashes, []byte(fh)...))
	assert.NoError(t, err, "db.Get()")
	assert.Equal(t, []byte(file.Name()), got, "hash")
}
