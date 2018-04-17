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
	pb "github.com/stratumn/alice/pb/indigo/store"
	"github.com/stratumn/go-indigocore/cs/cstesting"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	peer "gx/ipfs/QmZoWKhxUmZ2seW4BzX6fJkNR8hh9PsGModr7q171yq2SS/go-libp2p-peer"
	ic "gx/ipfs/QmaPbCnUMBohSGo3KnxEa2bHqyJVVeEEcwtqJAYxerieBo/go-libp2p-crypto"
)

// TestAddLink runs tests on the writer.AddLink method.
func (f Factory) TestAddLink(t *testing.T) {
	store, err := f.New()
	require.NoError(t, err)
	defer f.Free(store)

	sk, _, _ := ic.GenerateEd25519Key(rand.Reader)
	l, _ := pb.NewSignedLink(sk, cstesting.RandomLink())

	assert.NoError(t, store.AddLink(context.Background(), l), "store.AddLink()")

	peerID, err := peer.IDFromBytes(l.From)
	assert.NoError(t, err, "peer.IDFromBytes()")

	links, err := store.GetByPeer(context.Background(), peerID, audit.Pagination{})
	assert.NoError(t, err, "store.GetByPeer()")
	assert.Len(t, links, 1)
	assert.Equal(t, l, links)
}
