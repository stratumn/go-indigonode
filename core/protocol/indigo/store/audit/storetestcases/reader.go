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

// TestGetByPeer runs tests on the reader.GetByPeer method.
func (f Factory) TestGetByPeer(t *testing.T) {
	store, err := f.New()
	require.NoError(t, err)
	defer f.Free(store)

	sk1, _, _ := ic.GenerateEd25519Key(rand.Reader)
	peer1, _ := peer.IDFromPrivateKey(sk1)

	sk2, _, _ := ic.GenerateEd25519Key(rand.Reader)
	peer2, _ := peer.IDFromPrivateKey(sk2)

	for i := 0; i < 10; i++ {
		key := sk1
		if i%2 == 0 {
			key = sk2
		}

		l, _ := pb.NewSignedLink(key, cstesting.RandomLink())
		require.NoError(
			t,
			store.AddLink(context.Background(), l),
			"store.AddLink()",
		)
	}

	tests := []struct {
		name string
		run  func(*testing.T)
	}{{
		"peer-not-found",
		func(t *testing.T) {
			links, err := store.GetByPeer(context.Background(), "Sp0nG3b0B", audit.Pagination{})
			assert.NoError(t, err, "store.GetByPeer()")
			assert.Nil(t, links, "store.GetByPeer()")
		},
	}, {
		"small-pagination",
		func(t *testing.T) {
			linksPage1, err := store.GetByPeer(context.Background(), peer1, audit.Pagination{Top: 3})
			assert.NoError(t, err, "store.GetByPeer()")
			assert.Len(t, linksPage1, 3, "store.GetByPeer()")

			linksPage2, err := store.GetByPeer(context.Background(), peer1, audit.Pagination{Top: 3, Skip: 2})
			assert.NoError(t, err, "store.GetByPeer()")
			assert.Len(t, linksPage2, 3, "store.GetByPeer()")

			assert.Equal(t, linksPage1[2], linksPage2[0])
			assert.False(t, ContainsLink(linksPage2, linksPage1[0]))
			assert.False(t, ContainsLink(linksPage2, linksPage1[1]))
		},
	}, {
		"big-pagination",
		func(t *testing.T) {
			links, err := store.GetByPeer(context.Background(), peer2, audit.Pagination{Top: 10})
			assert.NoError(t, err, "store.GetByPeer()")
			assert.NotNil(t, links, "store.GetByPeer()")
			assert.Len(t, links, 5, "store.GetByPeer()")
		},
	}, {
		"out-of-bounds",
		func(t *testing.T) {
			links, err := store.GetByPeer(context.Background(), peer1, audit.Pagination{Skip: 10, Top: 3})
			assert.NoError(t, err, "store.GetByPeer()")
			assert.Nil(t, links, "store.GetByPeer()")
		},
	}, {
		"no-pagination",
		func(t *testing.T) {
			links, err := store.GetByPeer(context.Background(), peer1, audit.Pagination{})
			assert.NoError(t, err, "store.GetByPeer()")
			assert.Len(t, links, 5, "store.GetByPeer()")
		},
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.run(t)
		})
	}
}
