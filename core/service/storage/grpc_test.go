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
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	pb "github.com/stratumn/alice/grpc/storage"
	"github.com/stratumn/alice/grpc/storage/mockstorage"
	"github.com/stretchr/testify/assert"
)

func TestGRPCServer_UploadFile(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	stream := mockstorage.NewMockStorage_UploadServer(ctrl)

	var indexedFile *os.File
	server := &grpcServer{
		indexFile: func(ctx context.Context, file *os.File) error {
			indexedFile = file
			return nil
		},
		storagePath: "/tmp",
	}
	filename := fmt.Sprintf("grpc_test_%d", time.Now().UnixNano())

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
		stream.EXPECT().SendAndClose(&pb.Ack{}).Return(nil),
	)

	err := server.Upload(stream)
	assert.NoError(t, err, "server.Upload()")

	// Check that file has the right content.
	f, err := os.Open(fmt.Sprintf("/tmp/%s", filename))
	assert.NoError(t, err, "os.Open()")
	expected := []byte("coucou, tu veux voir mon fichier ?")
	content := make([]byte, 42)
	zeros := make([]byte, 42-len(expected))
	_, err = f.Read(content)
	assert.NoError(t, err, "ReadFile()")
	assert.Equal(t, expected, content[:len(expected)], "ReadFile()")
	assert.Equal(t, zeros, content[len(expected):], "ReadFile()")

	// Check that the file was correctly sent to be indexed.
	expect, err := f.Stat()
	assert.NoError(t, err, "f.Stat()")
	got, err := indexedFile.Stat()
	assert.NoError(t, err, "indexedFile.Stat()")

	assert.True(t, os.SameFile(expect, got), "SameFile")

	err = os.Remove(fmt.Sprintf("/tmp/%s", filename))
	assert.NoError(t, err, " Remove()")
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