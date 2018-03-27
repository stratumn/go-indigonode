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
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/stratumn/alice/core/db"
	"github.com/stretchr/testify/assert"

	peer "gx/ipfs/QmZoWKhxUmZ2seW4BzX6fJkNR8hh9PsGModr7q171yq2SS/go-libp2p-peer"
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
	_, err = file.Write(content)
	assert.NoError(t, err, "file.Write()")

	err = storage.IndexFile(context.Background(), file)
	assert.NoError(t, err, "IndexFile")

	// Check that the file name and hash are in the db.
	got, err := db.Get(append(prefixFilesHashes, []byte(filepath)...))
	assert.NoError(t, err, "db.Get()")
	assert.Equal(t, []byte(hash), got, "hash")
}

func TestStorageProtocol_Authorize(t *testing.T) {
	db, err := db.NewMemDB(nil)
	assert.NoError(t, err, "NewMemDB()")
	storage := &Storage{
		db: db,
	}
	hash := []byte("yolo")

	t.Run("no-entry", func(t *testing.T) {
		pid1, err := peer.IDB58Decode("QmVhJVRSYHNSHgR9dJNbDxu6G7GPPqJAeiJoVRvcexGN11")
		assert.NoError(t, err, "IDB58Decode")
		pid2, err := peer.IDB58Decode("QmVhJVRSYHNSHgR9dJNbDxu6G7GPPqJAeiJoVRvcexGN22")
		assert.NoError(t, err, "IDB58Decode")
		pid3, err := peer.IDB58Decode("QmVhJVRSYHNSHgR9dJNbDxu6G7GPPqJAeiJoVRvcexGN33")
		assert.NoError(t, err, "IDB58Decode")
		pids := [][]byte{[]byte(pid1), []byte(pid2), []byte(pid3)}

		err = storage.Authorize(context.Background(), pids, hash)
		assert.NoError(t, err, "Authorize()")

		e, err := db.Get(append(prefixAuthorizedPeers, hash...))
		assert.NoError(t, err, "db.Get")

		aps := authorizedPeersMap{}
		json.Unmarshal(e, &aps)

		assert.Len(t, aps, len(pids), "len(aps")
		for _, p := range pids {
			pid, err := peer.IDFromBytes(p)
			assert.NoError(t, err, "peer.IDFromBytes()")
			assert.Equal(t, struct{}{}, aps[pid], "AuthorizedPeersMap[pid]")
		}
	})
}
