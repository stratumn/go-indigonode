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

package proposal_test

import (
	"context"
	"testing"
	"time"

	"github.com/stratumn/alice/core/protocol/bootstrap/proposal"
	"github.com/stratumn/alice/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStore_Add(t *testing.T) {
	ctx := context.Background()
	peer1 := test.GeneratePeerID(t)
	peer1Addr := test.GeneratePeerMultiaddr(t, peer1)

	t.Run("add-new-request", func(t *testing.T) {
		s := proposal.NewInMemoryStore()
		err := s.Add(ctx, &proposal.Request{
			Type:     proposal.AddNode,
			PeerID:   peer1,
			PeerAddr: peer1Addr,
		})
		assert.NoError(t, err)
	})

	t.Run("reject-missing-peer-id", func(t *testing.T) {
		s := proposal.NewInMemoryStore()
		err := s.Add(ctx, &proposal.Request{
			Type: proposal.AddNode,
		})
		assert.EqualError(t, err, proposal.ErrInvalidPeerID.Error())
	})

	t.Run("reject-missing-peer-addr", func(t *testing.T) {
		s := proposal.NewInMemoryStore()
		err := s.Add(ctx, &proposal.Request{
			Type:   proposal.AddNode,
			PeerID: peer1,
		})
		assert.EqualError(t, err, proposal.ErrMissingPeerAddr.Error())
	})

	t.Run("overwrite-old-request", func(t *testing.T) {
		s := proposal.NewInMemoryStore()

		err := s.Add(ctx, &proposal.Request{
			Type:     proposal.AddNode,
			PeerID:   peer1,
			PeerAddr: peer1Addr,
		})
		require.NoError(t, err)

		err = s.Add(ctx, &proposal.Request{
			Type:     proposal.AddNode,
			PeerID:   peer1,
			PeerAddr: peer1Addr,
			Info:     []byte("nananana"),
		})
		require.NoError(t, err)

		r, err := s.Get(ctx, peer1)
		require.NoError(t, err)
		require.NotNil(t, r)
		assert.Equal(t, []byte("nananana"), r.Info)
	})
}

func TestStore_Remove(t *testing.T) {
	ctx := context.Background()
	peer1 := test.GeneratePeerID(t)
	peer1Addr := test.GeneratePeerMultiaddr(t, peer1)

	t.Run("no-matching-request", func(t *testing.T) {
		s := proposal.NewInMemoryStore()
		err := s.Remove(ctx, peer1)
		assert.NoError(t, err)
	})

	t.Run("remove-request", func(t *testing.T) {
		s := proposal.NewInMemoryStore()
		err := s.Add(ctx, &proposal.Request{
			Type:     proposal.AddNode,
			PeerID:   peer1,
			PeerAddr: peer1Addr,
		})
		require.NoError(t, err)

		err = s.Remove(ctx, peer1)
		require.NoError(t, err)

		r, err := s.Get(ctx, peer1)
		require.NoError(t, err)
		require.Nil(t, r)
	})
}

func TestStore_Get(t *testing.T) {
	ctx := context.Background()
	peer1 := test.GeneratePeerID(t)
	peer1Addr := test.GeneratePeerMultiaddr(t, peer1)

	t.Run("no-matching-request", func(t *testing.T) {
		s := proposal.NewInMemoryStore()
		r, err := s.Get(ctx, peer1)
		assert.NoError(t, err)
		assert.Nil(t, r)
	})

	t.Run("get-add-request", func(t *testing.T) {
		s := proposal.NewInMemoryStore()
		err := s.Add(ctx, &proposal.Request{
			Type:     proposal.AddNode,
			PeerID:   peer1,
			PeerAddr: peer1Addr,
		})
		require.NoError(t, err)

		r, err := s.Get(ctx, peer1)
		require.NoError(t, err)
		require.NotNil(t, r)

		assert.Equal(t, peer1, r.PeerID)
		assert.Equal(t, peer1Addr, r.PeerAddr)
		assert.Equal(t, proposal.AddNode, r.Type)
	})

	t.Run("get-remove-request", func(t *testing.T) {
		s := proposal.NewInMemoryStore()
		err := s.Add(ctx, &proposal.Request{
			Type:   proposal.RemoveNode,
			PeerID: peer1,
		})
		require.NoError(t, err)

		r, err := s.Get(ctx, peer1)
		require.NoError(t, err)
		require.NotNil(t, r)

		assert.Equal(t, peer1, r.PeerID)
		assert.Equal(t, proposal.RemoveNode, r.Type)
	})

	t.Run("expired-request", func(t *testing.T) {
		s := proposal.NewInMemoryStore()
		err := s.Add(ctx, &proposal.Request{
			Type:    proposal.RemoveNode,
			PeerID:  peer1,
			Expires: time.Now().UTC().Add(-1 * time.Minute),
		})
		require.NoError(t, err)

		r, err := s.Get(ctx, peer1)
		assert.NoError(t, err)
		assert.Nil(t, r)
	})
}

func TestStore_List(t *testing.T) {
	ctx := context.Background()
	peer1 := test.GeneratePeerID(t)
	peer2 := test.GeneratePeerID(t)
	peer2Addr := test.GeneratePeerMultiaddr(t, peer2)
	peer3 := test.GeneratePeerID(t)

	t.Run("no-matching-requests", func(t *testing.T) {
		s := proposal.NewInMemoryStore()
		rr, err := s.List(ctx)
		require.NoError(t, err)
		require.Len(t, rr, 0)
	})

	t.Run("list-requests", func(t *testing.T) {
		s := proposal.NewInMemoryStore()

		requests := []*proposal.Request{
			&proposal.Request{
				Type:   proposal.RemoveNode,
				PeerID: peer1,
			},
			&proposal.Request{
				Type:     proposal.AddNode,
				PeerID:   peer2,
				PeerAddr: peer2Addr,
			},
			&proposal.Request{
				Type:   proposal.RemoveNode,
				PeerID: peer3,
			},
		}

		for _, r := range requests {
			err := s.Add(ctx, r)
			require.NoError(t, err)
		}

		rr, err := s.List(ctx)
		require.NoError(t, err)
		require.Len(t, rr, 3)
		assert.ElementsMatch(t, requests, rr)
	})

	t.Run("remove-expired-request", func(t *testing.T) {
		s := proposal.NewInMemoryStore()

		requests := []*proposal.Request{
			&proposal.Request{
				Type:    proposal.RemoveNode,
				PeerID:  peer1,
				Expires: time.Now().UTC().Add(-1 * time.Minute),
			},
			&proposal.Request{
				Type:     proposal.AddNode,
				PeerID:   peer2,
				PeerAddr: peer2Addr,
				Expires:  time.Now().UTC().Add(-1 * time.Minute),
			},
			&proposal.Request{
				Type:   proposal.RemoveNode,
				PeerID: peer3,
			},
		}

		for _, r := range requests {
			err := s.Add(ctx, r)
			require.NoError(t, err)
		}

		rr, err := s.List(ctx)
		require.NoError(t, err)
		require.Len(t, rr, 1)
		assert.Equal(t, requests[2], rr[0])
	})
}
