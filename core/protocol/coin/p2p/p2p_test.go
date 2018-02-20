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

	"github.com/pkg/errors"
	p2pcore "github.com/stratumn/alice/core/p2p"
	pb "github.com/stratumn/alice/pb/coin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	inet "gx/ipfs/QmQm7WmgYCa4RSz76tKEYpRjApjnRw8ZTUVQC15b8JM4a2/go-libp2p-net"
	protobuf "gx/ipfs/QmRDePEiL4Yupq5EkcK3L3ko3iMgYaqUdLu7xc1kqs7dnV/go-multicodec/protobuf"
	testutil "gx/ipfs/QmV1axkk86DDkYwS269AvPy9eV5h7mUyHveJkSVHPjrQtY/go-libp2p-netutil"
)

func TestP2PRequestsHandler(t *testing.T) {
	ctx := context.Background()
	h1 := p2pcore.NewHost(ctx, testutil.GenSwarmNetwork(t, ctx))
	h2 := p2pcore.NewHost(ctx, testutil.GenSwarmNetwork(t, ctx))
	defer h1.Close()
	defer h2.Close()

	// connect h1 to h2
	h2pi := h2.Peerstore().PeerInfo(h2.ID())
	require.NoError(t, h1.Connect(ctx, h2pi), "Connect()")

	p2p := NewP2P(h1, "protocol")

	t.Run("RequestHeaderByHash/success", func(t *testing.T) {
		header := &pb.Header{Nonce: 42}
		rsp := &pb.Response{Msg: &pb.Response_HeaderRsp{HeaderRsp: header}}
		ctx := context.Background()
		defer ctx.Done()
		h2.SetStreamHandler("protocol", getStreamHandler(ctx, rsp))

		r, err := p2p.RequestHeaderByHash(ctx, h2.ID(), []byte("zoulou"))
		assert.NoError(t, err, "RequestHeaderByHash()")
		assert.Equal(t, header, r, "RequestHeaderByHash()")

	})

	t.Run("RequestHeaderByHash/bad-response-type", func(t *testing.T) {
		rsp := &pb.Response{}
		ctx := context.Background()
		defer ctx.Done()
		h2.SetStreamHandler("protocol", getStreamHandler(ctx, rsp))

		_, err := p2p.RequestHeaderByHash(ctx, h2.ID(), []byte("zoulou"))
		assert.EqualError(t, err, pb.ErrBadResponseType.Error(), "RequestHeaderByHash()")
	})

	t.Run("RequestHeadersByNumber/success", func(t *testing.T) {
		headers := []*pb.Header{
			&pb.Header{Nonce: 42},
			&pb.Header{Nonce: 43},
		}
		rsp := &pb.Response{Msg: &pb.Response_HeadersRsp{HeadersRsp: &pb.Headers{Headers: headers}}}
		ctx := context.Background()
		defer ctx.Done()
		h2.SetStreamHandler("protocol", getStreamHandler(ctx, rsp))

		r, err := p2p.RequestHeadersByNumber(ctx, h2.ID(), 1, 1)
		assert.NoError(t, err, "RequestHeadersByNumber()")
		assert.Equal(t, headers, r, "RequestHeadersByNumber()")

	})

	t.Run("RequestHeadersByNumber/bad-response-type", func(t *testing.T) {
		rsp := &pb.Response{}
		ctx := context.Background()
		defer ctx.Done()
		h2.SetStreamHandler("protocol", getStreamHandler(ctx, rsp))

		_, err := p2p.RequestHeadersByNumber(ctx, h2.ID(), 1, 2)
		assert.EqualError(t, err, pb.ErrBadResponseType.Error(), "RequestHeadersByNumber()")
	})

	t.Run("RequestBlockByHash/success", func(t *testing.T) {
		block := &pb.Block{Header: &pb.Header{Nonce: 42}}
		rsp := &pb.Response{Msg: &pb.Response_BlockRsp{BlockRsp: block}}
		ctx := context.Background()
		defer ctx.Done()
		h2.SetStreamHandler("protocol", getStreamHandler(ctx, rsp))

		r, err := p2p.RequestBlockByHash(ctx, h2.ID(), []byte("zoulou"))
		assert.NoError(t, err, "RequestBlockByHash()")
		assert.Equal(t, block, r, "RequestBlockByHash()")

	})

	t.Run("RequestBlockByHash/bad-response-type", func(t *testing.T) {
		rsp := &pb.Response{}
		ctx := context.Background()
		defer ctx.Done()
		h2.SetStreamHandler("protocol", getStreamHandler(ctx, rsp))

		_, err := p2p.RequestBlockByHash(ctx, h2.ID(), []byte("zoulou"))
		assert.EqualError(t, err, pb.ErrBadResponseType.Error(), "RequestBlockByHash()")
	})

	t.Run("RequestBlocksByNumber/success", func(t *testing.T) {
		blocks := []*pb.Block{
			&pb.Block{Header: &pb.Header{Nonce: 42}},
			&pb.Block{Header: &pb.Header{Nonce: 43}},
		}
		rsp := &pb.Response{Msg: &pb.Response_BlocksRsp{BlocksRsp: &pb.Blocks{Blocks: blocks}}}
		ctx := context.Background()
		defer ctx.Done()
		h2.SetStreamHandler("protocol", getStreamHandler(ctx, rsp))

		r, err := p2p.RequestBlocksByNumber(ctx, h2.ID(), 1, 1)
		assert.NoError(t, err, "RequestBlocksByNumber()")
		assert.Equal(t, blocks, r, "RequestBlocksByNumber()")

	})

	t.Run("RequestBlocksByNumber/bad-response-type", func(t *testing.T) {
		rsp := &pb.Response{}
		ctx := context.Background()
		defer ctx.Done()
		h2.SetStreamHandler("protocol", getStreamHandler(ctx, rsp))

		_, err := p2p.RequestBlocksByNumber(ctx, h2.ID(), 1, 2)
		assert.EqualError(t, err, pb.ErrBadResponseType.Error(), "RequestBlocksByNumber()")
	})

}

// Returns a tream handler that always returns response
func getStreamHandler(ctx context.Context, response *pb.Response) func(inet.Stream) {

	return func(stream inet.Stream) {
		dec := protobuf.Multicodec(nil).Decoder(stream)
		enc := protobuf.Multicodec(nil).Encoder(stream)
		ch := make(chan error, 1)

		go func() {
			var req pb.Request
			err := dec.Decode(&req)
			if err == io.EOF {
				return
			}
			if err != nil {
				ch <- errors.WithStack(err)
				return
			}

			err = enc.Encode(response)
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
