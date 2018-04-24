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

package store

import (
	"context"

	"github.com/stratumn/go-indigocore/cs"
	"github.com/stratumn/go-indigocore/store"

	ihost "gx/ipfs/QmNmJZL7FQySMtE2BQuLMuZg2EB2CLEunJJUSVSc9YnnbV/go-libp2p-host"
	logging "gx/ipfs/QmSpJByNKFX1sCsHBEp3R73FL4NF6FnQTEGyNAXHm2GS52/go-log"
	inet "gx/ipfs/QmXfkENeeBvh3zYA51MaSdGUdBjhQ99cP5WQe8zgr6wchG/go-libp2p-net"
	protocol "gx/ipfs/QmZNkThpqfVXs9GNbexPrfBbXSLNYeKrE7jwFM2oqHbyqN/go-libp2p-protocol"
	peer "gx/ipfs/QmZoWKhxUmZ2seW4BzX6fJkNR8hh9PsGModr7q171yq2SS/go-libp2p-peer"
)

var (
	// SingleNodeSyncProtocolID is the protocol ID of the sync engine
	// that connects to the node that created and shared the new link
	// to sync all missing links.
	SingleNodeSyncProtocolID = protocol.ID("/alice/indigo/store/sync/singlenode/v1.0.0")
)

// SingleNodeSyncEngine synchronously syncs with the node that created the new
// link. That node is expected to have all the previous links in the graph
// because otherwise it couldn't prove the validity of the newly created link.
type SingleNodeSyncEngine struct {
	host  ihost.Host
	store store.SegmentReader
}

// NewSingleNodeSyncEngine creates a new SingleNodeSyncEngine
// and registers its handlers.
func NewSingleNodeSyncEngine(host ihost.Host, store store.SegmentReader) *SingleNodeSyncEngine {
	engine := &SingleNodeSyncEngine{
		host:  host,
		store: store,
	}

	engine.host.SetStreamHandler(SingleNodeSyncProtocolID, engine.syncHandler)

	return engine
}

// Close cleans up resources and protocol handlers.
func (s *SingleNodeSyncEngine) Close(ctx context.Context) {
	s.host.RemoveStreamHandler(SingleNodeSyncProtocolID)
}

// GetMissingLinks connects to the node that published the link to get all
// missing links in the subgraph ending in this new link.
func (s *SingleNodeSyncEngine) GetMissingLinks(ctx context.Context, sender peer.ID, link *cs.Link, reader store.SegmentReader) ([]*cs.Link, error) {
	event := log.EventBegin(ctx, "GetMissingLinks")
	defer event.Done()

	return nil, nil
}

func (s *SingleNodeSyncEngine) syncHandler(stream inet.Stream) {
	ctx := context.Background()
	event := log.EventBegin(ctx, "SyncRequest", logging.Metadata{
		"from": stream.Conn().RemotePeer().Pretty(),
	})
	defer event.Done()

	// TODO: receive loop until stream closed
}
