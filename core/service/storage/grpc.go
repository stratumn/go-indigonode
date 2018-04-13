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
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/pkg/errors"
	pb "github.com/stratumn/alice/grpc/storage"

	logging "gx/ipfs/QmSpJByNKFX1sCsHBEp3R73FL4NF6FnQTEGyNAXHm2GS52/go-log"

	"github.com/satori/go.uuid"
)

const tmpStorageSubdir = "tmp"

var (
	// ErrFileNameMissing is returned when no file name was given.
	ErrFileNameMissing = errors.New("the first chunk should have the filename")

	// ErrUploadSessionNotFound is returned when no upload session could be found.
	ErrUploadSessionNotFound = errors.New("the given session was not found")

	// ErrInvalidUploadSession is returned when the session id could not be parsed.
	ErrInvalidUploadSession = errors.New("the given session could not be parsed")
)

type session struct {
	file            *os.File
	fileName        string
	storageFileName string
	id              uuid.UUID
	done            chan (struct{})
}

func newSession(fileName string) *session {
	id := uuid.NewV4()
	return &session{
		fileName:        fileName,
		storageFileName: id.String() + fileName,
		done:            make(chan (struct{})),
		id:              id,
	}
}

// grpcServer is a gRPC server for the storage service.
type grpcServer struct {
	indexFile      func(context.Context, *os.File, string) (fileHash []byte, err error)
	authorize      func(ctx context.Context, peerIds [][]byte, fileHash []byte) error
	storagePath    string
	tmpStoragePath string
	sessionsMu     sync.RWMutex
	sessions       map[uuid.UUID]*session
	uploadTimeout  time.Duration
}

func newGrpcServer(indexFile func(context.Context, *os.File, string) (fileHash []byte, err error),
	authorize func(ctx context.Context, peerIds [][]byte, fileHash []byte) error,
	storagePath string,
	uploadTimeout time.Duration) *grpcServer {
	tmpStoragePath := filepath.Join(storagePath, tmpStorageSubdir)

	err := os.RemoveAll(tmpStoragePath)
	if err != nil {
		log.Error(err)
	}
	err = os.MkdirAll(tmpStoragePath, 0777)
	if err != nil {
		log.Error(err)
	}

	return &grpcServer{
		indexFile:      indexFile,
		authorize:      authorize,
		storagePath:    storagePath,
		tmpStoragePath: tmpStoragePath,
		sessions:       make(map[uuid.UUID]*session),
		uploadTimeout:  uploadTimeout,
	}
}

// // SendFile sends a file to the specified peer.
// func (s grpcServer) SendFile(ctx context.Context, req *pb.File) (response *pb.Ack, err error) {
// 	response = &pb.Ack{}
// 	pid, err := peer.IDFromBytes(req.PeerId)
// 	if err != nil {
// 		err = errors.WithStack(err)
// 		return
// 	}

// 	pi := pstore.PeerInfo{ID: pid}

// 	// Make sure there is a connection to the peer.
// 	if err = s.Connect(ctx, pi); err != nil {
// 		return
// 	}

// 	if err = s.Send(ctx, pid, req.Path); err != nil {
// 		return
// 	}

// 	return
// }

// Upload saves a file on the alice node.
func (s *grpcServer) Upload(stream pb.Storage_UploadServer) (err error) {
	ctx := context.Background()
	event := log.EventBegin(ctx, "UploadFile")
	var file *os.File
	var fileName string

	defer func() {
		if err != nil {
			// Delete the partially written file.
			if file != nil {
				if err2 := os.Remove(file.Name()); err != nil {
					err = fmt.Errorf("error uploading file (%v); error deleting partially uploaded file (%v)", err, err2)
				}
			}
			event.SetError(err)
		}
		event.Done()
	}()

	// while there are messages coming process them.
	for {
		var chunk *pb.StreamFileChunk
		chunk, err = stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return
		}

		if file == nil && chunk.FileName == "" {
			event.SetError(ErrFileNameMissing)
			return ErrFileNameMissing
		}

		if file == nil {
			fileName = chunk.FileName
			file, err = os.Create(filepath.Join(s.storagePath, uuid.NewV4().String()+fileName))
			if err != nil {
				return
			}
		}

		event.Append(&logging.Metadata{"filename": file.Name()})

		if _, err = file.Write(chunk.Data); err != nil {
			return
		}
	}

	// TODO: Do we want to index the files with their hash?
	// This would be needed at some point to check file integrity
	// and to be able to handle multiple providers for a file.
	var fileHash []byte
	if fileHash, err = s.indexFile(ctx, file, fileName); err != nil {
		return
	}

	err = stream.SendAndClose(&pb.UploadAck{FileHash: fileHash})
	return
}

func (s *grpcServer) StartUpload(ctx context.Context, req *pb.UploadReq) (*pb.UploadSession, error) {
	event := log.EventBegin(ctx, "StartUpload")
	defer event.Done()

	s.sessionsMu.Lock()
	defer s.sessionsMu.Unlock()

	if req.FileName == "" {
		event.SetError(ErrFileNameMissing)
		return nil, ErrFileNameMissing
	}

	session := newSession(req.FileName)

	event.Append(logging.Metadata{"sessionID": session.id})

	file, err := os.Create(filepath.Join(s.tmpStoragePath, session.storageFileName))
	if err != nil {
		event.SetError(err)
		return nil, err
	}

	session.file = file

	go func() {
		select {
		case <-session.done:
			return
		case <-time.After(s.uploadTimeout):
			// Delete the partially written file.
			err := os.Remove(file.Name())
			if err != nil {
				log.Errorf("Temporary file could not be removed: %s", err)
			}
		}
	}()

	s.sessions[session.id] = session

	return &pb.UploadSession{
		Id: session.id.Bytes(),
	}, nil
}

func (s *grpcServer) UploadChunk(ctx context.Context, req *pb.FileChunk) (*pb.Ack, error) {
	event := log.EventBegin(ctx, "UploadChunk")
	defer event.Done()

	s.sessionsMu.RLock()
	defer s.sessionsMu.RUnlock()

	// TODO: handling of out of order chunks
	u, err := uuid.FromBytes(req.Id)
	if err != nil {
		event.SetError(ErrInvalidUploadSession)
		return nil, ErrInvalidUploadSession
	}
	event.Append(logging.Metadata{"sessionID": u})
	session, ok := s.sessions[u]
	if !ok {
		event.SetError(ErrUploadSessionNotFound)
		return nil, ErrUploadSessionNotFound
	}

	if _, err = session.file.Write(req.Data); err != nil {
		event.SetError(err)
		return nil, err
	}
	return &pb.Ack{}, nil
}

func (s *grpcServer) EndUpload(ctx context.Context, req *pb.UploadSession) (*pb.UploadAck, error) {
	event := log.EventBegin(ctx, "EndUpload")
	defer event.Done()

	s.sessionsMu.RLock()
	defer s.sessionsMu.RUnlock()

	u, err := uuid.FromBytes(req.Id)
	if err != nil {
		return nil, ErrInvalidUploadSession
	}
	event.Append(logging.Metadata{"sessionID": u})

	session, ok := s.sessions[u]
	if !ok {
		return nil, ErrUploadSessionNotFound
	}
	go func() { session.done <- struct{}{} }()

	finalPath := filepath.Join(s.storagePath, session.storageFileName)
	err = os.Rename(session.file.Name(), finalPath)
	if err != nil {
		return nil, err
	}
	file, err := os.Open(finalPath)
	if err != nil {
		return nil, err
	}

	fileHash, err := s.indexFile(ctx, file, session.fileName)
	event.Append(logging.Metadata{"fileHash": fileHash})

	return &pb.UploadAck{
		FileHash: fileHash,
	}, err
}

// AuthorizePeers gives access for a list of peers to a resource.
func (s *grpcServer) AuthorizePeers(ctx context.Context, req *pb.AuthRequest) (*pb.Ack, error) {
	if err := s.authorize(ctx, req.PeerIds, req.FileHash); err != nil {
		return nil, err
	}

	return &pb.Ack{}, nil
}
