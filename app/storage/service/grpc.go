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

package service

import (
	"context"
	"io"
	"path/filepath"
	"time"

	"github.com/pkg/errors"
	"github.com/satori/go.uuid"

	grpcpb "github.com/stratumn/alice/app/storage/grpc"
	"github.com/stratumn/alice/app/storage/pb"

	logging "gx/ipfs/QmSpJByNKFX1sCsHBEp3R73FL4NF6FnQTEGyNAXHm2GS52/go-log"
)

const (
	chunkSize = 1024
)

var (
	// ErrFileNameMissing is returned when no file name was given.
	ErrFileNameMissing = errors.New("the first chunk should have the filename")

	// ErrUploadSessionNotFound is returned when no upload session could be found.
	ErrUploadSessionNotFound = errors.New("the given session was not found")

	// ErrInvalidUploadSession is returned when the session id could not be parsed.
	ErrInvalidUploadSession = errors.New("the given session could not be parsed")
)

type chunkReader struct {
	first bool
	ss    grpcpb.Storage_DownloadServer
}

func newChunkReader(ss grpcpb.Storage_DownloadServer) *chunkReader {
	return &chunkReader{
		first: true,
		ss:    ss,
	}
}

func (cr *chunkReader) OnChunk(data []byte, filePath string) error {
	chunk := &pb.FileChunk{Data: data}
	if cr.first {
		cr.first = false
		chunk.FileName = filepath.Base(filePath)
	}

	return cr.ss.Send(chunk)
}

// grpcServer is a gRPC server for the storage service.
type grpcServer struct {
	beginWrite func(context.Context, string) (uuid.UUID, error)
	writeChunk func(context.Context, uuid.UUID, []byte) error
	endWrite   func(context.Context, uuid.UUID) ([]byte, error)
	abortWrite func(context.Context, uuid.UUID) error

	authorize func(ctx context.Context, peerIds [][]byte, fileHash []byte) error
	download  func(ctx context.Context, fileHash []byte, peerId []byte) error

	readChunks func(ctx context.Context, fileHash []byte, chunkSize int, cr *chunkReader) error

	uploadTimeout time.Duration
}

// Upload saves a file on the alice node.
// The first message must contain the file name.
func (s *grpcServer) Upload(stream grpcpb.Storage_UploadServer) error {
	ctx, cancel := context.WithTimeout(context.Background(), s.uploadTimeout)
	defer cancel()

	event := log.EventBegin(ctx, "UploadFile")
	defer event.Done()

	// init the file write.
	var chunk *pb.FileChunk
	chunk, err := stream.Recv()
	if err != nil {
		event.SetError(err)
		return err
	}

	sessionID, err := s.beginWrite(ctx, chunk.FileName)
	if err != nil {
		event.SetError(err)
		return err
	}

	// save the chunks.
LOOP:
	for {
		select {

		case <-ctx.Done():
			err := ctx.Err()
			if err2 := s.abortWrite(ctx, sessionID); err2 != nil {
				err = errors.Wrap(err, err2.Error())
			}
			event.SetError(err)
			return err

		default:
			err := s.writeChunk(ctx, sessionID, chunk.Data)
			if err != nil {
				event.SetError(err)
				return err
			}

			chunk, err = stream.Recv()

			if err == io.EOF {
				break LOOP
			}

			if err != nil {
				if err2 := s.abortWrite(ctx, sessionID); err2 != nil {
					err = errors.Wrap(err, err2.Error())
				}
				event.SetError(err)
				return err
			}
		}
	}

	// finalize the writing.
	fileHash, err := s.endWrite(ctx, sessionID)
	if err != nil {
		event.SetError(err)
		return err
	}

	err = stream.SendAndClose(&grpcpb.UploadAck{
		FileHash: fileHash,
	})

	if err != nil {
		event.SetError(err)
		return err
	}

	return nil
}

// AuthorizePeers gives access for a list of peers to a resource.
func (s *grpcServer) AuthorizePeers(ctx context.Context, req *grpcpb.AuthRequest) (*grpcpb.Ack, error) {

	if err := s.authorize(ctx, req.PeerIds, req.FileHash); err != nil {
		return nil, err
	}

	return &grpcpb.Ack{}, nil
}

func (s *grpcServer) Download(req *grpcpb.DownloadRequest, ss grpcpb.Storage_DownloadServer) error {
	err := s.download(ss.Context(), req.FileHash, req.PeerId)
	if err != nil {
		return err
	}

	return s.readChunks(ss.Context(), req.FileHash, chunkSize, newChunkReader(ss))
}

// ##############################################################
// #####		Sequential upload protocol					#####
// ##############################################################

func (s *grpcServer) StartUpload(ctx context.Context, req *grpcpb.UploadReq) (*grpcpb.UploadSession, error) {
	event := log.EventBegin(ctx, "StartUpload")
	defer event.Done()

	sessionID, err := s.beginWrite(ctx, req.FileName)
	if err != nil {
		event.SetError(err)
		return nil, err
	}

	return &grpcpb.UploadSession{
		Id: sessionID.Bytes(),
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

	err = s.writeChunk(ctx, u, req.Data)
	if err != nil {
		event.SetError(err)
		return nil, err
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

	fileHash, err := s.endWrite(ctx, u)
	if err != nil {
		event.SetError(err)
		return nil, err
	}

	return &grpcpb.UploadAck{
		FileHash: fileHash,
	}, nil

}
