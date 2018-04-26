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
	"io"

	"github.com/pkg/errors"
	"github.com/stratumn/alice/core/protocol/indigo/store/constants"
	pb "github.com/stratumn/alice/pb/indigo/store"
	"github.com/stratumn/go-indigocore/cs"
	"github.com/stratumn/go-indigocore/store"

	ihost "gx/ipfs/QmNmJZL7FQySMtE2BQuLMuZg2EB2CLEunJJUSVSc9YnnbV/go-libp2p-host"
	protobuf "gx/ipfs/QmRDePEiL4Yupq5EkcK3L3ko3iMgYaqUdLu7xc1kqs7dnV/go-multicodec/protobuf"
	logging "gx/ipfs/QmSpJByNKFX1sCsHBEp3R73FL4NF6FnQTEGyNAXHm2GS52/go-log"
	inet "gx/ipfs/QmXfkENeeBvh3zYA51MaSdGUdBjhQ99cP5WQe8zgr6wchG/go-libp2p-net"
	protocol "gx/ipfs/QmZNkThpqfVXs9GNbexPrfBbXSLNYeKrE7jwFM2oqHbyqN/go-libp2p-protocol"
	peer "gx/ipfs/QmZoWKhxUmZ2seW4BzX6fJkNR8hh9PsGModr7q171yq2SS/go-libp2p-peer"
)

var (
	// SingleNodeProtocolID is the protocol ID of the sync engine
	// that connects to the node that created and shared the new link
	// to sync all missing links.
	SingleNodeProtocolID = protocol.ID("/alice/indigo/store/sync/singlenode/v1.0.0")
)

// SingleNodeEngine synchronously syncs with the node that created the new
// link. That node is expected to have all the previous links in the graph
// because otherwise it couldn't prove the validity of the newly created link.
type SingleNodeEngine struct {
	host  ihost.Host
	store store.SegmentReader
}

// NewSingleNodeEngine creates a new SingleNodeEngine
// and registers its handlers.
func NewSingleNodeEngine(host ihost.Host, store store.SegmentReader) Engine {
	engine := &SingleNodeEngine{
		host:  host,
		store: store,
	}

	engine.host.SetStreamHandler(SingleNodeProtocolID, engine.syncHandler)

	return engine
}

// Close cleans up resources and protocol handlers.
func (s *SingleNodeEngine) Close(ctx context.Context) {
	s.host.RemoveStreamHandler(SingleNodeProtocolID)
}

// GetMissingLinks connects to the node that published the link to get all
// missing links in the subgraph ending in this new link.
func (s *SingleNodeEngine) GetMissingLinks(ctx context.Context, link *cs.Link, reader store.SegmentReader) ([]*cs.Link, error) {
	event := log.EventBegin(ctx, "GetMissingLinks")
	defer event.Done()

	toFetch, err := ListMissingLinkHashes(ctx, link, reader)
	if err != nil {
		event.SetError(err)
		return nil, err
	}

	if len(toFetch) == 0 {
		event.Append(logging.Metadata{"links_count": 0})
		return nil, nil
	}

	// We expect the sender to have been validated upstream,
	// so no need to check the error.
	sender, _ := peer.IDB58Decode(link.Meta.Data[constants.NodeIDKey].(string))
	stream, err := s.startStream(ctx, sender)
	if err != nil {
		event.SetError(err)
		return nil, err
	}

	defer func() {
		if err := stream.Close(); err != nil {
			event.Append(logging.Metadata{"stream_close_err": err.Error()})
		}
	}()

	linksMap, err := s.syncWithPeer(ctx, stream, toFetch, reader)
	if err != nil {
		event.SetError(err)
		return nil, err
	}

	links, err := OrderLinks(ctx, link, linksMap, reader)
	if err != nil {
		event.SetError(err)
		return nil, err
	}

	event.Append(logging.Metadata{"links_count": len(links)})

	return links, nil
}

// startStream starts a stream with the peer that created the link.
func (s *SingleNodeEngine) startStream(ctx context.Context, sender peer.ID) (inet.Stream, error) {
	event := log.EventBegin(ctx, "startStream", logging.Metadata{"peer": sender.Pretty()})
	defer event.Done()

	err := s.host.Connect(ctx, s.host.Peerstore().PeerInfo(sender))
	if err != nil {
		event.SetError(err)
		return nil, ErrNoConnectedPeers
	}

	stream, err := s.host.NewStream(ctx, sender, SingleNodeProtocolID)
	if err != nil {
		event.SetError(err)
		return nil, errors.Wrap(err, "could not start stream")
	}

	return stream, nil
}

// syncWithPeer syncs with the connected peer
// until all links have been fetched.
func (s *SingleNodeEngine) syncWithPeer(
	ctx context.Context,
	stream inet.Stream,
	toFetch []string,
	reader store.SegmentReader,
) (map[string]*cs.Link, error) {
	event := log.EventBegin(ctx, "syncWithPeer", logging.Metadata{
		"peer": stream.Conn().RemotePeer().Pretty(),
	})
	defer event.Done()

	enc := protobuf.Multicodec(nil).Encoder(stream)
	dec := protobuf.Multicodec(nil).Decoder(stream)

	receivedLinks := make(map[string]*cs.Link)

	for len(toFetch) > 0 {
		if err := enc.Encode(pb.FromLinkHashes(toFetch)); err != nil {
			event.SetError(err)
			return nil, err
		}

		var segments pb.Segments
		if err := dec.Decode(&segments); err != nil {
			event.SetError(err)
			return nil, err
		}

		// If we got an incorrect number of segments from our peer,
		// we can fail fast because we're sure we'll be missing
		// some links.
		if len(segments.Segments) != len(toFetch) {
			event.SetError(ErrInvalidLinkCount)
			return nil, ErrInvalidLinkCount
		}

		toFetch = nil
		toFetchMap := make(map[string]struct{})
		for _, segment := range segments.Segments {
			s, err := segment.ToSegment()
			if err != nil {
				event.SetError(err)
				return nil, err
			}

			linkHash := s.GetLinkHashString()
			_, ok := receivedLinks[linkHash]
			if ok {
				// No need to fetch a link multiple times.
				continue
			}

			event.Append(logging.Metadata{linkHash: "fetched"})
			receivedLinks[linkHash] = &s.Link

			linkDeps, err := ListMissingLinkHashes(ctx, &s.Link, reader)
			if err != nil {
				event.SetError(err)
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

				event.Append(logging.Metadata{lh: "fetching"})
				toFetchMap[lh] = struct{}{}
				toFetch = append(toFetch, lh)
			}
		}
	}

	return receivedLinks, nil
}

// syncHandler accepts sync requests from peers and sends them
// all the links they need to be up-to-date.
func (s *SingleNodeEngine) syncHandler(stream inet.Stream) {
	ctx := context.Background()
	event := log.EventBegin(ctx, "SyncRequest", logging.Metadata{
		"from": stream.Conn().RemotePeer().Pretty(),
	})
	defer event.Done()

	var err error
	enc := protobuf.Multicodec(nil).Encoder(stream)
	dec := protobuf.Multicodec(nil).Decoder(stream)

	// In success cases, it's the client's responsibility to close the stream.
	// In case of error, we'll close the stream here.
	for err == nil {
		var msg pb.LinkHashes
		err = dec.Decode(&msg)
		if err != nil {
			break
		}

		segFilter := &store.SegmentFilter{
			Pagination: store.Pagination{Limit: len(msg.LinkHashes)},
		}
		segFilter.LinkHashes, err = msg.ToLinkHashes()
		if err != nil {
			break
		}

		var segments cs.SegmentSlice
		segments, err = s.store.FindSegments(ctx, segFilter)
		if err != nil {
			break
		}

		var segMsg *pb.Segments
		segMsg, err = pb.FromSegments(segments)
		if err != nil {
			break
		}

		err = enc.Encode(segMsg)
	}

	if err != io.EOF {
		event.SetError(err)
		if err := stream.Close(); err != nil {
			event.Append(logging.Metadata{"stream_close_err": err.Error()})
		}
	}
}
