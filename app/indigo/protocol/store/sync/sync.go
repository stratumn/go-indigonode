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

	"github.com/pkg/errors"
	"github.com/stratumn/go-indigocore/cs"
	"github.com/stratumn/go-indigocore/store"
	"github.com/stratumn/go-indigocore/types"
	"github.com/stratumn/go-node/core/monitoring"
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

// Engine lets a node sync with other nodes to fetch
// missed content (links).
type Engine interface {
	// GetMissingLinks will sync all the links referenced until
	// it finds all the dependency graph.
	// The links returned will be properly ordered for inclusion in an Indigo store.
	// If Link1 references Link2, Link2 will appear before Link1 in the results list.
	GetMissingLinks(
		ctx context.Context,
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
	ctx, span := monitoring.StartSpan(ctx, "indigo.store.sync", "OrderLinks")
	defer span.End()

	span.AddIntAttribute("links_count", int64(len(linksMap)))

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
		span.SetUnknownError(err)
		return nil, ErrLinkNotFound
	}

	// The DFS adds the start link at the end of the list,
	// we need to remove it.
	return links[:len(links)-1], nil
}
