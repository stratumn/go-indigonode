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
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/pkg/errors"
	uuid "github.com/satori/go.uuid"
	grpcpb "github.com/stratumn/alice/grpc/storage"
	"github.com/stratumn/alice/grpc/storage/mockstorage"
	pb "github.com/stratumn/alice/pb/storage"
	"github.com/stratumn/alice/test"
	"github.com/stretchr/testify/assert"
)

const storagePath = "/tmp"

func TestGRPCServer_UploadFile(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	stream := mockstorage.NewMockStorage_UploadServer(ctrl)
	saveFnCalled := false

	server := newGrpcServer(
		// saveFile.
		func(ctx context.Context, ch <-chan *pb.FileChunk) ([]byte, error) {
			saveFnCalled = true
			content := []byte{}

			for {
				chunk, ok := <-ch
				if !ok {
					break
				}
				content = append(content, chunk.Data...)
			}

			assert.Equal(t, []byte("coucou, tu veux voir mon fichier ?"), content, "saveFile")
			return []byte("123"), nil
		},
		nil,
		nil,
		storagePath,
		1*time.Second)

	filename := "grpc_test"

	chunk1 := &pb.FileChunk{
		FileName: filename,
		Data:     []byte("coucou, "),
	}

	chunk2 := &pb.FileChunk{
		Data: []byte("tu veux voir "),
	}

	chunk3 := &pb.FileChunk{
		Data: []byte("mon fichier ?"),
	}

	gomock.InOrder(
		stream.EXPECT().Recv().Return(chunk1, nil),
		stream.EXPECT().Recv().Return(chunk2, nil),
		stream.EXPECT().Recv().Return(chunk3, nil),
		stream.EXPECT().Recv().Return(nil, io.EOF),
		stream.EXPECT().SendAndClose(&grpcpb.UploadAck{FileHash: []byte("123")}).Return(nil),
	)

	err := server.Upload(stream)
	assert.NoError(t, err, "server.Upload()")

	assert.True(t, saveFnCalled, "Index function was not called")
}

func TestGRPCServer_NewGrpcServer(t *testing.T) {
	file, err := os.Create(filepath.Join(storagePath, tmpStorageSubdir, "abc"))
	assert.NoError(t, err, "os.Create()")

	_, err = os.Stat(file.Name())
	assert.NoError(t, err, "Temporary file could not be found.")

	newGrpcServer(
		nil,
		nil,
		nil,
		storagePath,
		10*time.Millisecond)

	_, err = os.Stat(file.Name())
	assert.True(t, os.IsNotExist(err))
}

func TestGRPCServer_StartUpload(t *testing.T) {
	fileName := "grpc_test"

	saveFnCalled := false
	server := newGrpcServer(
		func(ctx context.Context, ch <-chan *pb.FileChunk) ([]byte, error) {
			saveFnCalled = true
			return []byte("zoulou"), nil
		},
		nil,
		nil,
		storagePath,
		10*time.Millisecond)

	uploadReq := &grpcpb.UploadReq{
		FileName: "grpc_test",
	}

	sess, err := server.StartUpload(context.Background(), uploadReq)
	assert.NoError(t, err, "server.StartUpload()")

	u, err := uuid.FromBytes(sess.Id)
	assert.NoError(t, err, "server.StartUpload()")

	session, ok := server.sessions[u]
	assert.True(t, ok, "server.serssions(uuid)")
	assert.Equal(t, fileName, session.fileName, "session.fielName")

	test.WaitUntil(t, 2*server.uploadTimeout, time.Millisecond*1, func() error {
		hash := <-session.resCh
		assert.Equal(t, []byte("zoulou"), hash, "<-session.resCh")
		return nil
	}, "incorrect file hash.")

	assert.True(t, saveFnCalled, "Save function was not called")
}

func TestGRPCServer_StartUpload_FileNameMissing(t *testing.T) {
	server := newGrpcServer(
		nil,
		nil,
		nil,
		storagePath,
		10*time.Millisecond)

	uploadReq := &grpcpb.UploadReq{
		FileName: "",
	}

	_, err := server.StartUpload(context.Background(), uploadReq)
	assert.EqualError(t, err, ErrFileNameMissing.Error())
}

func TestGRPCServer_UploadChunk(t *testing.T) {
	filename := fmt.Sprintf("grpc_test_%d", time.Now().UnixNano())
	u := uuid.NewV4()
	s := newSession(filename)

	server := newGrpcServer(nil, nil, nil, storagePath, 1*time.Second)
	server.sessions[u] = s

	t.Run("With a valid session ID", func(t *testing.T) {
		chunk := &grpcpb.SessionFileChunk{
			Id:   u.Bytes(),
			Data: []byte("coucou, "),
		}

		go func() {
			_, err := server.UploadChunk(context.Background(), chunk)
			assert.NoError(t, err, "server.UploadChunk()")
		}()

		receivedChunk := <-s.chunkCh
		assert.Equal(t, filename, receivedChunk.FileName, "chunk.FileName")
		assert.Equal(t, chunk.Data, receivedChunk.Data, "chunk.FileName")

	})

	t.Run("With an invalid session ID", func(t *testing.T) {
		chunk := &grpcpb.SessionFileChunk{
			Id:   []byte("123"),
			Data: []byte("coucou, "),
		}

		_, err := server.UploadChunk(context.Background(), chunk)
		assert.EqualError(t, err, ErrInvalidUploadSession.Error())
	})

	t.Run("With a new session ID", func(t *testing.T) {
		uNew := uuid.NewV4()
		chunk := &grpcpb.SessionFileChunk{
			Id:   uNew.Bytes(),
			Data: []byte("coucou, "),
		}

		_, err := server.UploadChunk(context.Background(), chunk)
		assert.EqualError(t, err, ErrUploadSessionNotFound.Error())
	})
}

func TestGRPCServer_EndUpload(t *testing.T) {
	filename := fmt.Sprintf("grpc_test_%d", time.Now().UnixNano())
	u := uuid.NewV4()

	server := newGrpcServer(
		nil,
		nil,
		nil,
		storagePath,
		1*time.Second)

	session := newSession(filename)
	server.sessions[u] = session

	t.Run("With a valid session ID", func(t *testing.T) {
		req := &grpcpb.UploadSession{
			Id: u.Bytes(),
		}

		go func() {
			uploadAck, err := server.EndUpload(context.Background(), req)
			assert.NoError(t, err, "server.EndUpload()")
			assert.Equal(t, uploadAck.FileHash, []byte("123"))

			// Check that the session has been deleted.
			_, ok := server.sessions[u]
			assert.False(t, ok, "server.sessions[u]")
		}()

		<-session.doneCh
		session.resCh <- []byte("123")

	})

	t.Run("With a write error", func(t *testing.T) {
		req := &grpcpb.UploadSession{
			Id: u.Bytes(),
		}

		go func() {
			_, err := server.EndUpload(context.Background(), req)
			assert.Error(t, err, "server.EndUpload()")
		}()

		<-session.doneCh
		session.errCh <- errors.New("https://goo.gl/WhZSSE")

	})

	t.Run("With an invalid session ID", func(t *testing.T) {
		req := &grpcpb.UploadSession{
			Id: []byte("123"),
		}

		_, err := server.EndUpload(context.Background(), req)
		assert.EqualError(t, err, ErrInvalidUploadSession.Error())
	})

	t.Run("With a new session ID", func(t *testing.T) {
		uNew := uuid.NewV4()
		req := &grpcpb.UploadSession{
			Id: uNew.Bytes(),
		}

		_, err := server.EndUpload(context.Background(), req)
		assert.EqualError(t, err, ErrUploadSessionNotFound.Error())
	})
}

func TestGRPCServer_AuthorizePeers(t *testing.T) {
	var fh []byte
	var pids [][]byte
	server := newGrpcServer(
		nil,
		func(ctx context.Context, peerIds [][]byte, fileHash []byte) error {
			fh = fileHash
			pids = peerIds
			return nil
		},
		nil,
		storagePath,
		1*time.Second)

	fileHash := []byte("file hash")
	peerIds := [][]byte{
		[]byte("QmVhJVRSYHNSHgR9dJNbDxu6G7GPPqJAeiJoVRvcexGNf1"),
		[]byte("QmVhJVRSYHNSHgR9dJNbDxu6G7GPPqJAeiJoVRvcexGNf2"),
		[]byte("QmVhJVRSYHNSHgR9dJNbDxu6G7GPPqJAeiJoVRvcexGNf3"),
	}

	server.AuthorizePeers(context.Background(), &grpcpb.AuthRequest{
		FileHash: fileHash,
		PeerIds:  peerIds,
	})

	assert.Equal(t, fileHash, fh, "FileHash")
	assert.Equal(t, peerIds, pids, "PeerIds")
}

func TestGRPCServer_Download(t *testing.T) {
	var fh []byte
	var pid []byte

	server := newGrpcServer(
		nil,
		nil,
		func(ctx context.Context, fileHash []byte, peerId []byte) error {
			fh = fileHash
			pid = peerId
			return nil
		},
		storagePath,
		1*time.Second)

	fileHash := []byte("file hash")
	peerID := []byte("QmVhJVRSYHNSHgR9dJNbDxu6G7GPPqJAeiJoVRvcexGNf1")

	server.Download(context.Background(), &grpcpb.DownloadRequest{
		FileHash: fileHash,
		PeerId:   peerID,
	})

	assert.Equal(t, fileHash, fh, "FileHash")
	assert.Equal(t, peerID, pid, "PeerId")
}
