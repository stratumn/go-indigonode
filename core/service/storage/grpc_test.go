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
	"github.com/satori/go.uuid"
	pb "github.com/stratumn/alice/grpc/storage"
	"github.com/stratumn/alice/grpc/storage/mockstorage"
	"github.com/stretchr/testify/assert"
)

const storagePath = "/tmp"

func TestGRPCServer_UploadFile(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	stream := mockstorage.NewMockStorage_UploadServer(ctrl)

	var indexedFile *os.File
	server := &grpcServer{
		indexFile: func(ctx context.Context, file *os.File) ([]byte, error) {
			indexedFile = file
			return []byte("123"), nil
		},
		storagePath: storagePath,
	}
	filename := fmt.Sprintf("grpc_test_%d", time.Now().UnixNano())

	chunk1 := &pb.StreamFileChunk{
		FileName: filename,
		Data:     []byte("coucou, "),
	}

	chunk2 := &pb.StreamFileChunk{
		Data: []byte("tu veux voir "),
	}

	chunk3 := &pb.StreamFileChunk{
		Data: []byte("mon fichier ?"),
	}

	gomock.InOrder(
		stream.EXPECT().Recv().Return(chunk1, nil),
		stream.EXPECT().Recv().Return(chunk2, nil),
		stream.EXPECT().Recv().Return(chunk3, nil),
		stream.EXPECT().Recv().Return(nil, io.EOF),
		stream.EXPECT().SendAndClose(&pb.UploadAck{FileHash: []byte("123")}).Return(nil),
	)

	err := server.Upload(stream)
	assert.NoError(t, err, "server.Upload()")

	f := checkFileContent(t, filename, []byte("coucou, tu veux voir mon fichier ?"))

	// Check that the file was correctly sent to be indexed.
	expect, err := f.Stat()
	assert.NoError(t, err, "f.Stat()")
	got, err := indexedFile.Stat()
	assert.NoError(t, err, "indexedFile.Stat()")

	assert.True(t, os.SameFile(expect, got), "SameFile")

	err = os.Remove(filepath.Join(storagePath, filename))
	assert.NoError(t, err, "Remove()")
}

func TestGRPCServer_StartUpload(t *testing.T) {
	server := &grpcServer{
		storagePath: storagePath,
	}
	filename := fmt.Sprintf("grpc_test_%d", time.Now().UnixNano())

	uploadReq := &pb.UploadReq{
		FileName: filename,
	}

	session, err := server.StartUpload(context.Background(), uploadReq)
	assert.NoError(t, err, "server.StartUpload()")

	_, err = uuid.FromBytes(session.Id)
	assert.NoError(t, err, "server.StartUpload()")

	err = os.Remove(filepath.Join(storagePath, filename))
	assert.NoError(t, err, "Remove()")
}

func TestGRPCServer_UploadChunk(t *testing.T) {
	filename := fmt.Sprintf("grpc_test_%d", time.Now().UnixNano())
	u := uuid.NewV4()

	file, err := os.Create(filepath.Join(storagePath, filename))
	assert.NoError(t, err, "os.Create()")

	server := &grpcServer{
		storagePath: storagePath,
		sessions: map[uuid.UUID]*os.File{
			u: file,
		},
	}

	t.Run("With a valid session ID", func(t *testing.T) {
		chunk := &pb.FileChunk{
			Id:   u.Bytes(),
			Data: []byte("coucou, "),
		}

		_, err = server.UploadChunk(context.Background(), chunk)
		assert.NoError(t, err, "server.UploadChunk()")

		checkFileContent(t, filename, []byte("coucou, "))

		err = os.Remove(filepath.Join(storagePath, filename))
		assert.NoError(t, err, " Remove()")
	})

	t.Run("With an invalid session ID", func(t *testing.T) {
		chunk := &pb.FileChunk{
			Id:   []byte("123"),
			Data: []byte("coucou, "),
		}

		_, err = server.UploadChunk(context.Background(), chunk)
		assert.EqualError(t, err, ErrInvalidUploadSession.Error())
	})

	t.Run("With a new session ID", func(t *testing.T) {
		uNew := uuid.NewV4()
		chunk := &pb.FileChunk{
			Id:   uNew.Bytes(),
			Data: []byte("coucou, "),
		}

		_, err = server.UploadChunk(context.Background(), chunk)
		assert.EqualError(t, err, ErrUploadSessionNotFound.Error())
	})
}

func TestGRPCServer_EndUpload(t *testing.T) {
	filename := fmt.Sprintf("grpc_test_%d", time.Now().UnixNano())
	u := uuid.NewV4()

	file, err := os.Create(filepath.Join(storagePath, filename))
	assert.NoError(t, err, "os.Create()")

	var indexedFile *os.File

	content := []byte("coucou, tu veux voir mon fichier ?")
	_, err = file.Write(content)
	assert.NoError(t, err, "file.Write()")

	server := &grpcServer{
		indexFile: func(ctx context.Context, file *os.File) ([]byte, error) {
			indexedFile = file
			return []byte("123"), nil
		},
		storagePath: storagePath,
		sessions: map[uuid.UUID]*os.File{
			u: file,
		},
	}

	t.Run("With a valid session ID", func(t *testing.T) {
		req := &pb.UploadSession{
			Id: u.Bytes(),
		}

		uploadAck, err := server.EndUpload(context.Background(), req)
		assert.NoError(t, err, "server.EndUpload()")
		assert.Equal(t, uploadAck.FileHash, []byte("123"))

		got, err := indexedFile.Stat()
		assert.NoError(t, err, "indexedFile.Stat()")

		f, err := os.Open(filepath.Join(storagePath, filename))
		assert.NoError(t, err, "os.Open()")

		expect, err := f.Stat()
		assert.NoError(t, err, "f.Stat()")
		assert.True(t, os.SameFile(expect, got), "SameFile")

		err = os.Remove(filepath.Join(storagePath, filename))
		assert.NoError(t, err, " Remove()")
	})

	t.Run("With an invalid session ID", func(t *testing.T) {
		req := &pb.UploadSession{
			Id: []byte("123"),
		}

		_, err = server.EndUpload(context.Background(), req)
		assert.EqualError(t, err, ErrInvalidUploadSession.Error())
	})

	t.Run("With a new session ID", func(t *testing.T) {
		uNew := uuid.NewV4()
		req := &pb.UploadSession{
			Id: uNew.Bytes(),
		}

		_, err = server.EndUpload(context.Background(), req)
		assert.EqualError(t, err, ErrUploadSessionNotFound.Error())
	})
}

func TestGRPCServer_AuthorizePeers(t *testing.T) {
	var fh []byte
	var pids [][]byte
	server := &grpcServer{
		authorize: func(ctx context.Context, peerIds [][]byte, fileHash []byte) error {
			fh = fileHash
			pids = peerIds
			return nil
		},
	}

	fileHash := []byte("file hash")
	peerIds := [][]byte{
		[]byte("QmVhJVRSYHNSHgR9dJNbDxu6G7GPPqJAeiJoVRvcexGNf1"),
		[]byte("QmVhJVRSYHNSHgR9dJNbDxu6G7GPPqJAeiJoVRvcexGNf2"),
		[]byte("QmVhJVRSYHNSHgR9dJNbDxu6G7GPPqJAeiJoVRvcexGNf3"),
	}

	server.AuthorizePeers(context.Background(), &pb.AuthRequest{
		FileHash: fileHash,
		PeerIds:  peerIds,
	})

	assert.Equal(t, fileHash, fh, "FileHash")
	assert.Equal(t, peerIds, pids, "FileHash")
}

func checkFileContent(t *testing.T, filename string, expected []byte) *os.File {
	f, err := os.Open(filepath.Join(storagePath, filename))
	assert.NoError(t, err, "os.Open()")

	content := make([]byte, 42)
	zeros := make([]byte, 42-len(expected))
	_, err = f.Read(content)
	assert.NoError(t, err, "ReadFile()")
	assert.Equal(t, expected, content[:len(expected)], "ReadFile()")
	assert.Equal(t, zeros, content[len(expected):], "ReadFile()")

	return f
}
