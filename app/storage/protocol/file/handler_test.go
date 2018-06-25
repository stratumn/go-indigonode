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

package file_test

import (
	"context"
	"crypto/sha256"
	"fmt"
	mh "gx/ipfs/QmZyZDi491cCNTLfAhwcaDii2Kg4pwKRkhqQzURGDvY6ua/go-multihash"
	"io/ioutil"
	"path"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/satori/go.uuid"
	"github.com/stretchr/testify/require"

	"github.com/stratumn/go-indigonode/app/storage/protocol/file"
	"github.com/stratumn/go-indigonode/app/storage/protocol/file/mockhandler"
	"github.com/stratumn/go-indigonode/core/db"
	"github.com/stretchr/testify/assert"
)

var (
	tmpStoragePath string
)

func init() {
	var err error
	tmpStoragePath, err = ioutil.TempDir("", "filetest")
	if err != nil {
		panic(err)
	}
}

func TestFileHandler_Write(t *testing.T) {
	db, err := db.NewMemDB(nil)
	require.NoError(t, err, "NewMemDB()")
	fileHandler := file.NewLocalFileHandler(tmpStoragePath, db)
	ctx := context.Background()
	fileName := fmt.Sprintf("TestFileHandler_BeginWrite-%d", time.Now().UnixNano())
	var id uuid.UUID
	var hash []byte

	t.Run("BeginWrite", func(t *testing.T) {
		var err error
		id, err = fileHandler.BeginWrite(ctx, fileName)
		assert.NoError(t, err, "BeginWrite")
	})

	t.Run("BeginWrite_Fail", func(t *testing.T) {
		_, err := fileHandler.BeginWrite(ctx, "")
		assert.EqualError(t, err, file.ErrFileNameMissing.Error(), "BeginWrite")
	})

	t.Run("WriteChunk", func(t *testing.T) {
		chunk := []byte(" some data")
		err = fileHandler.WriteChunk(ctx, id, chunk)
		assert.NoError(t, err, "WriteChunk")
	})

	t.Run("WriteChunk_Fail", func(t *testing.T) {
		t.Run("no-session", func(t *testing.T) {
			err = fileHandler.WriteChunk(ctx, uuid.NewV4(), []byte(" some data"))
			assert.EqualError(t, err, file.ErrNoSession.Error(), "WriteChunk")
		})

		t.Run("fail-write-and-delete", func(t *testing.T) {
			fileName := fmt.Sprintf("TestFileHandler_BeginWrite-%d", time.Now().UnixNano())
			id2, err := fileHandler.BeginWrite(ctx, fileName)
			assert.NoError(t, err, "BeginWrite")

			// close file
			_, err = fileHandler.EndWrite(ctx, id2)
			assert.NoError(t, err, "EndWrite")

			err = fileHandler.WriteChunk(ctx, id2, []byte("yo"))
			assert.Error(t, err, "WriteChunk")

			// Check that session has been deleted.
			err = fileHandler.WriteChunk(ctx, id2, []byte("yo"))
			assert.EqualError(t, err, file.ErrNoSession.Error(), "WriteChunk")
		})
	})

	t.Run("EndWrite", func(t *testing.T) {
		var err error
		hash, err = fileHandler.EndWrite(ctx, id)
		require.NoError(t, err, "EndWrite()")

		// Check that hash is correct.
		b, err := fileHandler.Read(ctx, hash)
		assert.NoError(t, err)

		sha := sha256.Sum256(b)
		expected, err := mh.Encode(sha[:], mh.SHA2_256)
		assert.NoError(t, err, "mh.Encode")
		assert.Equal(t, expected, hash, "file hash incorrect")
	})

	t.Run("Exists", func(t *testing.T) {
		exists, err := fileHandler.Exists(ctx, hash)
		assert.NoError(t, err, "Exists(hash)")
		assert.True(t, exists, "Exists(hash)")
	})

	t.Run("Exists_False", func(t *testing.T) {
		exists, err := fileHandler.Exists(ctx, []byte("123"))
		assert.NoError(t, err, "Exists(hash)")
		assert.False(t, exists, "Exists(hash)")
	})

	t.Run("EndWrite_fail", func(t *testing.T) {
		t.Run("no-session", func(t *testing.T) {
			_, err = fileHandler.EndWrite(ctx, uuid.NewV4())
			assert.EqualError(t, err, file.ErrNoSession.Error(), "WriteChunk")
		})

		t.Run("fail-and-delete-file", func(t *testing.T) {
			id3, err := fileHandler.BeginWrite(ctx, fileName)
			assert.NoError(t, err, "BeginWrite")

			err = db.Close()
			assert.NoError(t, err, "db.Close()")

			_, err = fileHandler.EndWrite(ctx, id3)
			assert.Error(t, err, "WriteChunk")

			// Check that session has been deleted.
			err = fileHandler.WriteChunk(ctx, id3, []byte("yo"))
			assert.EqualError(t, err, file.ErrNoSession.Error(), "WriteChunk")
		})
	})
}

func TestFileHandler_ReadChunks(t *testing.T) {
	ctx := context.Background()
	db, err := db.NewMemDB(nil)
	require.NoError(t, err, "NewMemDB()")
	fileHandler := file.NewLocalFileHandler(tmpStoragePath, db)
	fileName := fmt.Sprintf("TestFileHandler_BeginWrite-%d", time.Now().UnixNano())

	id, err := fileHandler.BeginWrite(context.Background(), fileName)
	assert.NoError(t, err, "BeginWrite")
	err = fileHandler.WriteChunk(ctx, id, []byte("who wants to download "+"my juicy file ?"))
	assert.NoError(t, err, "WriteChunk")
	fileHash, err := fileHandler.EndWrite(ctx, id)
	assert.NoError(t, err, "EndWrite")

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	reader := mockhandler.NewMockReader(ctrl)

	gomock.InOrder(
		reader.EXPECT().OnChunk([]byte("who wants to download "), path.Join(tmpStoragePath, fileName)),
		reader.EXPECT().OnChunk([]byte("my juicy file ?"), path.Join(tmpStoragePath, fileName)),
	)

	err = fileHandler.ReadChunks(context.Background(), fileHash, 22, reader)
	assert.NoError(t, err, "SendFile")
}
