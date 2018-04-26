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
	pb "github.com/stratumn/alice/pb/indigo/store"
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
	// MultiNodeProtocolID is the protocol ID of the sync engine
	// that splits the sync load between several nodes.
	MultiNodeProtocolID = protocol.ID("/alice/indigo/store/sync/multinode/v1.0.0")
)

// MultiNodeEngine synchronously syncs with peers to get
// all missing links in one shot. It asks directly each of its
// connected peers for those missing links. The connected peers
// don't recursively ask their own peers if they don't have the link,
// so this sync engine will not work if you aren't connected to
// at least one peer having the link.
type MultiNodeEngine struct {
	host  ihost.Host
	store store.SegmentReader
}

// NewMultiNodeEngine creates a new MultiNodeEngine
// and registers its handlers.
func NewMultiNodeEngine(host ihost.Host, store store.SegmentReader) Engine {
	engine := &MultiNodeEngine{
		host:  host,
		store: store,
	}

	engine.host.SetStreamHandler(MultiNodeProtocolID, engine.syncHandler)

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
	stream, err := s.host.NewStream(ctx, pid, MultiNodeProtocolID)
	if err != nil {
		return nil, errors.Wrap(err, "couldn't start peer stream")
	}
	defer func() {
		if err := stream.Close(); err != nil {
			log.Event(ctx, "StreamCloseError", logging.Metadata{"err": err})
		}
	}()

	enc := protobuf.Multicodec(nil).Encoder(stream)
	err = enc.Encode(pb.FromLinkHash(lh))
	if err != nil {
		return nil, err
	}

	dec := protobuf.Multicodec(nil).Decoder(stream)
	var link pb.Link
	err = dec.Decode(&link)
	if err != nil {
		return nil, err
	}

	return link.ToLink()
}

func (s *MultiNodeEngine) syncHandler(stream inet.Stream) {
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
	var linkHash pb.LinkHash
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

	link, err := pb.FromLink(&seg.Link)
	if err != nil {
		return
	}

	enc := protobuf.Multicodec(nil).Encoder(stream)
	err = enc.Encode(link)
	if err != nil {
		return
	}
}
