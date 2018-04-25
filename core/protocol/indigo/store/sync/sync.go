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

//go:generate mockgen -package mocksync -destination mocksync/mocksync.go github.com/stratumn/alice/core/protocol/indigo/store/sync Engine

package sync

import (
	"context"

	"github.com/pkg/errors"
	"github.com/stratumn/go-indigocore/cs"
	"github.com/stratumn/go-indigocore/store"
	"github.com/stratumn/go-indigocore/types"

	logging "gx/ipfs/QmSpJByNKFX1sCsHBEp3R73FL4NF6FnQTEGyNAXHm2GS52/go-log"
	peer "gx/ipfs/QmZoWKhxUmZ2seW4BzX6fJkNR8hh9PsGModr7q171yq2SS/go-libp2p-peer"
)

var (
	// ErrInvalidLink is returned when an invalid link is provided.
	ErrInvalidLink = errors.New("invalid input link")

	// ErrNoConnectedPeers is returned when there are no connections to
	// peers available.
	ErrNoConnectedPeers = errors.New("not connected to any peer")

	// ErrLinkNotFound is returned when a link cannot be synced from our peers.
	ErrLinkNotFound = errors.New("link could not be synced from peers")

	// ErrInvalidLinkCount is returned when the sync protocol doesn't return
	// the expected number of links.
	ErrInvalidLinkCount = errors.New("invalid number of links found")
)

var log = logging.Logger("indigo.store.sync")

// Engine lets a node sync with other nodes to fetch
// missed content (links).
type Engine interface {
	// GetMissingLinks will sync all the links referenced until
	// it finds all the dependency graph.
	// The links returned will be properly ordered for inclusion in an Indigo store.
	// If Link1 references Link2, Link2 will appear before Link1 in the results list.
	GetMissingLinks(
		ctx context.Context,
		sender peer.ID,
		link *cs.Link,
		storeReader store.SegmentReader,
	) ([]*cs.Link, error)

	// Close cleans up resources used by the sync engine before deletion.
	Close(context.Context)
}

// ListMissingLinkHashes returns all the link hashes referenced by the given
// link that are currently missing from the store.
func ListMissingLinkHashes(
	ctx context.Context,
	link *cs.Link,
	reader store.SegmentReader,
) ([]string, error) {
	var referencedLinkHashes []string

	if link.Meta.PrevLinkHash != "" {
		referencedLinkHashes = append(referencedLinkHashes, link.Meta.PrevLinkHash)
	}

	for _, ref := range link.Meta.Refs {
		if ref.LinkHash != "" {
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

	nonMissingLinkHashes := make(map[string]struct{})
	for _, seg := range storedSegments {
		nonMissingLinkHashes[seg.GetLinkHashString()] = struct{}{}
	}

	var missingLinkHashes []string
	for _, lh := range referencedLinkHashes {
		_, ok := nonMissingLinkHashes[lh]
		if !ok {
			missingLinkHashes = append(missingLinkHashes, lh)
			nonMissingLinkHashes[lh] = struct{}{}
		}
	}

	return missingLinkHashes, nil
}

// OrderLinks takes a starting link and a map that should contain
// all the link's dependencies and returns a slice containing the
// links properly ordered by dependency.
func OrderLinks(
	ctx context.Context,
	start *cs.Link,
	linksMap map[string]*cs.Link,
	reader store.SegmentReader,
) ([]*cs.Link, error) {
	// We simply need to do a DFS on our DAG to achieve our goals.
	// Because the graph is acyclic this should always terminate.
	// Since links are referenced by hashes, and the references
	// are included in the hash computation, it's impossible to create a cycle.
	// Another option would be to do a BFS, reverse the results and
	// remove the duplicates.
	event := log.EventBegin(ctx, "OrderLinks", logging.Metadata{
		"links_count": len(linksMap),
	})
	defer event.Done()

	added := make(map[string]struct{})
	var links []*cs.Link

	var dfs func(*cs.Link) error
	dfs = func(current *cs.Link) error {
		currentLinkHash, err := current.HashString()
		if err != nil {
			return err
		}

		_, alreadyAdded := added[currentLinkHash]
		if alreadyAdded {
			return nil
		}

		recurse := func(linkHash string) error {
			if linkHash == "" {
				return nil
			}

			_, alreadyAdded = added[linkHash]
			if !alreadyAdded {
				linkHashBytes, _ := types.NewBytes32FromString(linkHash)
				seg, err := reader.GetSegment(ctx, linkHashBytes)
				if err != nil {
					return errors.WithStack(err)
				}

				if seg == nil {
					link, ok := linksMap[linkHash]
					if !ok {
						return ErrLinkNotFound
					}

					if err := dfs(link); err != nil {
						return err
					}
				}
			}

			return nil
		}

		if err := recurse(current.Meta.PrevLinkHash); err != nil {
			return err
		}

		for _, ref := range current.Meta.Refs {
			if err := recurse(ref.LinkHash); err != nil {
				return err
			}
		}

		links = append(links, current)
		added[currentLinkHash] = struct{}{}

		return nil
	}

	if err := dfs(start); err != nil {
		event.SetError(err)
		return nil, ErrLinkNotFound
	}

	// The DFS adds the start link at the end of the list,
	// we need to remove it.
	return links[:len(links)-1], nil
}
