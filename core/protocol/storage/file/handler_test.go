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

package file

import (
	"context"
	"crypto/sha256"
	"fmt"
	mh "gx/ipfs/QmZyZDi491cCNTLfAhwcaDii2Kg4pwKRkhqQzURGDvY6ua/go-multihash"
	"io"
	"os"
	"testing"
	"time"

	"github.com/satori/go.uuid"
	"github.com/stretchr/testify/require"

	"github.com/stratumn/alice/core/db"
	"github.com/stretchr/testify/assert"
)

// ============================================================================
// ==== 															Write																 ====
// ============================================================================

func TestFileHandler_BeginWrite(t *testing.T) {

	fileHandler := &localFileHandler{
		storagePath:   "/tmp",
		writeSessions: make(map[uuid.UUID]*session),
	}

	fileName := fmt.Sprintf("TestFileHandler_BeginWrite-%d", time.Now().UnixNano())
	id, err := fileHandler.BeginWrite(context.Background(), fileName)
	assert.NoError(t, err, "BeginWrite")

	// Check that the session and the file have been created.
	session, ok := fileHandler.writeSessions[id]
	require.True(t, ok, "fileHandler.writeSessions[id]")
	require.NotNil(t, session.file, "session.file")

	assert.Equal(t, "/tmp/"+fileName, session.file.Name())

	err = os.Remove("/tmp/" + fileName)
	assert.NoError(t, err, "os.Remove")
}

func TestFileHandler_BeginWrite_Fail(t *testing.T) {

	fileHandler := &localFileHandler{
		storagePath:   "/tmp",
		writeSessions: make(map[uuid.UUID]*session),
	}

	_, err := fileHandler.BeginWrite(context.Background(), "")
	assert.EqualError(t, err, ErrFileNameMissing.Error(), "BeginWrite")
}

func TestFileHandler_WriteChunk(t *testing.T) {

	filePath := fmt.Sprintf("/tmp/TestFileHandler_WriteChunk-%d", time.Now().UnixNano())

	file, err := os.Create(filePath)
	require.NoError(t, err, "os.Create")

	_, err = file.Write([]byte("I love to write"))
	assert.NoError(t, err, "os.Write")

	sess := newSession(file)

	fileHandler := &localFileHandler{
		storagePath:   "/tmp",
		writeSessions: map[uuid.UUID]*session{sess.id: sess},
	}

	chunk := []byte(" some data")
	err = fileHandler.WriteChunk(context.Background(), sess.id, chunk)
	assert.NoError(t, err, "WriteChunk")

	// Check that session is still open and that chunk was appended to data .
	session, ok := fileHandler.writeSessions[sess.id]
	require.True(t, ok, "fileHandler.writeSessions[id]")
	require.NotNil(t, session.file, "session.file")

	data := make([]byte, 100)
	session.file.Seek(0, 0)
	n, err := session.file.Read(data)
	assert.NoError(t, err, "session.file.Read()")
	assert.Equal(t, []byte("I love to write some data"), data[:n], "session.file.Read()")

	err = os.Remove(filePath)
	assert.NoError(t, err, "os.Remove")
}

func TestFileHandler_WriteChunk_Fail(t *testing.T) {

	filePath := fmt.Sprintf("/tmp/TestFileHandler_WriteChunk_Fail-%d", time.Now().UnixNano())

	file, err := os.Create(filePath)
	require.NoError(t, err, "os.Create")

	sess := newSession(file)

	fileHandler := &localFileHandler{
		storagePath:   "/tmp",
		writeSessions: map[uuid.UUID]*session{sess.id: sess},
	}

	t.Run("no-session", func(t *testing.T) {
		err = fileHandler.WriteChunk(context.Background(), uuid.NewV4(), []byte(" some data"))
		assert.EqualError(t, err, ErrNoSession.Error(), "WriteChunk")
	})

	t.Run("fail-write-and-delete", func(t *testing.T) {
		err := sess.file.Close()
		require.NoError(t, err, "session.File.Close()")

		err = fileHandler.WriteChunk(context.Background(), sess.id, []byte("yo"))
		assert.Error(t, err, "WriteChunk")

		// Check that session has been deleted.
		_, ok := fileHandler.writeSessions[sess.id]
		require.False(t, ok, "fileHandler.writeSessions[id]")

		// Check that teh file has been deleted.
		_, err = os.Stat(sess.file.Name())
		assert.True(t, os.IsNotExist(err), "File should be deleted")
	})
}

func TestFileHandler_EndWrite(t *testing.T) {
	filePath := fmt.Sprintf("/tmp/TestFileHandler_EndWrite-%d", time.Now().UnixNano())

	file, err := os.Create(filePath)
	require.NoError(t, err, "os.Create")

	_, err = file.Write([]byte("I love to write"))
	assert.NoError(t, err, "os.Write")

	sess := newSession(file)

	db, err := db.NewMemDB(nil)
	require.NoError(t, err, "NewMemDB()")

	fileHandler := &localFileHandler{
		db:            db,
		storagePath:   "/tmp",
		writeSessions: map[uuid.UUID]*session{sess.id: sess},
	}

	hash, err := fileHandler.EndWrite(context.Background(), sess.id)
	require.NoError(t, err, "EndWrite()")

	// Check that hash is correct.
	f, err := os.Open(filePath)
	assert.NoError(t, err)

	h := sha256.New()
	_, err = io.Copy(h, f)
	assert.NoError(t, err, "io.Copy")

	expected, err := mh.Encode(h.Sum(nil), mh.SHA2_256)
	assert.NoError(t, err, "mh.Encode")
	assert.Equal(t, expected, hash, "file hash incorrect")

	// Check that file hash and path are in db.
	p, err := db.Get(append(prefixFilesHashes, hash...))
	assert.NoError(t, err, "db.Get()")
	assert.Equal(t, filePath, string(p), "incorrect path saved")

	err = os.Remove(filePath)
	assert.NoError(t, err, "os.Remove")
}

func TestFileHandler_EndWrite_Fail(t *testing.T) {
	filePath := fmt.Sprintf("/tmp/TestFileHandler_EndWrite_Fail-%d", time.Now().UnixNano())

	file, err := os.Create(filePath)
	require.NoError(t, err, "os.Create")

	_, err = file.Write([]byte("I love to write"))
	assert.NoError(t, err, "os.Write")

	sess := newSession(file)

	db, err := db.NewMemDB(nil)
	require.NoError(t, err, "NewMemDB()")

	fileHandler := &localFileHandler{
		db:            db,
		storagePath:   "/tmp",
		writeSessions: map[uuid.UUID]*session{sess.id: sess},
	}

	t.Run("no-session", func(t *testing.T) {
		_, err = fileHandler.EndWrite(context.Background(), uuid.NewV4())
		assert.EqualError(t, err, ErrNoSession.Error(), "WriteChunk")
	})

	t.Run("fail-and-delete-file", func(t *testing.T) {
		err := db.Close()
		assert.NoError(t, err, "db.Close()")

		_, err = fileHandler.EndWrite(context.Background(), sess.id)
		assert.Error(t, err, "WriteChunk")

		// Check that session has been deleted.
		_, ok := fileHandler.writeSessions[sess.id]
		require.False(t, ok, "fileHandler.writeSessions[id]")

		// Check that teh file has been deleted.
		_, err = os.Stat(sess.file.Name())
		assert.True(t, os.IsNotExist(err), "File should be deleted")

	})
}

// ============================================================================
// ==== 															Read																 ====
// ============================================================================

func TestFileHandler_BeginRead(t *testing.T) {
	filePath := fmt.Sprintf("/tmp/TestFileHandler_BeginRead-%d", time.Now().UnixNano())
	fileHash := []byte("file hash")

	_, err := os.Create(filePath)
	require.NoError(t, err, "os.Create")
	defer os.Remove(filePath)

	db, err := db.NewMemDB(nil)
	require.NoError(t, err)

	err = db.Put(append(prefixFilesHashes, fileHash...), []byte(filePath))
	require.NoError(t, err)

	fileHandler := &localFileHandler{
		db:           db,
		readSessions: make(map[uuid.UUID]*session),
		storagePath:  "/tmp",
	}

	id, path, err := fileHandler.BeginRead(context.Background(), fileHash)
	assert.NoError(t, err, "BeginRead")
	assert.Equal(t, filePath, path, "BeginRead")

	// Check that the session has teh right file.
	session, ok := fileHandler.readSessions[id]
	require.True(t, ok, "fileHandler.readSessions[id]")
	require.NotNil(t, session.file, "session.file")
	assert.Equal(t, filePath, session.file.Name(), "session.file.Name()")
}

func TestFileHandler_ReadChunk(t *testing.T) {

	filePath := fmt.Sprintf("/tmp/TestFileHandler_ReadChunk-%d", time.Now().UnixNano())

	file, err := os.Create(filePath)
	require.NoError(t, err, "os.Create")
	defer os.Remove(filePath)

	data := []byte("The file has some data")
	_, err = file.Write(data)
	assert.NoError(t, err, "os.Write")

	_, err = file.Seek(0, 0)
	assert.NoError(t, err)

	sess := newSession(file)
	chunkSize := 12

	fileHandler := &localFileHandler{
		readSessions: map[uuid.UUID]*session{sess.id: sess},
		storagePath:  "/tmp",
	}

	// Read first chunk.
	chunk, err := fileHandler.ReadChunk(context.Background(), sess.id, chunkSize)
	assert.NoError(t, err, "ReadChunk")
	assert.Equal(t, data[:12], chunk, "first chunk")

	// Check that second chunk is read second.
	chunk, err = fileHandler.ReadChunk(context.Background(), sess.id, chunkSize)
	assert.NoError(t, err, "ReadChunk")
	assert.Equal(t, data[chunkSize:], chunk, "second chunk")

	// Check that we have EOF as there should be no more data to read.
	_, err = fileHandler.ReadChunk(context.Background(), sess.id, chunkSize)
	assert.EqualError(t, err, io.EOF.Error(), "ReadChunk")
}

func TestFileHandler_ReadChunk_Fail(t *testing.T) {
	filePath := fmt.Sprintf("/tmp/TestFileHandler_ReadChunk-%d", time.Now().UnixNano())

	file, err := os.Create(filePath)
	require.NoError(t, err, "os.Create")
	defer os.Remove(filePath)

	sess := newSession(file)
	chunkSize := 12

	fileHandler := &localFileHandler{
		readSessions: map[uuid.UUID]*session{sess.id: sess},
		storagePath:  "/tmp",
	}

	t.Run("no-session", func(t *testing.T) {
		_, err := fileHandler.ReadChunk(context.Background(), uuid.NewV4(), chunkSize)
		assert.EqualError(t, err, ErrNoSession.Error(), "ReadChunk")
	})

	t.Run("fail-and-delete-session", func(t *testing.T) {
		err = file.Close()
		assert.NoError(t, err, "file.Close()")

		_, err := fileHandler.ReadChunk(context.Background(), sess.id, chunkSize)
		assert.Error(t, err, "ReadChunk")

		_, ok := fileHandler.readSessions[sess.id]
		assert.False(t, ok, "readSessions[id]")
	})
}

func TestFileHandler_EndRead(t *testing.T) {
	fileName := fmt.Sprintf("TestFileHandler_EndRead-%d", time.Now().UnixNano())

	file, err := os.Create("/tmp/" + fileName)
	require.NoError(t, err, "os.Create")

	sess := newSession(file)

	fileHandler := &localFileHandler{
		readSessions: map[uuid.UUID]*session{sess.id: sess},
		storagePath:  "/tmp",
	}

	err = fileHandler.EndRead(context.Background(), sess.id)
	require.NoError(t, err, "EndRead()")

	// Check that session was deleted and file closed.
	_, ok := fileHandler.readSessions[sess.id]
	assert.False(t, ok, "readSessions[id]")
}

func TestFileHandler_EndRead_Fail(t *testing.T) {

	fileHandler := &localFileHandler{
		readSessions: map[uuid.UUID]*session{},
		storagePath:  "/tmp",
	}

	err := fileHandler.EndRead(context.Background(), uuid.NewV4())
	assert.EqualError(t, err, ErrNoSession.Error(), "ReadChunk")
}
