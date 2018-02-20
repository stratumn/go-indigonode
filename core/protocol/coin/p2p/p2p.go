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

//go:generate mockgen -package mockchain -destination mockp2p/mockp2p.go github.com/stratumn/alice/core/protocol/coin/p2p P2P

package p2p

import (
	"context"

	"github.com/pkg/errors"
	pb "github.com/stratumn/alice/pb/coin"

	protobuf "gx/ipfs/QmRDePEiL4Yupq5EkcK3L3ko3iMgYaqUdLu7xc1kqs7dnV/go-multicodec/protobuf"
	logging "gx/ipfs/QmSpJByNKFX1sCsHBEp3R73FL4NF6FnQTEGyNAXHm2GS52/go-log"
	protocol "gx/ipfs/QmZNkThpqfVXs9GNbexPrfBbXSLNYeKrE7jwFM2oqHbyqN/go-libp2p-protocol"
	peer "gx/ipfs/Qma7H6RW8wRrfZpNSXwxYGcd1E149s42FpWNpDNieSVrnU/go-libp2p-peer"
	ihost "gx/ipfs/QmfCtHMCd9xFvehvHeVxtKVXJTMVTuHhyPRVHEXetn87vL/go-libp2p-host"
)

// log is the logger for the coin p2p layer.
var log = logging.Logger("coin.p2p")

var (
	// ErrBadResponseType is returned when a call to a peer returns an
	// unexpected response type.
	ErrBadResponseType = errors.New("bad response type")
)

// P2P is where the p2p APIs are defined
type P2P interface {
	// RequestHeaderByHash request a header given its hash from a peer.
	RequestHeaderByHash(ctx context.Context, peerID peer.ID, hash []byte) (*pb.Header, error)

	// RequestBlockByHash request a block given its header's hash from a peer.
	RequestBlockByHash(ctx context.Context, peerID peer.ID, hash []byte) (*pb.Block, error)

	// RequestHeadersByNumber request a batch of headers within a range in the main branch of a peer.
	RequestHeadersByNumber(ctx context.Context, peerID peer.ID, from, amount uint64) ([]*pb.Header, error)

	// RequestHeadersByNumber request a batch of blocks within a range in the main branch of a peer.
	RequestBlocksByNumber(ctx context.Context, peerID peer.ID, from, amount uint64) ([]*pb.Block, error)
}

// P2P is where the p2p APIs are defined
type p2p struct {
	// host coin.Host
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
	req := pb.NewHeaderRequest(hash)
	rsp, err := p.request(ctx, peerID, req)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return rsp.GetHeader()
}

// RequestBlockByHash requests the header for the given hash
// on the main chain from peer peerID.
func (p *p2p) RequestBlockByHash(ctx context.Context, peerID peer.ID, hash []byte) (*pb.Block, error) {
	req := pb.NewBlockRequest(hash)
	rsp, err := p.request(ctx, peerID, req)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return rsp.GetBlock()
}

// RequestHeadersByNumber requests a batch of headers reanging from `from`
// and with length `amount`.
func (p *p2p) RequestHeadersByNumber(ctx context.Context, peerID peer.ID, from, amount uint64) ([]*pb.Header, error) {
	req := pb.NewHeadersRequest(from, amount)
	rsp, err := p.request(ctx, peerID, req)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return rsp.GetHeaders()
}

// RequestBlocksByNumber requests a batch of blocks reanging from `from`
// and with length `amount`.
func (p *p2p) RequestBlocksByNumber(ctx context.Context, peerID peer.ID, from, amount uint64) ([]*pb.Block, error) {
	req := pb.NewBlocksRequest(from, amount)
	rsp, err := p.request(ctx, peerID, req)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return rsp.GetBlocks()
}

// request sends a request message to a peer. and return its response.
func (p *p2p) request(ctx context.Context, pid peer.ID, message *pb.Request) (*pb.Response, error) {
	event := log.EventBegin(ctx, "Send", logging.Metadata{
		"peerID": pid.Pretty(),
	})
	defer event.Done()

	successCh := make(chan *pb.Response, 1)
	errCh := make(chan error, 1)

	go func() {
		stream, err := p.host.NewStream(ctx, pid, p.protocolID)
		if err != nil {
			event.SetError(err)
			errCh <- errors.WithStack(err)
			return
		}

		// Send the request
		enc := protobuf.Multicodec(nil).Encoder(stream)
		if err = enc.Encode(message); err != nil {
			event.SetError(err)
			errCh <- errors.WithStack(err)
			return
		}

		// Get the response
		dec := protobuf.Multicodec(nil).Decoder(stream)
		rsp := pb.Response{}

		if err = dec.Decode(&rsp); err != nil {
			event.SetError(err)
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
