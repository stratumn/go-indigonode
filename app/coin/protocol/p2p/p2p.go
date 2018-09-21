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
	"encoding/hex"

	"github.com/pkg/errors"
	"github.com/stratumn/go-indigonode/app/coin/pb"
	"github.com/stratumn/go-indigonode/app/coin/protocol/chain"

	peer "gx/ipfs/QmQsErDt8Qgw1XrsXf2BpEzDgGWtB1YLsTAARBup5b6B9W/go-libp2p-peer"
	logging "gx/ipfs/QmSpJByNKFX1sCsHBEp3R73FL4NF6FnQTEGyNAXHm2GS52/go-log"
	protocol "gx/ipfs/QmZNkThpqfVXs9GNbexPrfBbXSLNYeKrE7jwFM2oqHbyqN/go-libp2p-protocol"
	ihost "gx/ipfs/QmeMYW7Nj8jnnEfs9qhm7SxKkoDPUWXu3MsxX6BFwz34tf/go-libp2p-host"
	protobuf "gx/ipfs/QmewJ1Zp9Hwz5HcMd7JYjhLXwvEHTL2UBCCz3oLt1E2N5z/go-multicodec/protobuf"
)

// log is the logger for the coin p2p layer.
var log = logging.Logger("coin.p2p")

var (
	// ErrBadResponseType is returned when a call to a peer returns an
	// unexpected response type.
	ErrBadResponseType = errors.New("bad response type")
)

// Encoder is an interface that implements an Encode method.
type Encoder interface {
	Encode(interface{}) error
}

// P2P is where the p2p APIs are defined.
type P2P interface {
	// RequestHeaderByHash request a header given its hash from a peer.
	RequestHeaderByHash(ctx context.Context, peerID peer.ID, hash []byte) (*pb.Header, error)

	// RequestBlockByHash request a block given its header's hash from a peer.
	RequestBlockByHash(ctx context.Context, peerID peer.ID, hash []byte) (*pb.Block, error)

	// RequestHeadersByNumber request a batch of headers within a range in the main branch of a peer.
	RequestHeadersByNumber(ctx context.Context, peerID peer.ID, from, amount uint64) ([]*pb.Header, error)

	// RequestHeadersByNumber request a batch of blocks within a range in the main branch of a peer.
	RequestBlocksByNumber(ctx context.Context, peerID peer.ID, from, amount uint64) ([]*pb.Block, error)

	// RespondHeaderByHash responds to a HeaderRequest.
	RespondHeaderByHash(ctx context.Context, req *pb.HeaderRequest, enc Encoder, chain chain.Reader) error

	// RespondHeadersByNumber responds to a HeadersRequest.
	RespondHeadersByNumber(ctx context.Context, req *pb.HeadersRequest, enc Encoder, c chain.Reader) error

	// RespondBlockByHash responds to a BlockRequest.
	RespondBlockByHash(ctx context.Context, req *pb.BlockRequest, enc Encoder, chain chain.Reader) error

	// RespondBlocksByNumber responds to a BlocksRequest.
	RespondBlocksByNumber(ctx context.Context, req *pb.BlocksRequest, enc Encoder, c chain.Reader) error
}

// P2P is where the p2p APIs are defined.
type p2p struct {
	host       ihost.Host
	protocolID protocol.ID
}

// NewP2P returns a new p2p handler.
func NewP2P(host ihost.Host, p protocol.ID) P2P {
	return &p2p{
		host:       host,
		protocolID: p,
	}
}

// RequestHeaderByHash requests the header for the given hash
// on the main chain from peer peerID.
func (p *p2p) RequestHeaderByHash(ctx context.Context, peerID peer.ID, hash []byte) (*pb.Header, error) {
	e := log.EventBegin(ctx, "RequestHeaderByHash", logging.Metadata{
		"peerID": peerID.Loggable(),
		"hash":   hex.EncodeToString(hash),
	})
	defer e.Done()

	req := pb.NewHeaderRequest(hash)
	rsp, err := p.request(ctx, peerID, req)
	if err != nil {
		e.SetError(err)
		return nil, errors.WithStack(err)
	}

	return rsp.GetHeader()
}

// RequestBlockByHash requests the header for the given hash
// on the main chain from peer peerID.
func (p *p2p) RequestBlockByHash(ctx context.Context, peerID peer.ID, hash []byte) (*pb.Block, error) {
	e := log.EventBegin(ctx, "RequestBlockByHash", logging.Metadata{
		"peerID": peerID.Loggable(),
		"hash":   hex.EncodeToString(hash),
	})
	defer e.Done()

	req := pb.NewBlockRequest(hash)
	rsp, err := p.request(ctx, peerID, req)
	if err != nil {
		e.SetError(err)
		return nil, errors.WithStack(err)
	}

	return rsp.GetBlock()
}

// RequestHeadersByNumber requests a batch of headers reanging from `from`
// and with length `amount`.
func (p *p2p) RequestHeadersByNumber(ctx context.Context, peerID peer.ID, from, amount uint64) ([]*pb.Header, error) {
	e := log.EventBegin(ctx, "RequestHeadersByNumber", logging.Metadata{"peerID": peerID.Loggable(), "from": from, "amount": amount})
	defer e.Done()

	req := pb.NewHeadersRequest(from, amount)
	rsp, err := p.request(ctx, peerID, req)
	if err != nil {
		e.SetError(err)
		return nil, errors.WithStack(err)
	}

	return rsp.GetHeaders()
}

// RequestBlocksByNumber requests a batch of blocks reanging from `from`
// and with length `amount`.
func (p *p2p) RequestBlocksByNumber(ctx context.Context, peerID peer.ID, from, amount uint64) ([]*pb.Block, error) {
	e := log.EventBegin(ctx, "RequestBlocksByNumber", logging.Metadata{"peerID": peerID.Loggable(), "from": from, "amount": amount})
	defer e.Done()

	req := pb.NewBlocksRequest(from, amount)
	rsp, err := p.request(ctx, peerID, req)
	if err != nil {
		e.SetError(err)
		return nil, errors.WithStack(err)
	}

	return rsp.GetBlocks()
}

// request sends a request message to a peer and return its response.
func (p *p2p) request(ctx context.Context, pid peer.ID, message *pb.Request) (*pb.Response, error) {
	successCh := make(chan *pb.Response, 1)
	errCh := make(chan error, 1)

	go func() {
		stream, err := p.host.NewStream(ctx, pid, p.protocolID)
		if err != nil {
			errCh <- errors.WithStack(err)
			return
		}

		// Send the request
		enc := protobuf.Multicodec(nil).Encoder(stream)
		if err = enc.Encode(message); err != nil {
			errCh <- errors.WithStack(err)
			return
		}

		// Get the response
		dec := protobuf.Multicodec(nil).Decoder(stream)
		rsp := pb.Response{}

		if err = dec.Decode(&rsp); err != nil {
			errCh <- errors.WithStack(err)
			return
		}

		successCh <- &rsp
	}()

	select {
	case <-ctx.Done():
		return nil, errors.WithStack(ctx.Err())
	case rsp := <-successCh:
		return rsp, nil
	case err := <-errCh:
		return nil, err
	}
}

// RespondHeaderByHash responds to a HeaderRequest.
func (p *p2p) RespondHeaderByHash(ctx context.Context, req *pb.HeaderRequest, enc Encoder, chain chain.Reader) error {
	e := log.EventBegin(ctx, "RespondHeaderByHash", logging.Metadata{"hash": hex.EncodeToString(req.Hash)})
	defer e.Done()

	h, err := chain.GetHeaderByHash(req.Hash)
	if err != nil {
		e.SetError(err)
		return err
	}
	return enc.Encode(pb.NewHeaderResponse(h))
}

// RespondHeadersByNumber responds to a HeadersRequest.
func (p *p2p) RespondHeadersByNumber(ctx context.Context, req *pb.HeadersRequest, enc Encoder, c chain.Reader) error {
	e := log.EventBegin(ctx, "RespondHeadersByNumber", logging.Metadata{"from": req.From, "amount": req.Amount})
	defer e.Done()

	headers := make([]*pb.Header, req.Amount)
	i := uint64(0)
	for ; i < req.Amount; i++ {
		h, err := c.GetHeaderByNumber(req.From + i)
		if err == chain.ErrBlockNotFound {
			break
		} else if err != nil {
			e.SetError(err)
			return err
		}
		headers[i] = h
	}

	rsp := pb.NewHeadersResponse(headers[:i])
	if err := enc.Encode(rsp); err != nil {
		e.SetError(err)
		return err
	}
	return nil
}

// RespondBlockByHash responds to a BlockRequest.
func (p *p2p) RespondBlockByHash(ctx context.Context, req *pb.BlockRequest, enc Encoder, chain chain.Reader) error {
	e := log.EventBegin(ctx, "RespondBlockByHash", logging.Metadata{"hash": hex.EncodeToString(req.Hash)})
	defer e.Done()
	b, err := chain.GetBlockByHash(req.Hash)
	if err != nil {
		e.SetError(err)
		return err
	}
	return enc.Encode(pb.NewBlockResponse(b))
}

// RespondBlocksByNumber responds to a BlocksRequest.
func (p *p2p) RespondBlocksByNumber(ctx context.Context, req *pb.BlocksRequest, enc Encoder, c chain.Reader) error {
	e := log.EventBegin(ctx, "RespondBlocksByNumber", logging.Metadata{"from": req.From, "amount": req.Amount})
	defer e.Done()

	blocks := make([]*pb.Block, req.Amount)
	i := uint64(0)
	for ; i < req.Amount; i++ {
		h, err := c.GetBlockByNumber(req.From + i)
		if err == chain.ErrBlockNotFound {
			break
		} else if err != nil {
			e.SetError(err)
			return err
		}
		blocks[i] = h
	}

	rsp := pb.NewBlocksResponse(blocks[:i])
	if err := enc.Encode(rsp); err != nil {
		e.SetError(err)
		return err
	}
	return nil
}
