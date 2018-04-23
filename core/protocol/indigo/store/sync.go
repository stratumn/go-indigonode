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

//go:generate mockgen -package mocksync -destination mocksync/mocksync.go github.com/stratumn/alice/core/protocol/indigo/store SyncEngine

package store

import (
	"context"
	"fmt"
	"strings"

	"github.com/pkg/errors"
	rpcpb "github.com/stratumn/alice/grpc/indigo/store"
	"github.com/stratumn/go-indigocore/cs"
	"github.com/stratumn/go-indigocore/store"
	"github.com/stratumn/go-indigocore/types"

	ihost "gx/ipfs/QmNmJZL7FQySMtE2BQuLMuZg2EB2CLEunJJUSVSc9YnnbV/go-libp2p-host"
	protobuf "gx/ipfs/QmRDePEiL4Yupq5EkcK3L3ko3iMgYaqUdLu7xc1kqs7dnV/go-multicodec/protobuf"
	logging "gx/ipfs/QmSpJByNKFX1sCsHBEp3R73FL4NF6FnQTEGyNAXHm2GS52/go-log"
	inet "gx/ipfs/QmXfkENeeBvh3zYA51MaSdGUdBjhQ99cP5WQe8zgr6wchG/go-libp2p-net"
	protocol "gx/ipfs/QmZNkThpqfVXs9GNbexPrfBbXSLNYeKrE7jwFM2oqHbyqN/go-libp2p-protocol"
	peer "gx/ipfs/QmZoWKhxUmZ2seW4BzX6fJkNR8hh9PsGModr7q171yq2SS/go-libp2p-peer"
)

var (
	// ErrNoConnectedPeers is returned when there are no connections to
	// peers available.
	ErrNoConnectedPeers = errors.New("not connected to any peer")

	// ErrLinkNotFound is returned when a link cannot be synced from our peers.
	ErrLinkNotFound = errors.New("link could not be synced from peers")
)

var (
	// SynchronousSyncProtocolID is the protocol ID of synchronous sync engine.
	SynchronousSyncProtocolID = protocol.ID("/alice/indigo/store/sync/synchronous/v1.0.0")
)

// SyncEngine lets a node sync with other nodes to fetch
// missed content (links).
type SyncEngine interface {
	// GetMissingLinks will sync all the links referenced until
	// it finds all the dependency graph.
	// The links returned will be properly ordered for inclusion in an Indigo store.
	// If Link1 references Link2, Link2 will appear before Link1 in the results list.
	GetMissingLinks(context.Context, *cs.Link, store.SegmentReader) ([]*cs.Link, error)
}

// SynchronousSyncEngine synchronously syncs with peers to get
// all missing links in one shot.
type SynchronousSyncEngine struct {
	host  ihost.Host
	store store.SegmentReader
}

// NewSynchronousSyncEngine creates a new SynchronousSyncEngine
// and registers its handlers.
func NewSynchronousSyncEngine(host ihost.Host, store store.SegmentReader) *SynchronousSyncEngine {
	engine := &SynchronousSyncEngine{
		host:  host,
		store: store,
	}

	engine.host.SetStreamHandler(SynchronousSyncProtocolID, engine.syncHandler)

	return engine
}

// Close cleans up resources and protocol handlers.
func (s *SynchronousSyncEngine) Close(ctx context.Context) {
	s.host.RemoveStreamHandler(SynchronousSyncProtocolID)
}

// GetMissingLinks will recursively walk the link graph and get all the missing
// links in one pass. This might be a long operation.
func (s *SynchronousSyncEngine) GetMissingLinks(ctx context.Context, link *cs.Link, reader store.SegmentReader) ([]*cs.Link, error) {
	event := log.EventBegin(ctx, "GetMissingLinks")
	defer event.Done()

	lh, err := link.Hash()
	if err != nil {
		return nil, errors.Wrap(err, "invalid link hash")
	}

	event.Append(logging.Metadata{"link_hash": lh.String()})

	toFetch, err := listMissingLinkHashes(ctx, link, reader)
	if err != nil {
		return nil, err
	}

	event.Append(logging.Metadata{"fetching": strings.Join(toFetch, ",")})

	connectedPeers := s.host.Network().Peers()
	if len(connectedPeers) == 0 {
		return nil, ErrNoConnectedPeers
	}

	var fetchedLinks []*cs.Link

	for _, linkHashStr := range toFetch {
		linkHash, _ := types.NewBytes32FromString(linkHashStr)
		found := false

		// We currently test each peer one after the other.
		// This wastes less bandwidth than trying all of them in parallel,
		// but is probably slower.
		// We'll need real-world data to tweak this to something that is
		// fine-tuned for our usecases.
		for _, pid := range s.host.Network().Peers() {
			l, err := s.getLinkFromPeer(ctx, pid, linkHash)
			if err != nil {
				event.Append(logging.Metadata{
					fmt.Sprintf("stream_error_%s", pid.Pretty()): err,
				})
				continue
			}

			fetchedLinks = append(fetchedLinks, l)
			found = true
			break
		}

		if !found {
			return nil, ErrLinkNotFound
		}
	}

	return fetchedLinks, nil
}

// getLinkFromPeer requests a link from a chosen peer.
func (s *SynchronousSyncEngine) getLinkFromPeer(ctx context.Context, pid peer.ID, lh *types.Bytes32) (*cs.Link, error) {
	stream, err := s.host.NewStream(ctx, pid, SynchronousSyncProtocolID)
	if err != nil {
		return nil, errors.Wrap(err, "couldn't start peer stream")
	}
	defer stream.Close()

	enc := protobuf.Multicodec(nil).Encoder(stream)
	err = enc.Encode(rpcpb.FromLinkHash(lh))
	if err != nil {
		return nil, err
	}

	dec := protobuf.Multicodec(nil).Decoder(stream)
	var link rpcpb.Link
	err = dec.Decode(&link)
	if err != nil {
		return nil, err
	}

	return link.ToLink()
}

// listMissingLinkHashes returns all the link hashes referenced by the given link
// that can't be found in the local store.
func listMissingLinkHashes(ctx context.Context, link *cs.Link, reader store.SegmentReader) ([]string, error) {
	var referencedLinkHashes []string
	referencedLinkHashes = append(referencedLinkHashes, link.Meta.PrevLinkHash)

	for _, ref := range link.Meta.Refs {
		// TODO: deal with refs including full link (blindly add to the results list?)
		if ref.Segment == nil && ref.LinkHash != "" {
			referencedLinkHashes = append(referencedLinkHashes, ref.LinkHash)
		}
	}

	storedSegments, err := reader.FindSegments(ctx, &store.SegmentFilter{
		LinkHashes: referencedLinkHashes,
		Pagination: store.Pagination{Limit: len(referencedLinkHashes)},
	})
	if err != nil {
		return nil, errors.Wrap(err, "couldn't find segments")
	}

	storedLinkHashes := make(map[string]struct{})
	for _, seg := range storedSegments {
		storedLinkHashes[seg.GetLinkHashString()] = struct{}{}
	}

	var missingLinkHashes []string
	for _, lh := range referencedLinkHashes {
		_, ok := storedLinkHashes[lh]
		if !ok {
			missingLinkHashes = append(missingLinkHashes, lh)
		}
	}

	return missingLinkHashes, nil
}

func (s *SynchronousSyncEngine) syncHandler(stream inet.Stream) {
	ctx := context.Background()
	var err error
	event := log.EventBegin(ctx, "SyncRequest", logging.Metadata{
		"from": stream.Conn().RemotePeer().Pretty(),
	})
	defer func() {
		if err != nil {
			event.SetError(err)
		}

		event.Done()
		stream.Close()
	}()

	dec := protobuf.Multicodec(nil).Decoder(stream)
	var linkHash rpcpb.LinkHash
	err = dec.Decode(&linkHash)
	if err != nil {
		return
	}

	lh, err := linkHash.ToLinkHash()
	if err != nil {
		return
	}

	event.Append(logging.Metadata{"link_hash": lh.String()})

	seg, err := s.store.GetSegment(ctx, lh)
	if err != nil {
		return
	}
	if seg == nil {
		return
	}

	link, err := rpcpb.FromLink(&seg.Link)
	if err != nil {
		return
	}

	enc := protobuf.Multicodec(nil).Encoder(stream)
	err = enc.Encode(link)
	if err != nil {
		return
	}
}
