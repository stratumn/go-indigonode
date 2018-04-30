// Copyright 2017 Stratumn SAS. All rights reserved.
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

package storetestcases

import (
	"context"
	"crypto/rand"
	"testing"

	"github.com/stratumn/alice/core/protocol/indigo/store/audit"
	"github.com/stratumn/alice/core/protocol/indigo/store/constants"
	"github.com/stratumn/go-indigocore/cs/cstesting"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	ic "gx/ipfs/QmaPbCnUMBohSGo3KnxEa2bHqyJVVeEEcwtqJAYxerieBo/go-libp2p-crypto"
)

// TestAddSegment runs tests on the writer.AddSegment method.
func (f Factory) TestAddSegment(t *testing.T) {
	ctx := context.Background()

	store, err := f.New()
	require.NoError(t, err)
	defer f.Free(store)

	sk, _, _ := ic.GenerateEd25519Key(rand.Reader)
	segment1, _ := audit.SignLink(ctx, sk, cstesting.RandomLink())
	segment2, _ := audit.SignLink(ctx, sk, cstesting.RandomLink())

	assert.NoError(t, store.AddSegment(ctx, segment1), "store.AddSegment()")
	assert.NoError(t, store.AddSegment(ctx, segment2), "store.AddSegment()")

	peerID, err := constants.GetLinkNodeID(&segment1.Link)
	assert.NoError(t, err, "peer.IDB58Decode()")

	segments, err := store.GetByPeer(ctx, peerID, audit.Pagination{})
	assert.NoError(t, err, "store.GetByPeer()")
	assert.Len(t, segments, 2)
	assert.Equal(t, *segment1, segments[0])
	assert.Equal(t, *segment2, segments[1])
}
