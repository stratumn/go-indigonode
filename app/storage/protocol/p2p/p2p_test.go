// Copyright © 2017-2018 Stratumn SAS
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as
// published by the Free Software Foundation, either version 3 of the
// License, or (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <https://www.gnu.org/licenses/>.

package p2p

import (
	"context"
	"io"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/pkg/errors"
	uuid "github.com/satori/go.uuid"
	"github.com/stratumn/go-node/app/storage/pb"
	"github.com/stratumn/go-node/app/storage/protocol/constants"
	"github.com/stratumn/go-node/app/storage/protocol/file"
	"github.com/stratumn/go-node/app/storage/protocol/file/mockhandler"
	"github.com/stratumn/go-node/app/storage/protocol/p2p/mockencoder"
	p2pcore "github.com/stratumn/go-node/core/p2p"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	inet "gx/ipfs/QmZNJyx9GGCX4GeuHnLB8fxaxMLs4MjTjHokxfQcCd6Nve/go-libp2p-net"
	swarmtesting "gx/ipfs/QmeDpqUwwdye8ABKVMPXKuWwPVURFdqTqssbTUB39E2Nwd/go-libp2p-swarm/testing"
	protobuf "gx/ipfs/QmewJ1Zp9Hwz5HcMd7JYjhLXwvEHTL2UBCCz3oLt1E2N5z/go-multicodec/protobuf"
)

func TestP2P_PullFile(t *testing.T) {
	ctx := context.Background()
	h1 := p2pcore.NewHost(ctx, swarmtesting.GenSwarm(t, ctx))
	h2 := p2pcore.NewHost(ctx, swarmtesting.GenSwarm(t, ctx))
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
		defer inet.FullClose(stream)
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
