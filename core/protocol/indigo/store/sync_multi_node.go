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
	"fmt"
	"io"

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
	// MultiNodeSyncProtocolID is the protocol ID of the sync engine
	// that splits the sync load between several nodes.
	MultiNodeSyncProtocolID = protocol.ID("/alice/indigo/store/sync/multinode/v1.0.0")
)

// MultiNodeSyncEngine synchronously syncs with peers to get
// all missing links in one shot. It asks directly each of its
// connected peers for those missing links. The connected peers
// don't recursively ask their own peers if they don't have the link,
// so this sync engine will not work if you aren't connected to
// at least one peer having the link.
type MultiNodeSyncEngine struct {
	host  ihost.Host
	store store.SegmentReader
}

// NewMultiNodeSyncEngine creates a new MultiNodeSyncEngine
// and registers its handlers.
func NewMultiNodeSyncEngine(host ihost.Host, store store.SegmentReader) *MultiNodeSyncEngine {
	engine := &MultiNodeSyncEngine{
		host:  host,
		store: store,
	}

	engine.host.SetStreamHandler(MultiNodeSyncProtocolID, engine.syncHandler)

	return engine
}

// Close cleans up resources and protocol handlers.
func (s *MultiNodeSyncEngine) Close(ctx context.Context) {
	s.host.RemoveStreamHandler(MultiNodeSyncProtocolID)
}

// GetMissingLinks will recursively walk the link graph and get all the missing
// links in one pass. This might be a long operation.
func (s *MultiNodeSyncEngine) GetMissingLinks(ctx context.Context, link *cs.Link, reader store.SegmentReader) ([]*cs.Link, error) {
	event := log.EventBegin(ctx, "GetMissingLinks")
	defer event.Done()

	lh, err := link.Hash()
	if err != nil {
		event.SetError(err)
		return nil, errors.Wrap(err, "invalid link hash")
	}

	event.Append(logging.Metadata{"link_hash": lh.String()})

	connectedPeers := s.host.Network().Peers()
	if len(connectedPeers) == 0 {
		event.SetError(ErrNoConnectedPeers)
		return nil, ErrNoConnectedPeers
	}

	toFetch, fetchedLinks, err := listMissingLinkHashes(ctx, link, reader)
	if err != nil {
		return nil, err
	}

	linksCount := 0
	for len(toFetch) > 0 {
		linksCount++
		linkHashStr := toFetch[0]
		toFetch = toFetch[1:]
		linkHash, _ := types.NewBytes32FromString(linkHashStr)

		event.Append(logging.Metadata{fmt.Sprintf("fetching_%d", linksCount): linkHashStr})

		found := false

		// We currently test each peer one after the other.
		// This wastes less bandwidth than trying all of them in parallel,
		// but is probably slower.
		// We'll need real-world data to tweak this to something that is
		// fine-tuned for our usecases.
		for _, pid := range s.host.Network().Peers() {
			l, err := s.getLinkFromPeer(ctx, pid, linkHash)
			if err != nil {
				if err != io.EOF {
					event.Append(logging.Metadata{
						"peer":         pid.Pretty(),
						"stream_error": err,
					})
				}
				continue
			}

			moreToFetch, moreIncluded, err := listMissingLinkHashes(ctx, l, reader)
			if err != nil {
				return nil, err
			}

			toFetch = append(toFetch, moreToFetch...)

			// We need to prepend to make sure links are correctly dependency-ordered.
			fetchedLinks = append(append(moreIncluded, l), fetchedLinks...)
			found = true

			break
		}

		if !found {
			event.SetError(errors.Wrap(ErrLinkNotFound, linkHashStr))
			return nil, ErrLinkNotFound
		}
	}

	return fetchedLinks, nil
}

// getLinkFromPeer requests a link from a chosen peer.
func (s *MultiNodeSyncEngine) getLinkFromPeer(ctx context.Context, pid peer.ID, lh *types.Bytes32) (*cs.Link, error) {
	stream, err := s.host.NewStream(ctx, pid, MultiNodeSyncProtocolID)
	if err != nil {
		return nil, errors.Wrap(err, "couldn't start peer stream")
	}
	defer func() {
		if err := stream.Close(); err != nil {
			log.Event(ctx, "StreamCloseError", logging.Metadata{"err": err})
		}
	}()

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
// It also returns links that are fully included in references.
func listMissingLinkHashes(ctx context.Context, link *cs.Link, reader store.SegmentReader) ([]string, []*cs.Link, error) {
	var referencedLinkHashes []string
	var referencedLinks []*cs.Link

	if link.Meta.PrevLinkHash != "" {
		referencedLinkHashes = append(referencedLinkHashes, link.Meta.PrevLinkHash)
	}

	for _, ref := range link.Meta.Refs {
		if ref.Segment != nil {
			lhs, ls, err := listMissingLinkHashes(ctx, &ref.Segment.Link, reader)
			if err != nil {
				return nil, nil, err
			}

			referencedLinkHashes = append(referencedLinkHashes, lhs...)
			referencedLinks = append(referencedLinks, ls...)
			referencedLinks = append(referencedLinks, &ref.Segment.Link)
		} else if ref.LinkHash != "" {
			referencedLinkHashes = append(referencedLinkHashes, ref.LinkHash)
		}
	}

	storedSegments, err := reader.FindSegments(ctx, &store.SegmentFilter{
		LinkHashes: referencedLinkHashes,
		Pagination: store.Pagination{Limit: len(referencedLinkHashes)},
	})
	if err != nil {
		return nil, nil, errors.Wrap(err, "couldn't find segments")
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

	return missingLinkHashes, referencedLinks, nil
}

func (s *MultiNodeSyncEngine) syncHandler(stream inet.Stream) {
	ctx := context.Background()
	var err error
	event := log.EventBegin(ctx, "SyncRequest", logging.Metadata{
		"from": stream.Conn().RemotePeer().Pretty(),
	})
	defer func() {
		if err != nil {
			event.SetError(err)
		}

		if err := stream.Close(); err != nil {
			event.Append(logging.Metadata{"stream_close_err": err})
		}

		event.Done()
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
