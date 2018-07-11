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

package sync

import (
	"context"
	"fmt"
	"io"

	"github.com/pkg/errors"
	"github.com/stratumn/go-indigocore/cs"
	"github.com/stratumn/go-indigocore/store"
	"github.com/stratumn/go-indigocore/types"
	pb "github.com/stratumn/go-indigonode/app/indigo/pb/store"
	"github.com/stratumn/go-indigonode/core/monitoring"
	"github.com/stratumn/go-indigonode/core/streamutil"

	logging "gx/ipfs/QmSpJByNKFX1sCsHBEp3R73FL4NF6FnQTEGyNAXHm2GS52/go-log"
	inet "gx/ipfs/QmXoz9o2PT3tEzf7hicegwex5UgVP54n3k82K7jrWFyN86/go-libp2p-net"
	protocol "gx/ipfs/QmZNkThpqfVXs9GNbexPrfBbXSLNYeKrE7jwFM2oqHbyqN/go-libp2p-protocol"
	peer "gx/ipfs/QmcJukH2sAFjY3HdBKq35WDzWoL3UUu2gt9wdfqZTUyM74/go-libp2p-peer"
	ihost "gx/ipfs/QmfZTdmunzKzAGJrSvXXQbQ5kLLUiEMX5vdwux7iXkdk7D/go-libp2p-host"
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
	event := log.EventBegin(ctx, "GetMissingLinks")
	defer event.Done()

	lh, err := link.Hash()
	if err != nil {
		event.SetError(err)
		return nil, ErrInvalidLink
	}

	event.Append(logging.Metadata{"link_hash": lh.String()})

	connectedPeers := s.host.Network().Peers()
	if len(connectedPeers) == 0 {
		event.SetError(ErrNoConnectedPeers)
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
			event.SetError(errors.Wrap(ErrLinkNotFound, linkHashStr))
			return nil, ErrLinkNotFound
		}
	}

	links, err := OrderLinks(ctx, link, fetchedLinks, reader)
	if err != nil {
		event.SetError(err)
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
