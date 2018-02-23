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

//go:generate mockgen -package mocksynchronizer -destination mocksynchronizer/mocksynchronizer.go github.com/stratumn/alice/core/protocol/coin/synchronizer Synchronizer
//go:generate mockgen -package mocksynchronizer -destination mocksynchronizer/mockcontentproviderfinder.go github.com/stratumn/alice/core/protocol/coin/synchronizer ContentProviderFinder

package synchronizer

import (
	"bytes"
	"context"
	"fmt"
	"time"

	"github.com/pkg/errors"
	"github.com/stratumn/alice/core/protocol/coin/chain"
	"github.com/stratumn/alice/core/protocol/coin/p2p"

	pb "github.com/stratumn/alice/pb/coin"

	peer "gx/ipfs/Qma7H6RW8wRrfZpNSXwxYGcd1E149s42FpWNpDNieSVrnU/go-libp2p-peer"
	cid "gx/ipfs/QmcZfnkapfECQGcLZaf9B79NRg7cRa9EnZh4LSbkCzwNvY/go-cid"
	pstore "gx/ipfs/QmeZVQzUrXqaszo24DAoHfGzcmCptN9JyngLkGAiEfk2x7/go-libp2p-peerstore"
)

const (
	// MaxHeadersPerBatch is the number of headers to be retrieved per batch.
	MaxHeadersPerBatch = 64

	// MaxBlocksPerBatch is the number of blocks to be retrieved per batch.
	MaxBlocksPerBatch = 42

	// ProviderSearchTimeout is the max duration to find a peer for a resource.
	ProviderSearchTimeout = time.Second * 10
)

var (
	// ErrNoProvider is returned when no peer was found for a given resource.
	ErrNoProvider = errors.New("no peer could provide the resource")

	// ErrNoCommonAncestor is returned when no common ancestor is found with a peer's chain.
	ErrNoCommonAncestor = errors.New("the peer's chain does not intersect with ours")
)

// Synchronizer is used to sync the local chain with a peer's chain.
type Synchronizer interface {
	// Synchronize syncs the local chain with the network.
	// The given hash tells where to start the sync from.
	// e.g, genesis block hash for full resync when a new node comes.
	// Returns the blocks in a channel.
	Synchronize(context.Context, []byte, chain.Reader) (<-chan *pb.Block, <-chan error)
}

// ContentProviderFinder is an interface used to get the peers that provide a resource.
// The resource is identified by a content ID.
type ContentProviderFinder interface {
	FindProviders(ctx context.Context, c *cid.Cid) ([]pstore.PeerInfo, error)
}

type synchronizer struct {
	p2p    p2p.P2P
	kaddht ContentProviderFinder

	maxHeadersPerBatch uint64
	maxBlocksPerBatch  uint64
}

// Opt is an option for Synchronizer.
type Opt func(*synchronizer)

// OptMaxBatchSizes sets a max for the number of items in batch calls.
var OptMaxBatchSizes = func(m uint64) Opt {
	return func(c *synchronizer) {
		c.maxHeadersPerBatch = m
		c.maxBlocksPerBatch = m
	}
}

// NewSynchronizer initializes a new synchronizer.
func NewSynchronizer(p p2p.P2P, dht ContentProviderFinder, opts ...Opt) Synchronizer {
	s := &synchronizer{
		p2p:                p,
		kaddht:             dht,
		maxHeadersPerBatch: MaxHeadersPerBatch,
		maxBlocksPerBatch:  MaxBlocksPerBatch,
	}

	for _, o := range opts {
		o(s)
	}

	return s
}

// getPeerForBlock find a peer that has a given block hash and returns the block and the peer.
func (s *synchronizer) getPeerForBlock(ctx context.Context, hash []byte) (peer.ID, error) {
	cid, err := cid.Cast(hash)
	if err != nil {
		return "", err
	}

	// add a timeout to the call to dht.
	dhtCtx, cancel := context.WithTimeout(ctx, ProviderSearchTimeout)
	defer cancel()

	// TODO: find a way to invalidate the DHT cache.
	peers, err := s.kaddht.FindProviders(dhtCtx, cid)
	if err != nil {
		return "", err
	}

	for _, peer := range peers {
		block, err := s.p2p.RequestBlockByHash(ctx, peer.ID, hash)
		if err == nil && block != nil {
			return peer.ID, nil
		}
	}
	return "", ErrNoProvider
}

// Synchronize syncs the local chain with the network. Load the
// missing blocks and process them.
func (s *synchronizer) Synchronize(ctx context.Context, hash []byte, chainReader chain.Reader) (<-chan *pb.Block, <-chan error) {
	resCh := make(chan *pb.Block)
	errCh := make(chan error)

	cur, err := chainReader.CurrentHeader()
	if err != nil && err != chain.ErrBlockNotFound {
		go func() {
			errCh <- err
			close(errCh)
		}()
		return resCh, errCh
	}

	// Find a peer that really has the block by getting it.
	pid, err := s.getPeerForBlock(ctx, hash)
	if err != nil {
		go func() {
			errCh <- err
			close(errCh)
		}()
		return resCh, errCh
	}

	num := uint64(0)
	if cur != nil {
		head, err := s.findCommonAncestor(ctx, cur.BlockNumber, pid, chainReader)

		if err != nil {
			go func() {
				errCh <- err
				close(errCh)
			}()
			return resCh, errCh
		}
		num = head.BlockNumber + 1
	}

	go s.fetchBlocks(ctx, pid, num, resCh, errCh)
	return resCh, errCh
}

// fetchBlocks gets blocks from the main branch of a given peer per batches.
// All received blocks (and errors) are written to channels.
func (s *synchronizer) fetchBlocks(ctx context.Context, pid peer.ID, from uint64, resCh chan<- *pb.Block, errCh chan<- error) {
	fmt.Println("Get blocks ", from)
	for {
		select {
		case <-ctx.Done():
			err := errors.WithStack(ctx.Err())
			if err != nil {
				errCh <- err
			}
			return
		default:
			blocks, err := s.p2p.RequestBlocksByNumber(ctx, pid, from, s.maxBlocksPerBatch)
			if err != nil {
				errCh <- err
				return
			}

			// TODO: Should we order the blocks first or assume they are ordered ?
			for _, block := range blocks {
				resCh <- block
			}

			if uint64(len(blocks)) < s.maxBlocksPerBatch {
				close(resCh)
				return
			}

			from = blocks[s.maxBlocksPerBatch-1].BlockNumber() + 1
		}
	}
}

// findCommonAncestor finds and returns the header of the first common ancestor
// between a node's main chain and a given local header.
func (s *synchronizer) findCommonAncestor(ctx context.Context, height uint64, peerID peer.ID, chain chain.Reader) (*pb.Header, error) {
	var ancestorHash []byte
	// The common ancestor is necessarily below height.
	// Get the 64 blocks before height.
	for {
		var cursor *pb.Header
		from := uint64(0)
		if height >= s.maxHeadersPerBatch {
			from = height - s.maxHeadersPerBatch + 1
		}

		headers, err := s.p2p.RequestHeadersByNumber(ctx, peerID, from, s.maxHeadersPerBatch)
		if err != nil {
			return nil, err
		}

		for i := len(headers) - 1; i >= 0; i-- {
			h := headers[i]
			if cursor != nil && h.BlockNumber < cursor.BlockNumber {
				continue
			}

			localHeader, err := chain.GetHeaderByNumber(h.BlockNumber)
			if err != nil {
				return nil, err
			}
			if bytes.Equal(h.PreviousHash, localHeader.PreviousHash) {
				cursor = h
				ancestorHash = h.PreviousHash
			}
		}

		if ancestorHash != nil || height < s.maxHeadersPerBatch {
			break
		}

		height -= s.maxHeadersPerBatch
	}

	if ancestorHash == nil {
		return nil, ErrNoCommonAncestor
	}

	return chain.GetHeaderByHash(ancestorHash)
}
