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

package sync

import (
	"context"

	"github.com/stratumn/go-indigocore/cs"
	"github.com/stratumn/go-indigocore/store"
	pb "github.com/stratumn/go-indigonode/app/indigo/pb/store"
	"github.com/stratumn/go-indigonode/app/indigo/protocol/store/constants"
	"github.com/stratumn/go-indigonode/core/monitoring"
	"github.com/stratumn/go-indigonode/core/streamutil"

	inet "gx/ipfs/QmZNJyx9GGCX4GeuHnLB8fxaxMLs4MjTjHokxfQcCd6Nve/go-libp2p-net"
	protocol "gx/ipfs/QmZNkThpqfVXs9GNbexPrfBbXSLNYeKrE7jwFM2oqHbyqN/go-libp2p-protocol"
	ihost "gx/ipfs/QmeMYW7Nj8jnnEfs9qhm7SxKkoDPUWXu3MsxX6BFwz34tf/go-libp2p-host"
)

var (
	// SingleNodeProtocolID is the protocol ID of the sync engine
	// that connects to the node that created and shared the new link
	// to sync all missing links.
	SingleNodeProtocolID = protocol.ID("/indigo/node/indigo/store/sync/singlenode/v1.0.0")
)

// SingleNodeEngine synchronously syncs with the node that created the new
// link. That node is expected to have all the previous links in the graph
// because otherwise it couldn't prove the validity of the newly created link.
type SingleNodeEngine struct {
	host           ihost.Host
	store          store.SegmentReader
	streamProvider streamutil.Provider
}

// NewSingleNodeEngine creates a new SingleNodeEngine
// and registers its handlers.
func NewSingleNodeEngine(host ihost.Host, store store.SegmentReader, provider streamutil.Provider) Engine {
	engine := &SingleNodeEngine{
		host:           host,
		store:          store,
		streamProvider: provider,
	}

	engine.host.SetStreamHandler(
		SingleNodeProtocolID,
		streamutil.WithAutoClose("indigo.store.sync", "SyncRequest", engine.handleSync),
	)

	return engine
}

// Close cleans up resources and protocol handlers.
func (s *SingleNodeEngine) Close(ctx context.Context) {
	_, span := monitoring.StartSpan(ctx, "indigo.store.sync", "Close")
	defer span.End()

	s.host.RemoveStreamHandler(SingleNodeProtocolID)
}

// GetMissingLinks connects to the node that published the link to get all
// missing links in the subgraph ending in this new link.
func (s *SingleNodeEngine) GetMissingLinks(ctx context.Context, link *cs.Link, reader store.SegmentReader) ([]*cs.Link, error) {
	ctx, span := monitoring.StartSpan(ctx, "indigo.store.sync", "GetMissingLinks")
	defer span.End()

	linkHash, _ := link.HashString()
	span.AddStringAttribute("link_hash", linkHash)

	toFetch, err := ListMissingLinkHashes(ctx, link, reader)
	if err != nil {
		span.SetUnknownError(err)
		return nil, err
	}

	span.AddIntAttribute("links_to_fetch", int64(len(toFetch)))

	if len(toFetch) == 0 {
		return nil, nil
	}

	// We expect the sender to have been validated upstream,
	// so no need to check the error.
	sender, _ := constants.GetLinkNodeID(link)

	err = s.host.Connect(ctx, s.host.Peerstore().PeerInfo(sender))
	if err != nil {
		span.SetUnknownError(err)
		return nil, ErrNoConnectedPeers
	}

	stream, err := s.streamProvider.NewStream(ctx,
		s.host,
		streamutil.OptPeerID(sender),
		streamutil.OptProtocolIDs(SingleNodeProtocolID),
	)
	if err != nil {
		return nil, err
	}

	defer stream.Close()

	linksMap, err := s.syncWithPeer(ctx, stream, toFetch, reader)
	if err != nil {
		span.SetUnknownError(err)
		return nil, err
	}

	links, err := OrderLinks(ctx, link, linksMap, reader)
	if err != nil {
		span.SetUnknownError(err)
		return nil, err
	}

	span.AddIntAttribute("links_fetched", int64(len(links)))

	return links, nil
}

// syncWithPeer syncs with the connected peer
// until all links have been fetched.
func (s *SingleNodeEngine) syncWithPeer(
	ctx context.Context,
	stream streamutil.Stream,
	toFetch []string,
	reader store.SegmentReader,
) (map[string]*cs.Link, error) {
	ctx, span := monitoring.StartSpan(ctx, "indigo.store.sync", "syncWithPeer",
		monitoring.SpanOptionPeerID(stream.Conn().RemotePeer()))
	defer span.End()

	receivedLinks := make(map[string]*cs.Link)

	for len(toFetch) > 0 {
		if err := stream.Codec().Encode(pb.FromLinkHashes(toFetch)); err != nil {
			span.SetUnknownError(err)
			return nil, err
		}

		var segments pb.Segments
		if err := stream.Codec().Decode(&segments); err != nil {
			span.SetUnknownError(err)
			return nil, err
		}

		// If we got an incorrect number of segments from our peer,
		// we can fail fast because we're sure we'll be missing
		// some links.
		if len(segments.Segments) != len(toFetch) {
			span.SetStatus(monitoring.NewStatus(monitoring.StatusCodeFailedPrecondition, ErrInvalidLinkCount.Error()))
			return nil, ErrInvalidLinkCount
		}

		toFetch = nil
		toFetchMap := make(map[string]struct{})
		for _, segment := range segments.Segments {
			s, err := segment.ToSegment()
			if err != nil {
				span.SetUnknownError(err)
				return nil, err
			}

			linkHash := s.GetLinkHashString()
			_, ok := receivedLinks[linkHash]
			if ok {
				// No need to fetch a link multiple times.
				continue
			}

			span.AddBoolAttribute(linkHash, true)
			receivedLinks[linkHash] = &s.Link

			linkDeps, err := ListMissingLinkHashes(ctx, &s.Link, reader)
			if err != nil {
				span.SetUnknownError(err)
				return nil, err
			}

			for _, lh := range linkDeps {
				_, ok := receivedLinks[lh]
				if ok {
					continue
				}

				_, ok = toFetchMap[lh]
				if ok {
					continue
				}

				span.AddStringAttribute(lh, "fetching")
				toFetchMap[lh] = struct{}{}
				toFetch = append(toFetch, lh)
			}
		}
	}

	return receivedLinks, nil
}

// syncHandler accepts sync requests from peers and sends them
// all the links they need to be up-to-date.
func (s *SingleNodeEngine) handleSync(
	ctx context.Context,
	span *monitoring.Span,
	stream inet.Stream,
	codec streamutil.Codec,
) (err error) {
	for {
		var msg pb.LinkHashes
		err = codec.Decode(&msg)
		if err != nil {
			return
		}

		segFilter := &store.SegmentFilter{
			Pagination: store.Pagination{Limit: len(msg.LinkHashes)},
		}
		segFilter.LinkHashes, err = msg.ToLinkHashes()
		if err != nil {
			return
		}

		var segments cs.SegmentSlice
		segments, err = s.store.FindSegments(ctx, segFilter)
		if err != nil {
			return
		}

		var segMsg *pb.Segments
		segMsg, err = pb.FromSegments(segments)
		if err != nil {
			return
		}

		err = codec.Encode(segMsg)
		if err != nil {
			return
		}
	}
}
