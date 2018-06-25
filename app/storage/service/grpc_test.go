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
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/pkg/errors"
	uuid "github.com/satori/go.uuid"
	grpcpb "github.com/stratumn/go-indigonode/app/storage/grpc"
	"github.com/stratumn/go-indigonode/app/storage/grpc/mockstorage"
	"github.com/stratumn/go-indigonode/app/storage/pb"
	"github.com/stretchr/testify/assert"
)

func TestGRPCServer_UploadFile(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	stream := mockstorage.NewMockStorage_UploadServer(ctrl)
	endWriteCalled := false
	filename := "grpc_test"
	uid := uuid.NewV4()
	receivedContent := []byte{}

	server := &grpcServer{
		beginWrite: func(ctx context.Context, name string) (uuid.UUID, error) {
			assert.Equal(t, filename, name, "beginWrite")
			return uid, nil
		},
		writeChunk: func(ctx context.Context, id uuid.UUID, data []byte) error {

			assert.Equal(t, uid, id, "writeChunk")
			receivedContent = append(receivedContent, data...)

			return nil
		},
		endWrite: func(ctx context.Context, id uuid.UUID) ([]byte, error) {
			endWriteCalled = true
			assert.Equal(t, uid, id, "writeChunk")
			return []byte("123"), nil
		},
		uploadTimeout: 1 * time.Second,
	}

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

	assert.Equal(t, []byte("coucou, tu veux voir mon fichier ?"), receivedContent, "received content incorrect")
	assert.True(t, endWriteCalled, "endWrite must be called")
}

func TestGRPCServer_UploadFile_Fail(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	stream := mockstorage.NewMockStorage_UploadServer(ctrl)
	endWriteCalled := false
	abortWriteCalled := false
	filename := "grpc_test"
	uid := uuid.NewV4()
	receivedContent := []byte{}

	server := &grpcServer{
		beginWrite: func(ctx context.Context, name string) (uuid.UUID, error) {
			assert.Equal(t, filename, name, "beginWrite")
			return uid, nil
		},
		writeChunk: func(ctx context.Context, id uuid.UUID, data []byte) error {

			assert.Equal(t, uid, id, "writeChunk")
			receivedContent = append(receivedContent, data...)

			return nil
		},
		endWrite: func(ctx context.Context, id uuid.UUID) ([]byte, error) {
			endWriteCalled = true
			return []byte("123"), nil
		},
		abortWrite: func(ctx context.Context, id uuid.UUID) error {
			abortWriteCalled = true
			assert.Equal(t, uid, id, "abortWrite")
			return nil
		},
		uploadTimeout: 1 * time.Second,
	}

	chunk1 := &pb.FileChunk{
		FileName: filename,
		Data:     []byte("coucou, "),
	}

	gomock.InOrder(
		stream.EXPECT().Recv().Return(chunk1, nil),
		stream.EXPECT().Recv().Return(nil, errors.New("https://goo.gl/YMfBcQ")),
	)

	err := server.Upload(stream)
	assert.Error(t, err, "server.Upload()")

	assert.False(t, endWriteCalled, "endWrite must not be called")
	assert.True(t, abortWriteCalled, "abortWrite must be called")
}

func TestGRPCServer_StartUpload(t *testing.T) {
	beginCalled := false
	uid := uuid.NewV4()
	filename := "grpc_test"

	server := &grpcServer{
		beginWrite: func(ctx context.Context, name string) (uuid.UUID, error) {
			beginCalled = true
			assert.Equal(t, filename, name, "beginWrite")
			return uid, nil
		},
	}

	uploadReq := &grpcpb.UploadReq{
		FileName: filename,
	}

	rsp, err := server.StartUpload(context.Background(), uploadReq)
	assert.NoError(t, err, "server.StartUpload()")

	u, err := uuid.FromBytes(rsp.Id)
	assert.NoError(t, err, "server.StartUpload()")
	assert.Equal(t, uid, u, "rsp.Id")

	assert.True(t, beginCalled, "beginWrite should be called")
}

func TestGRPCServer_UploadChunk(t *testing.T) {
	u := uuid.NewV4()
	content := []byte("some data")

	server := &grpcServer{
		writeChunk: func(ctx context.Context, id uuid.UUID, data []byte) error {
			assert.Equal(t, u, id, "writeChunk")
			assert.Equal(t, content, data, "writeChunk")

			return nil
		},
	}

	chunk := &grpcpb.SessionFileChunk{
		Id:   u.Bytes(),
		Data: content,
	}

	_, err := server.UploadChunk(context.Background(), chunk)
	assert.NoError(t, err, "UploadChunk")
}

func TestGRPCServer_EndUpload(t *testing.T) {
	u := uuid.NewV4()
	hash := []byte("yolo")

	server := &grpcServer{
		endWrite: func(ctx context.Context, id uuid.UUID) ([]byte, error) {
			assert.Equal(t, u, id, "writeChunk")
			return hash, nil
		},
	}

	req := &grpcpb.UploadSession{
		Id: u.Bytes(),
	}

	uploadAck, err := server.EndUpload(context.Background(), req)
	assert.NoError(t, err, "server.EndUpload()")
	assert.Equal(t, uploadAck.FileHash, hash)
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

	server.AuthorizePeers(context.Background(), &grpcpb.AuthRequest{
		FileHash: fileHash,
		PeerIds:  peerIds,
	})

	assert.Equal(t, fileHash, fh, "FileHash")
	assert.Equal(t, peerIds, pids, "PeerIds")
}

func TestGRPCServer_Download(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	var fh []byte
	var pid []byte

	server := &grpcServer{
		download: func(ctx context.Context, fileHash []byte, peerId []byte) error {
			fh = fileHash
			pid = peerId
			return nil
		},
		readChunks: func(ctx context.Context, fileHash []byte, chunkSize int, cr *chunkReader) error {
			err := cr.OnChunk([]byte("123"), "blah.pdf")
			assert.NoError(t, err, "cr.OnChunk()")

			return cr.OnChunk([]byte("456"), "blah.pdf")
		},
	}

	fileHash := []byte("file hash")
	peerID := []byte("QmVhJVRSYHNSHgR9dJNbDxu6G7GPPqJAeiJoVRvcexGNf1")

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ss := mockstorage.NewMockStorage_DownloadServer(ctrl)

	ss.EXPECT().Context().Return(ctx).AnyTimes()
	gomock.InOrder(
		ss.EXPECT().Send(&pb.FileChunk{
			FileName: "blah.pdf",
			Data:     []byte("123"),
		}),
		ss.EXPECT().Send(&pb.FileChunk{
			Data: []byte("456"),
		}),
	)

	server.Download(&grpcpb.DownloadRequest{
		FileHash: fileHash,
		PeerId:   peerID,
	}, ss)

	assert.Equal(t, fileHash, fh, "FileHash")
	assert.Equal(t, peerID, pid, "PeerId")
}
