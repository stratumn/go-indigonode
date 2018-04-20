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

	"github.com/stratumn/go-indigocore/cs"
	"github.com/stratumn/go-indigocore/store"
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
