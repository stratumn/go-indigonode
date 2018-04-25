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
	"io"
	"time"

	"github.com/pkg/errors"
	"github.com/satori/go.uuid"

	grpcpb "github.com/stratumn/alice/grpc/storage"
	pb "github.com/stratumn/alice/pb/storage"

	logging "gx/ipfs/QmSpJByNKFX1sCsHBEp3R73FL4NF6FnQTEGyNAXHm2GS52/go-log"
)

var (
	// ErrFileNameMissing is returned when no file name was given.
	ErrFileNameMissing = errors.New("the first chunk should have the filename")

	// ErrUploadSessionNotFound is returned when no upload session could be found.
	ErrUploadSessionNotFound = errors.New("the given session was not found")

	// ErrInvalidUploadSession is returned when the session id could not be parsed.
	ErrInvalidUploadSession = errors.New("the given session could not be parsed")
	// UploadTimeout is the timeout duration for the upload.
	// TODO: a timeout per request ?
	UploadTimeout = time.Minute * 20
)

// grpcServer is a gRPC server for the storage service.
type grpcServer struct {
	saveFile  func(context.Context, <-chan *pb.FileChunk) ([]byte, error)
	authorize func(ctx context.Context, peerIds [][]byte, fileHash []byte) error
	download  func(ctx context.Context, fileHash []byte, peerId []byte) error

	storagePath   string
	sessions      map[uuid.UUID]*session
	uploadTimeout time.Duration
}

func newGrpcServer(
	saveFile func(context.Context, <-chan *pb.FileChunk) ([]byte, error),
	authorize func(ctx context.Context, peerIds [][]byte, fileHash []byte) error,
	download func(ctx context.Context, fileHash []byte, peerId []byte) error,
	storagePath string,
	uploadTimeout time.Duration,
) *grpcServer {

	return &grpcServer{
		saveFile:      saveFile,
		authorize:     authorize,
		download:      download,
		storagePath:   storagePath,
		sessions:      make(map[uuid.UUID]*session),
		uploadTimeout: uploadTimeout,
	}
}

// Upload saves a file on the alice node.
// The first message must contain the file name.
func (s grpcServer) Upload(stream grpcpb.Storage_UploadServer) (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), UploadTimeout)
	event := log.EventBegin(ctx, "UploadFile")

	defer func() {
		if err != nil {
			event.SetError(err)
		}
		event.Done()
		cancel()
	}()

	// Save the file
	chunkCh := make(chan *pb.FileChunk)
	errCh := make(chan error)
	resCh := make(chan []byte)
	go func() {
		hash, err := s.saveFile(ctx, chunkCh)
		if err != nil {
			errCh <- err
		}
		resCh <- hash
	}()

	for {
		var chunk *pb.FileChunk
		chunk, err = stream.Recv()

		if err == io.EOF {
			close(chunkCh)
			err = nil
			break
		}
		if err != nil {
			return
		}

		select {

		case hash := <-resCh:
			err = stream.SendAndClose(&grpcpb.UploadAck{
				FileHash: hash,
			})
			return

		case err = <-errCh:
			return

		case chunkCh <- chunk:
			continue
		}

	}

	select {
	case hash := <-resCh:
		err = stream.SendAndClose(&grpcpb.UploadAck{
			FileHash: hash,
		})
		return

	case err = <-errCh:
		return
	}
}

// AuthorizePeers gives access for a list of peers to a resource.
func (s grpcServer) AuthorizePeers(ctx context.Context, req *grpcpb.AuthRequest) (*grpcpb.Ack, error) {

	if err := s.authorize(ctx, req.PeerIds, req.FileHash); err != nil {
		return nil, err
	}

	return &grpcpb.Ack{}, nil
}

func (s grpcServer) Download(ctx context.Context, req *grpcpb.DownloadRequest) (*grpcpb.Ack, error) {

	if err := s.download(ctx, req.FileHash, req.PeerId); err != nil {
		return nil, err
	}

	return &grpcpb.Ack{}, nil
}

// ####################################################################################################################
// #####																		 Sequential upload protocol																						#####
// ####################################################################################################################

type session struct {
	id       uuid.UUID
	fileName string
	errCh    chan error
	chunkCh  chan *pb.FileChunk
	// Returns the file hash when we are done writing the file.
	resCh chan ([]byte)
}

func newSession(fileName string) *session {
	id := uuid.NewV4()
	return &session{
		id:       id,
		fileName: fileName,
		chunkCh:  make(chan (*pb.FileChunk)),
		// doneCh:   make(chan struct{}),
		errCh: make(chan error),
		resCh: make(chan ([]byte)),
	}
}

func (s *grpcServer) StartUpload(ctx context.Context, req *grpcpb.UploadReq) (*grpcpb.UploadSession, error) {
	event := log.EventBegin(ctx, "StartUpload")
	defer event.Done()

	if req.FileName == "" {
		event.SetError(ErrFileNameMissing)
		return nil, ErrFileNameMissing
	}

	session := newSession(req.FileName)

	event.Append(logging.Metadata{"sessionID": session.id})

	go func() {
		writeCtx, cancel := context.WithTimeout(context.Background(), s.uploadTimeout)
		defer cancel()
		hash, err := s.saveFile(writeCtx, session.chunkCh)

		if err != nil {
			log.Errorf(err.Error())
			session.errCh <- err
		} else {
			session.resCh <- hash
		}
	}()

	s.sessions[session.id] = session

	return &grpcpb.UploadSession{
		Id: session.id.Bytes(),
	}, nil
}

func (s *grpcServer) UploadChunk(ctx context.Context, req *grpcpb.SessionFileChunk) (*grpcpb.Ack, error) {
	event := log.EventBegin(ctx, "UploadChunk")
	defer event.Done()

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

	session.chunkCh <- &pb.FileChunk{
		Data:     req.Data,
		FileName: session.fileName,
	}

	return &grpcpb.Ack{}, nil
}

func (s *grpcServer) EndUpload(ctx context.Context, req *grpcpb.UploadSession) (*grpcpb.UploadAck, error) {
	event := log.EventBegin(ctx, "EndUpload")
	defer event.Done()

	u, err := uuid.FromBytes(req.Id)
	if err != nil {
		return nil, ErrInvalidUploadSession
	}
	event.Append(logging.Metadata{"sessionID": u})

	session, ok := s.sessions[u]
	if !ok {
		return nil, ErrUploadSessionNotFound
	}

	close(session.chunkCh)

	select {
	case err := <-session.errCh:
		return nil, err
	case fileHash := <-session.resCh:
		delete(s.sessions, u)

		return &grpcpb.UploadAck{
			FileHash: fileHash,
		}, nil

	}
}
