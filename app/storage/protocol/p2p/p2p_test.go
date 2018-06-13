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

package p2p

import (
	"context"
	"io"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/pkg/errors"
	uuid "github.com/satori/go.uuid"
	"github.com/stratumn/alice/app/storage/pb"
	"github.com/stratumn/alice/app/storage/protocol/constants"
	"github.com/stratumn/alice/app/storage/protocol/file"
	"github.com/stratumn/alice/app/storage/protocol/file/mockhandler"
	"github.com/stratumn/alice/app/storage/protocol/p2p/mockencoder"
	p2pcore "github.com/stratumn/alice/core/p2p"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	protobuf "gx/ipfs/QmRDePEiL4Yupq5EkcK3L3ko3iMgYaqUdLu7xc1kqs7dnV/go-multicodec/protobuf"
	inet "gx/ipfs/QmXoz9o2PT3tEzf7hicegwex5UgVP54n3k82K7jrWFyN86/go-libp2p-net"
	testutil "gx/ipfs/Qmb6BsZf6Y3kxffXMNTubGPF1w1bkHtpvhfYbmnwP3NQyw/go-libp2p-netutil"
)

func TestP2P_PullFile(t *testing.T) {
	ctx := context.Background()
	h1 := p2pcore.NewHost(ctx, testutil.GenSwarmNetwork(t, ctx))
	h2 := p2pcore.NewHost(ctx, testutil.GenSwarmNetwork(t, ctx))
	defer h1.Close()
	defer h2.Close()

	// connect h1 to h2
	h2pi := h2.Peerstore().PeerInfo(h2.ID())
	require.NoError(t, h1.Connect(ctx, h2pi), "Connect()")

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	fileHandler := mockhandler.NewMockHandler(ctrl)

	fileName := "filename"
	data := []byte("hello")

	uid := uuid.NewV4()
	p2p := NewP2P(h1, fileHandler)

	t.Run("successfully-write-file", func(t *testing.T) {

		gomock.InOrder(
			fileHandler.EXPECT().BeginWrite(gomock.Any(), fileName).Return(uid, nil),
			fileHandler.EXPECT().WriteChunk(gomock.Any(), uid, []byte("h")).Return(nil),
			fileHandler.EXPECT().WriteChunk(gomock.Any(), uid, []byte("e")).Return(nil),
			fileHandler.EXPECT().WriteChunk(gomock.Any(), uid, []byte("l")).Return(nil),
			fileHandler.EXPECT().WriteChunk(gomock.Any(), uid, []byte("l")).Return(nil),
			fileHandler.EXPECT().WriteChunk(gomock.Any(), uid, []byte("o")).Return(nil),
			fileHandler.EXPECT().EndWrite(gomock.Any(), uid).Return(nil, nil),
		)

		h2.SetStreamHandler(constants.ProtocolID, getStreamHandler(ctx, fileName, data, 0))

		err := p2p.PullFile(ctx, []byte("fileHash"), h2.ID())
		assert.NoError(t, err, "PullFile")
	})

	t.Run("fail-write-file", func(t *testing.T) {

		gomock.InOrder(
			fileHandler.EXPECT().BeginWrite(gomock.Any(), fileName).Return(uid, nil),
			fileHandler.EXPECT().WriteChunk(gomock.Any(), uid, []byte("h")).Return(nil),
			fileHandler.EXPECT().WriteChunk(gomock.Any(), uid, []byte("e")).Return(errors.New("https://goo.gl/YMfBcQ")),
		)

		h2.SetStreamHandler(constants.ProtocolID, getStreamHandler(ctx, fileName, data, 0))

		err := p2p.PullFile(ctx, []byte("fileHash"), h2.ID())
		assert.Error(t, err, "PullFile")
	})

	t.Run("fail-receive-chunk", func(t *testing.T) {

		gomock.InOrder(
			fileHandler.EXPECT().BeginWrite(gomock.Any(), fileName).Return(uid, nil),
			fileHandler.EXPECT().WriteChunk(gomock.Any(), uid, []byte("h")).Return(nil),
			fileHandler.EXPECT().AbortWrite(gomock.Any(), uid).Return(nil),
		)

		h2.SetStreamHandler(constants.ProtocolID, getStreamHandler(ctx, fileName, data, 1))

		err := p2p.PullFile(ctx, []byte("fileHash"), h2.ID())
		assert.Error(t, err, "PullFile")
	})

}

func TestP2P_SendFile(t *testing.T) {
	fileHash := []byte("file hash")
	fileName := "yolo"
	filePath := "/the/path/to/" + fileName
	chunkSize := 42

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	enc := mockencoder.NewMockEncoder(ctrl)
	fileHandler := mockhandler.NewMockHandler(ctrl)

	p2p := &p2p{
		fileHandler: fileHandler,
		chunkSize:   chunkSize,
	}

	chunk1 := []byte("who wants to download ")

	gomock.InOrder(
		fileHandler.EXPECT().ReadChunks(gomock.Any(), fileHash, gomock.Any(), gomock.Any()).Do(func(ctx context.Context, hash []byte, size int, r file.Reader) {
			assert.Equal(t, chunkSize, size, "chunkSize")
			assert.Equal(t, fileHash, hash, "chunkSize")
			r.OnChunk(chunk1, filePath)
		}),
		enc.EXPECT().Encode(&pb.FileChunk{Data: chunk1, FileName: fileName}),
		enc.EXPECT().Encode(&pb.FileChunk{Data: nil}),
	)

	err := p2p.SendFile(context.Background(), enc, fileHash)
	assert.NoError(t, err, "SendFile")
}

// Returns a stream handler that streams the byte array.
// if failAfter > 0, the transmission will fail after failAfter messages
func getStreamHandler(ctx context.Context, name string, data []byte, failAfter int) func(inet.Stream) {

	return func(stream inet.Stream) {
		defer stream.Close()
		dec := protobuf.Multicodec(nil).Decoder(stream)
		enc := protobuf.Multicodec(nil).Encoder(stream)
		ch := make(chan error, 1)

		go func() {
			var req pb.FileInfo
			err := dec.Decode(&req)
			if err == io.EOF {
				return
			}
			if err != nil {
				ch <- errors.WithStack(err)
				return
			}

			first := true
			for i, b := range data {
				if failAfter > 0 && i == failAfter {
					ch <- errors.New("https://goo.gl/YMfBcQ")
					return
				}

				rsp := &pb.FileChunk{Data: []byte{b}}
				if first {
					rsp.FileName = name
					first = false
				}
				err = enc.Encode(rsp)
				if err != nil {
					ch <- errors.WithStack(err)
					return
				}
			}

			// Send empty chunk.
			err = enc.Encode(&pb.FileChunk{})
			if err != nil {
				ch <- errors.WithStack(err)
				return
			}

			ch <- nil
		}()

		select {
		case <-ctx.Done():
			return

		case err := <-ch:
			if err != nil {
				return
			}
		}
	}
}