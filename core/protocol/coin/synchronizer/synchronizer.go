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

//go:generate mockgen -package mockchain -destination mocksynchronizer/mocksynchronizer.go github.com/stratumn/alice/core/protocol/coin/synchronizer Synchronizer

package synchronizer

import (
	"context"

	"github.com/stratumn/alice/core/protocol/coin/chain"
	"github.com/stratumn/alice/core/protocol/coin/p2p"

	pb "github.com/stratumn/alice/pb/coin"

	cid "gx/ipfs/QmcZfnkapfECQGcLZaf9B79NRg7cRa9EnZh4LSbkCzwNvY/go-cid"
	pstore "gx/ipfs/QmeZVQzUrXqaszo24DAoHfGzcmCptN9JyngLkGAiEfk2x7/go-libp2p-peerstore"
)

const (
	// GenesisHash is the has of the genesis block
	GenesisHash = "QmWFYJ4DHy1yS6T3fyJQTP8f66gW2jW7vMGtL4DnwUSBd9"
)

// Synchronizer is used to sync the local chain with a peer's chain.
type Synchronizer interface {
	// Synchronize syncs the local chain with the network.
	// The given hash tells where to start the sync from.
	// e.g, genesis block hash for full resync when a new node comes.
	Synchronize(context.Context, []byte, chain.Reader) (<-chan *pb.Block, <-chan error)
}

// ContentProviderFinder is an interface used to get the peers that provide a resource..
// The resource is identified by a content ID.
type ContentProviderFinder interface {
	FindProviders(ctx context.Context, c *cid.Cid) ([]pstore.PeerInfo, error)
	FindProvidersAsync(context.Context, *cid.Cid, int) <-chan pstore.PeerInfo
}

type synchronizer struct {
	p2p    p2p.P2P
	kaddht ContentProviderFinder
}

// NewSynchronizer initializes a new synchronizer.
func NewSynchronizer(p p2p.P2P, dht ContentProviderFinder) Synchronizer {
	return &synchronizer{
		p2p:    p,
		kaddht: dht,
	}
}

// Synchronize syncs the local chain with the network. Load the
// missing blocks and process them.
func (s *synchronizer) Synchronize(ctx context.Context, hash []byte, chain chain.Reader) (<-chan *pb.Block, <-chan error) {
	resCh := make(chan *pb.Block)
	errCh := make(chan error)
	return resCh, errCh
}
