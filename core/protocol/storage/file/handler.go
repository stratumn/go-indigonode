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

//go:generate mockgen -package mockhandler -destination mockhandler/mockhandler.go github.com/stratumn/alice/core/protocol/storage/file Handler

package file

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"

	"github.com/pkg/errors"
	"github.com/satori/go.uuid"
	"github.com/stratumn/alice/core/db"
	pb "github.com/stratumn/alice/pb/storage"

	logging "gx/ipfs/QmSpJByNKFX1sCsHBEp3R73FL4NF6FnQTEGyNAXHm2GS52/go-log"
	mh "gx/ipfs/QmZyZDi491cCNTLfAhwcaDii2Kg4pwKRkhqQzURGDvY6ua/go-multihash"
)

var log = logging.Logger("storage.file_handler")

var (
	// ErrFileNameMissing is returned when no file name was given.
	ErrFileNameMissing = errors.New("the first chunk should have the filename")

	// ErrNoSession is returned when no session id found for a given id.
	ErrNoSession = errors.New("no open file session with this id")

	// ErrUnauthorized is returned when a peer tries to access a file he
	// is not allowed to get
	ErrUnauthorized = errors.New("peer not authorized for requested file")
)

var (
	prefixFilesHashes = []byte("fh") // prefixFilesHashes + filehash -> filepath
)

// Handler contains the methods to handle a file on the alice node.
type Handler interface {
	WriteFile(context.Context, <-chan *pb.FileChunk) (fileName string, err error)

	// BeginWrite creates an empty file.
	BeginWrite(ctx context.Context, fileName string) (uuid.UUID, error)

	// WriteChunk writes a chunk of data to a file identified by its ID.
	WriteChunk(ctx context.Context, sessionID uuid.UUID, chunk []byte) (err error)

	// EndWrite is called to finalize the file writing.
	EndWrite(ctx context.Context, sessionID uuid.UUID) (fileHash []byte, err error)
}

// session represents one file write.
type session struct {
	id   uuid.UUID
	file *os.File
}

func newSession(file *os.File) *session {
	id := uuid.NewV4()
	return &session{
		id:   id,
		file: file,
	}
}

type localFileHandler struct {
	db              db.DB
	writeSessionsMu sync.RWMutex
	writeSessions   map[uuid.UUID]*session
	readSessionsMu  sync.RWMutex
	readSessions    map[uuid.UUID]*session
	storagePath     string
}

// NewLocalFileHandler create a new file Handler.
func NewLocalFileHandler(path string, db db.DB) Handler {
	return &localFileHandler{
		db:            db,
		storagePath:   path,
		writeSessions: make(map[uuid.UUID]*session),
		readSessions:  make(map[uuid.UUID]*session),
	}
}

// SaveFile saves a file locally.
func (h *localFileHandler) WriteFile(ctx context.Context, chunkCh <-chan *pb.FileChunk) (fileName string, err error) {
	event := log.EventBegin(ctx, "SaveFile")

	var file *os.File
	defer func() {
		cerr := file.Close()
		if err == nil {
			err = cerr
		}
	}()

	defer func() {
		if err != nil {
			// Delete the partially written file.
			if file != nil {
				if err2 := os.Remove(file.Name()); err2 != nil {
					err = errors.Wrap(err, err2.Error())
				}
			}
			event.SetError(err)
		}
		event.Done()
	}()

	for {
		select {
		case <-ctx.Done():
			err = ctx.Err()
			return

		case chunk, ok := <-chunkCh:
			if !ok {
				return
			}
			if file == nil && chunk.FileName == "" {
				err = ErrFileNameMissing
				return
			}

			if file == nil {
				file, err = os.Create(fmt.Sprintf("%s/%s", h.storagePath, chunk.FileName))
				if err != nil {
					return
				}
				fileName = file.Name()
				event.Append(&logging.Metadata{"filename": file.Name()})
			}

			if _, err = file.Write(chunk.Data); err != nil {
				return
			}
		}
	}
}

// ============================================================================
// ====															Sequential write											 ====
// ============================================================================

// BeginWrite creates an empty file and attaches it to a session.
func (h *localFileHandler) BeginWrite(ctx context.Context, fileName string) (uuid.UUID, error) {
	event := log.EventBegin(ctx, "BeginWrite", &logging.Metadata{"fileName": fileName})
	defer event.Done()

	file, err := os.Create(filepath.Join(h.storagePath, fileName))
	if err != nil {
		event.SetError(err)
		return uuid.Nil, errors.WithStack(err)
	}

	session := newSession(file)
	h.writeSessionsMu.Lock()
	h.writeSessions[session.id] = session
	h.writeSessionsMu.Unlock()
	event.Append(&logging.Metadata{"sessionID": session.id})

	return session.id, nil
}

// WriteChunk writes a chunk of data to a file identified by its session ID.
func (h *localFileHandler) WriteChunk(ctx context.Context, sessionID uuid.UUID, chunk []byte) error {
	event := log.EventBegin(ctx, "WriteChunk", &logging.Metadata{"sessionID": sessionID})
	defer event.Done()

	h.writeSessionsMu.RLock()
	session, ok := h.writeSessions[sessionID]
	h.writeSessionsMu.RUnlock()
	if !ok {
		event.SetError(ErrNoSession)
		return ErrNoSession
	}

	_, err := session.file.Write(chunk)
	if err != nil {
		err2 := h.deleteFile(ctx, sessionID)
		if err2 != nil {
			err = errors.Wrap(err, err2.Error())
		}
		event.SetError(err)
		return err
	}
	return nil
}

// EndWrite must be called at the end of the writing process.
// It indexes the file, cleans the session and returns the filehash.
func (h *localFileHandler) EndWrite(ctx context.Context, sessionID uuid.UUID) ([]byte, error) {
	event := log.EventBegin(ctx, "EndWrite", &logging.Metadata{"sessionID": sessionID})

	h.writeSessionsMu.RLock()
	session, ok := h.writeSessions[sessionID]
	h.writeSessionsMu.RUnlock()

	if !ok {
		event.SetError(ErrNoSession)
		return nil, ErrNoSession
	}

	fileHash, err := h.indexFile(ctx, session.file)
	if err != nil {
		err2 := h.deleteFile(ctx, sessionID)
		if err2 != nil {
			err = errors.Wrap(err, err2.Error())
		}
		event.SetError(err)
		return nil, err
	}

	h.writeSessionsMu.Lock()
	delete(h.writeSessions, sessionID)
	h.writeSessionsMu.Unlock()

	if err = session.file.Close(); err != nil {
		event.Append(&logging.Metadata{"close_file_error": err.Error()})
	}

	return fileHash, nil
}

// deleteFile deletes a file and its session.
// Used to clean partially written files when an error occurs.
func (h *localFileHandler) deleteFile(ctx context.Context, sessionID uuid.UUID) (err error) {
	event := log.EventBegin(ctx, "DeleteFile", &logging.Metadata{"sessionID": sessionID})
	defer func() {
		if err != nil {
			event.SetError(err)
		}
		event.Done()
	}()

	h.writeSessionsMu.RLock()
	session, ok := h.writeSessions[sessionID]
	h.writeSessionsMu.RUnlock()

	if !ok {
		err = ErrNoSession
		return
	}

	if err = session.file.Close(); err != nil {
		event.Append(&logging.Metadata{"close_file_error": err.Error()})
	}

	err = os.Remove(session.file.Name())
	if err != nil {
		return
	}

	h.writeSessionsMu.Lock()
	delete(h.writeSessions, sessionID)
	h.writeSessionsMu.Unlock()

	return
}

// ============================================================================
// ====															Sequential read											 	 ====
// ============================================================================

// BeginRead opens a file given its hash and attaches it to a session.
func (h *localFileHandler) BeginRead(ctx context.Context, fileHash []byte) (uuid.UUID, error) {
	event := log.EventBegin(ctx, "BeginRead", &logging.Metadata{"fileHash": hex.EncodeToString(fileHash)})
	defer event.Done()

	filePath, err := h.getFilePath(ctx, fileHash)
	if err != nil {
		event.SetError(err)
		return uuid.Nil, err
	}

	file, err := os.Open(filepath.Join(h.storagePath, filePath))
	if err != nil {
		event.SetError(err)
		return uuid.Nil, errors.WithStack(err)
	}

	session := newSession(file)

	h.readSessionsMu.Lock()
	h.readSessions[session.id] = session
	h.readSessionsMu.Unlock()

	event.Append(&logging.Metadata{"sessionID": session.id})

	return session.id, nil
}

// ReadChunk reads a chunk of data.
func (h *localFileHandler) ReadChunk(ctx context.Context, sessionID uuid.UUID, chunkSize int) ([]byte, error) {
	event := log.EventBegin(ctx, "ReadChunk", &logging.Metadata{"sessionID": sessionID})
	defer event.Done()

	h.readSessionsMu.RLock()
	session, ok := h.readSessions[sessionID]
	h.readSessionsMu.RUnlock()
	if !ok {
		event.SetError(ErrNoSession)
		return nil, ErrNoSession
	}

	chunk := make([]byte, chunkSize)
	n, err := session.file.Read(chunk)
	if err != nil {
		if err != io.EOF {
			event.SetError(err)
			h.readSessionsMu.Lock()
			delete(h.readSessions, sessionID)
			h.readSessionsMu.Unlock()

			if err = session.file.Close(); err != nil {
				event.Append(&logging.Metadata{"close_file_error": err.Error()})
			}
		}
		return nil, err
	}

	return chunk[:n], nil
}

// EndRead must be called at the end of the read process to delete the session.
func (h *localFileHandler) EndRead(ctx context.Context, sessionID uuid.UUID) error {
	event := log.EventBegin(ctx, "EndRead", &logging.Metadata{"sessionID": sessionID})
	defer event.Done()

	h.readSessionsMu.RLock()
	session, ok := h.readSessions[sessionID]
	h.readSessionsMu.RUnlock()

	if !ok {
		event.SetError(ErrNoSession)
		return ErrNoSession
	}

	h.readSessionsMu.Lock()
	delete(h.readSessions, sessionID)
	h.readSessionsMu.Unlock()

	if err := session.file.Close(); err != nil {
		event.Append(&logging.Metadata{"close_file_error": err.Error()})
	}
	return nil
}

// ============================================================================
// ====															indexing														 	 ====
// ============================================================================

// indexFile adds the file hash and name to the db.
func (h *localFileHandler) indexFile(ctx context.Context, file *os.File) ([]byte, error) {
	// go back to the beginning of the file.
	if _, err := file.Seek(0, 0); err != nil {
		return nil, err
	}

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return nil, err
	}

	fileHash, err := mh.Encode(hash.Sum(nil), mh.SHA2_256)
	if err != nil {
		return nil, err
	}

	if err = h.db.Put(append(prefixFilesHashes, fileHash...), []byte(file.Name())); err != nil {
		return nil, err
	}

	return fileHash, nil
}

// getFilePath returns the file path given its hash.
func (h *localFileHandler) getFilePath(ctx context.Context, fileHash []byte) (string, error) {
	p, err := h.db.Get(append(prefixFilesHashes, fileHash...))
	if err != nil {
		return "", err
	}

	return string(p), nil
}
