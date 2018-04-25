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
	"fmt"
	"io"
	"os"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/pkg/errors"
	p2pcore "github.com/stratumn/alice/core/p2p"
	"github.com/stratumn/alice/core/protocol/storage/constants"
	"github.com/stratumn/alice/core/protocol/storage/file/mockhandler"
	"github.com/stratumn/alice/core/protocol/storage/p2p/mockencoder"
	pb "github.com/stratumn/alice/pb/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	protobuf "gx/ipfs/QmRDePEiL4Yupq5EkcK3L3ko3iMgYaqUdLu7xc1kqs7dnV/go-multicodec/protobuf"
	inet "gx/ipfs/QmXfkENeeBvh3zYA51MaSdGUdBjhQ99cP5WQe8zgr6wchG/go-libp2p-net"
	testutil "gx/ipfs/QmYVR3C8DWPHdHxvLtNFYfjsXgaRAdh6hPMNH3KiwCgu4o/go-libp2p-netutil"
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
	data := []byte("hello there")

	t.Run("successfully-write-file", func(t *testing.T) {
		fileHandler.EXPECT().WriteFile(gomock.Any(), gomock.Any()).
			DoAndReturn(func(_ context.Context, ch <-chan *pb.FileChunk) (interface{}, error) {
				first := true
				for _, b := range data {
					chunk := <-ch
					if first {
						assert.Equal(t, fileName, chunk.FileName)
						first = false
					}
					assert.Equal(t, []byte{b}, chunk.Data, "chunk.Data")
				}

				return nil, nil
			})

		p2p := NewP2P(h1, fileHandler)

		h2.SetStreamHandler(constants.ProtocolID, getStreamHandler(ctx, fileName, data))

		err := p2p.PullFile(ctx, []byte("fileHash"), h2.ID())
		assert.NoError(t, err, "PullFile")
	})

	t.Run("fail-write-file", func(t *testing.T) {
		fileHandler.EXPECT().WriteFile(gomock.Any(), gomock.Any()).
			DoAndReturn(func(_ context.Context, ch <-chan *pb.FileChunk) (interface{}, error) {
				chunk := <-ch
				assert.Equal(t, fileName, chunk.FileName)
				assert.Equal(t, []byte{data[0]}, chunk.Data, "chunk.Data")

				return nil, errors.New("https://goo.gl/YMfBcQ")
			})

		p2p := NewP2P(h1, fileHandler)

		h2.SetStreamHandler(constants.ProtocolID, getStreamHandler(ctx, fileName, data))

		err := p2p.PullFile(ctx, []byte("fileHash"), h2.ID())
		assert.Error(t, err, "PullFile")
	})

}

func TestP2P_SendFile(t *testing.T) {

	content := []byte("Who wants to download my juicy file ?")
	storagePath := "/tmp/"
	fileName := fmt.Sprintf("TestP2P_SendFile-%d", time.Now().UnixNano())
	filePath := storagePath + fileName

	f, err := os.Create(filePath)
	require.NoError(t, err, "os.Create")

	_, err = f.Write(content)
	require.NoError(t, err, "f.Write")

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	enc := mockencoder.NewMockEncoder(ctrl)

	p2p := &p2p{
		chunkSize: 10,
	}

	gomock.InOrder(
		enc.EXPECT().Encode(&pb.FileChunk{FileName: fileName, Data: content[0:10]}),
		enc.EXPECT().Encode(&pb.FileChunk{Data: content[10:20]}),
		enc.EXPECT().Encode(&pb.FileChunk{Data: content[20:30]}),
		enc.EXPECT().Encode(&pb.FileChunk{Data: content[30:37]}),
		enc.EXPECT().Encode(&pb.FileChunk{Data: nil}),
	)

	err = p2p.SendFile(context.Background(), enc, filePath)
	assert.NoError(t, err, "SendFile")
}

func TestP2P_SendFile_ExactChunkSize(t *testing.T) {

	content := []byte("0123456789")
	storagePath := "/tmp/"
	fileName := fmt.Sprintf("TestP2P_SendFile-%d", time.Now().UnixNano())
	filePath := storagePath + fileName

	f, err := os.Create(filePath)
	require.NoError(t, err, "os.Create")

	_, err = f.Write(content)
	require.NoError(t, err, "f.Write")

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	enc := mockencoder.NewMockEncoder(ctrl)

	p2p := &p2p{
		chunkSize: 10,
	}

	gomock.InOrder(
		enc.EXPECT().Encode(&pb.FileChunk{FileName: fileName, Data: content}),
		enc.EXPECT().Encode(&pb.FileChunk{Data: nil}),
	)

	err = p2p.SendFile(context.Background(), enc, filePath)
	assert.NoError(t, err, "SendFile")
}

// Returns a stream handler that streams the byte array.
func getStreamHandler(ctx context.Context, name string, data []byte) func(inet.Stream) {

	return func(stream inet.Stream) {
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
			for _, b := range data {
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
