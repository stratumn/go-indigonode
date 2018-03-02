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
	p2pcore "github.com/stratumn/alice/core/p2p"
	"github.com/stratumn/alice/core/protocol/coin/chain"
	"github.com/stratumn/alice/core/protocol/coin/chain/mockchain"
	"github.com/stratumn/alice/core/protocol/coin/p2p/mockencoder"
	pb "github.com/stratumn/alice/pb/coin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	protobuf "gx/ipfs/QmRDePEiL4Yupq5EkcK3L3ko3iMgYaqUdLu7xc1kqs7dnV/go-multicodec/protobuf"
	inet "gx/ipfs/QmXfkENeeBvh3zYA51MaSdGUdBjhQ99cP5WQe8zgr6wchG/go-libp2p-net"
	testutil "gx/ipfs/QmYVR3C8DWPHdHxvLtNFYfjsXgaRAdh6hPMNH3KiwCgu4o/go-libp2p-netutil"
	protocol "gx/ipfs/QmZNkThpqfVXs9GNbexPrfBbXSLNYeKrE7jwFM2oqHbyqN/go-libp2p-protocol"
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

	t.Run("RequestHeaderByHash", func(t *testing.T) {
		t.Run("success", func(t *testing.T) {
			header := &pb.Header{Nonce: 42}
			rsp := &pb.Response{Msg: &pb.Response_HeaderRsp{HeaderRsp: header}}
			ctx := context.Background()
			defer ctx.Done()
			h2.SetStreamHandler("protocol", getSingleResponseStreamHandler(ctx, rsp))

			r, err := p2p.RequestHeaderByHash(ctx, h2.ID(), []byte("zoulou"))
			assert.NoError(t, err, "RequestHeaderByHash()")
			assert.Equal(t, header, r, "RequestHeaderByHash()")
		})
		t.Run("bad-response-type", func(t *testing.T) {
			rsp := &pb.Response{}
			ctx := context.Background()
			defer ctx.Done()
			h2.SetStreamHandler("protocol", getSingleResponseStreamHandler(ctx, rsp))

			_, err := p2p.RequestHeaderByHash(ctx, h2.ID(), []byte("zoulou"))
			assert.EqualError(t, err, pb.ErrBadResponseType.Error(), "RequestHeaderByHash()")
		})
	})

	t.Run("RequestHeadersByNumber", func(t *testing.T) {
		t.Run("success", func(t *testing.T) {
			headers := []*pb.Header{
				&pb.Header{Nonce: 42},
				&pb.Header{Nonce: 43},
			}
			rsp := &pb.Response{Msg: &pb.Response_HeadersRsp{HeadersRsp: &pb.Headers{Headers: headers}}}
			ctx := context.Background()
			defer ctx.Done()
			h2.SetStreamHandler("protocol", getSingleResponseStreamHandler(ctx, rsp))

			r, err := p2p.RequestHeadersByNumber(ctx, h2.ID(), 1, 1)
			assert.NoError(t, err, "RequestHeadersByNumber()")
			assert.Equal(t, headers, r, "RequestHeadersByNumber()")
		})
		t.Run("bad-response-type", func(t *testing.T) {
			rsp := &pb.Response{}
			ctx := context.Background()
			defer ctx.Done()
			h2.SetStreamHandler("protocol", getSingleResponseStreamHandler(ctx, rsp))

			_, err := p2p.RequestHeadersByNumber(ctx, h2.ID(), 1, 2)
			assert.EqualError(t, err, pb.ErrBadResponseType.Error(), "RequestHeadersByNumber()")
		})
	})

	t.Run("RequestBlockByHash", func(t *testing.T) {
		t.Run("RequestBlockByHash/success", func(t *testing.T) {
			block := &pb.Block{Header: &pb.Header{Nonce: 42}}
			rsp := &pb.Response{Msg: &pb.Response_BlockRsp{BlockRsp: block}}
			ctx := context.Background()
			defer ctx.Done()
			h2.SetStreamHandler("protocol", getSingleResponseStreamHandler(ctx, rsp))

			r, err := p2p.RequestBlockByHash(ctx, h2.ID(), []byte("zoulou"))
			assert.NoError(t, err, "RequestBlockByHash()")
			assert.Equal(t, block, r, "RequestBlockByHash()")
		})
		t.Run("bad-response-type", func(t *testing.T) {
			rsp := &pb.Response{}
			ctx := context.Background()
			defer ctx.Done()
			h2.SetStreamHandler("protocol", getSingleResponseStreamHandler(ctx, rsp))

			_, err := p2p.RequestBlockByHash(ctx, h2.ID(), []byte("zoulou"))
			assert.EqualError(t, err, pb.ErrBadResponseType.Error(), "RequestBlockByHash()")
		})
	})

	t.Run("RequestBlocksByNumber", func(t *testing.T) {
		t.Run("success", func(t *testing.T) {
			blocks := []*pb.Block{
				&pb.Block{Header: &pb.Header{Nonce: 42}},
				&pb.Block{Header: &pb.Header{Nonce: 43}},
			}
			rsp := &pb.Response{Msg: &pb.Response_BlocksRsp{BlocksRsp: &pb.Blocks{Blocks: blocks}}}
			ctx := context.Background()
			defer ctx.Done()
			h2.SetStreamHandler("protocol", getSingleResponseStreamHandler(ctx, rsp))

			r, err := p2p.RequestBlocksByNumber(ctx, h2.ID(), 1, 1)
			assert.NoError(t, err, "RequestBlocksByNumber()")
			assert.Equal(t, blocks, r, "RequestBlocksByNumber()")
		})
		t.Run("bad-response-type", func(t *testing.T) {
			rsp := &pb.Response{}
			ctx := context.Background()
			defer ctx.Done()
			h2.SetStreamHandler("protocol", getSingleResponseStreamHandler(ctx, rsp))

			_, err := p2p.RequestBlocksByNumber(ctx, h2.ID(), 1, 2)
			assert.EqualError(t, err, pb.ErrBadResponseType.Error(), "RequestBlocksByNumber()")
		})
	})
}

func TestP2PResponsesHandler(t *testing.T) {
	ctx := context.Background()
	h1 := p2pcore.NewHost(ctx, testutil.GenSwarmNetwork(t, ctx))

	protocolID := protocol.ID("proptocolID")
	p2p := NewP2P(h1, protocolID)

	t.Run("RespondHeaderByHash", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		ch := mockchain.NewMockReader(ctrl)
		enc := mockencoder.NewMockEncoder(ctrl)

		req := &pb.HeaderRequest{Hash: []byte("zoulou")}
		h := &pb.Header{Nonce: 42}
		ch.EXPECT().GetHeaderByHash(gomock.Any()).Return(h, nil).Times(1)
		enc.EXPECT().Encode(pb.NewHeaderResponse(h)).Return(nil).Times(1)

		err := p2p.RespondHeaderByHash(ctx, req, enc, ch)
		assert.NoError(t, err, "RespondHeaderByHash()")
	})

	t.Run("RespondHeadersByNumber", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		ch := mockchain.NewMockReader(ctrl)
		enc := mockencoder.NewMockEncoder(ctrl)

		req := &pb.HeadersRequest{From: 3, Amount: 42}
		h := &pb.Header{Nonce: 42}
		ch.EXPECT().GetHeaderByNumber(uint64(3)).Return(h, nil).Times(1)
		ch.EXPECT().GetHeaderByNumber(uint64(4)).Return(h, nil).Times(1)
		ch.EXPECT().GetHeaderByNumber(uint64(5)).Return(nil, chain.ErrBlockNotFound).Times(1)
		enc.EXPECT().Encode(pb.NewHeadersResponse([]*pb.Header{h, h})).Return(nil).Times(1)

		err := p2p.RespondHeadersByNumber(ctx, req, enc, ch)
		assert.NoError(t, err, "RespondHeaderByHash()")
	})

	t.Run("RespondBlockByHash", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		ch := mockchain.NewMockReader(ctrl)
		enc := mockencoder.NewMockEncoder(ctrl)

		req := &pb.BlockRequest{Hash: []byte("zoulou")}
		b := &pb.Block{Header: &pb.Header{Nonce: 42}}
		ch.EXPECT().GetBlockByHash(gomock.Any()).Return(b, nil).Times(1)
		enc.EXPECT().Encode(pb.NewBlockResponse(b)).Return(nil).Times(1)

		err := p2p.RespondBlockByHash(ctx, req, enc, ch)
		assert.NoError(t, err, "RespondBlockByHash()")
	})

	t.Run("RespondBlocksByNumber", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		ch := mockchain.NewMockReader(ctrl)
		enc := mockencoder.NewMockEncoder(ctrl)

		req := &pb.BlocksRequest{From: 3, Amount: 42}
		b := &pb.Block{Header: &pb.Header{Nonce: 42}}
		ch.EXPECT().GetBlockByNumber(uint64(3)).Return(b, nil).Times(1)
		ch.EXPECT().GetBlockByNumber(uint64(4)).Return(b, nil).Times(1)
		ch.EXPECT().GetBlockByNumber(uint64(5)).Return(nil, chain.ErrBlockNotFound).Times(1)
		enc.EXPECT().Encode(pb.NewBlocksResponse([]*pb.Block{b, b})).Return(nil).Times(1)

		err := p2p.RespondBlocksByNumber(ctx, req, enc, ch)
		assert.NoError(t, err, "RespondBlockByHash()")
	})

}

// Returns a stream handler that always returns response
func getSingleResponseStreamHandler(ctx context.Context, response *pb.Response) func(inet.Stream) {

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
