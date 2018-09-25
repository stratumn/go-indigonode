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
	"io"

	"github.com/pkg/errors"
	"github.com/stratumn/go-indigocore/cs"
	"github.com/stratumn/go-indigocore/store"
	"github.com/stratumn/go-indigocore/types"
	pb "github.com/stratumn/go-indigonode/app/indigo/pb/store"
	"github.com/stratumn/go-indigonode/core/monitoring"
	"github.com/stratumn/go-indigonode/core/streamutil"

	peer "gx/ipfs/QmQsErDt8Qgw1XrsXf2BpEzDgGWtB1YLsTAARBup5b6B9W/go-libp2p-peer"
	inet "gx/ipfs/QmZNJyx9GGCX4GeuHnLB8fxaxMLs4MjTjHokxfQcCd6Nve/go-libp2p-net"
	protocol "gx/ipfs/QmZNkThpqfVXs9GNbexPrfBbXSLNYeKrE7jwFM2oqHbyqN/go-libp2p-protocol"
	ihost "gx/ipfs/QmeMYW7Nj8jnnEfs9qhm7SxKkoDPUWXu3MsxX6BFwz34tf/go-libp2p-host"
)

var (
	// MultiNodeProtocolID is the protocol ID of the sync engine
	// that splits the sync load between several nodes.
	MultiNodeProtocolID = protocol.ID("/indigo/node/indigo/store/sync/multinode/v1.0.0")
)

// MultiNodeEngine synchronously syncs with peers to get
// all missing links in one shot. It asks directly each of its
// connected peers for those missing links. The connected peers
// don't recursively ask their own peers if they don't have the link,
// so this sync engine will not work if you aren't connected to
// at least one peer having the link.
type MultiNodeEngine struct {
	host           ihost.Host
	store          store.SegmentReader
	streamProvider streamutil.Provider
}

// NewMultiNodeEngine creates a new MultiNodeEngine
// and registers its handlers.
func NewMultiNodeEngine(host ihost.Host, store store.SegmentReader, provider streamutil.Provider) Engine {
	engine := &MultiNodeEngine{
		host:           host,
		store:          store,
		streamProvider: provider,
	}

	engine.host.SetStreamHandler(
		MultiNodeProtocolID,
		streamutil.WithAutoClose("indigo.store.sync", "SyncRequest", engine.handleSync),
	)

	return engine
}

// Close cleans up resources and protocol handlers.
func (s *MultiNodeEngine) Close(ctx context.Context) {
	s.host.RemoveStreamHandler(MultiNodeProtocolID)
}

// GetMissingLinks will recursively walk the link graph and get all the missing
// links in one pass. This might be a long operation.
func (s *MultiNodeEngine) GetMissingLinks(ctx context.Context, link *cs.Link, reader store.SegmentReader) ([]*cs.Link, error) {
	ctx, span := monitoring.StartSpan(ctx, "indigo.store.sync", "GetMissingLinks")
	defer span.End()

	lh, err := link.Hash()
	if err != nil {
		span.SetStatus(monitoring.NewStatus(monitoring.StatusCodeInvalidArgument, ErrInvalidLink.Error()))
		return nil, ErrInvalidLink
	}

	span.AddStringAttribute("link_hash", lh.String())

	connectedPeers := s.host.Network().Peers()
	if len(connectedPeers) == 0 {
		span.SetUnknownError(ErrNoConnectedPeers)
		return nil, ErrNoConnectedPeers
	}

	toFetch, err := ListMissingLinkHashes(ctx, link, reader)
	if err != nil {
		return nil, err
	}

	linksCount := 0
	fetchedLinks := make(map[string]*cs.Link)

	for len(toFetch) > 0 {
		linkHashStr := toFetch[0]
		toFetch = toFetch[1:]

		_, alreadyFetched := fetchedLinks[linkHashStr]
		if alreadyFetched {
			continue
		}

		linkHash, _ := types.NewBytes32FromString(linkHashStr)

		linksCount++
		span.AddStringAttribute(linkHashStr, "fetching")

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
					span.Annotate(ctx, pid.Pretty(), err.Error())
				}
				continue
			}

			found = true
			fetchedLinks[linkHashStr] = l

			moreToFetch, err := ListMissingLinkHashes(ctx, l, reader)
			if err != nil {
				return nil, err
			}

			toFetch = append(toFetch, moreToFetch...)

			break
		}

		if !found {
			span.SetStatus(monitoring.NewStatus(monitoring.StatusCodeNotFound, ErrLinkNotFound.Error()))
			return nil, ErrLinkNotFound
		}
	}

	links, err := OrderLinks(ctx, link, fetchedLinks, reader)
	if err != nil {
		span.SetUnknownError(err)
		return nil, ErrLinkNotFound
	}

	return links, nil
}

// getLinkFromPeer requests a link from a chosen peer.
func (s *MultiNodeEngine) getLinkFromPeer(ctx context.Context, pid peer.ID, lh *types.Bytes32) (*cs.Link, error) {
	stream, err := s.streamProvider.NewStream(ctx,
		s.host,
		streamutil.OptPeerID(pid),
		streamutil.OptProtocolIDs(MultiNodeProtocolID))
	if err != nil {
		return nil, errors.Wrap(err, "couldn't start peer stream")
	}

	defer stream.Close()

	err = stream.Codec().Encode(pb.FromLinkHash(lh))
	if err != nil {
		return nil, err
	}

	var link pb.Link
	err = stream.Codec().Decode(&link)
	if err != nil {
		return nil, err
	}

	return link.ToLink()
}

func (s *MultiNodeEngine) handleSync(
	ctx context.Context,
	span *monitoring.Span,
	stream inet.Stream,
	codec streamutil.Codec,
) (err error) {
	var linkHash pb.LinkHash
	err = codec.Decode(&linkHash)
	if err != nil {
		return
	}

	lh, err := linkHash.ToLinkHash()
	if err != nil {
		return
	}

	span.AddStringAttribute("link_hash", lh.String())

	seg, err := s.store.GetSegment(ctx, lh)
	if err != nil {
		return err
	}
	if seg == nil {
		return
	}

	link, err := pb.FromLink(&seg.Link)
	if err != nil {
		return
	}

	return codec.Encode(link)
}
