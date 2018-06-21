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
	"encoding/json"
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
		err := s.AddRequest(ctx, &proposal.Request{
			Type:     proposal.AddNode,
			PeerID:   peer1,
			PeerAddr: peer1Addr,
		})
		assert.NoError(t, err)
	})

	t.Run("reject-missing-peer-id", func(t *testing.T) {
		s := proposal.NewInMemoryStore()
		err := s.AddRequest(ctx, &proposal.Request{
			Type: proposal.AddNode,
		})
		assert.EqualError(t, err, proposal.ErrInvalidPeerID.Error())
	})

	t.Run("reject-missing-peer-addr", func(t *testing.T) {
		s := proposal.NewInMemoryStore()
		err := s.AddRequest(ctx, &proposal.Request{
			Type:   proposal.AddNode,
			PeerID: peer1,
		})
		assert.EqualError(t, err, proposal.ErrMissingPeerAddr.Error())
	})

	t.Run("overwrite-old-request", func(t *testing.T) {
		s := proposal.NewInMemoryStore()

		err := s.AddRequest(ctx, &proposal.Request{
			Type:     proposal.AddNode,
			PeerID:   peer1,
			PeerAddr: peer1Addr,
		})
		require.NoError(t, err)

		err = s.AddRequest(ctx, &proposal.Request{
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

func TestStore_AddVote(t *testing.T) {
	ctx := context.Background()

	key := test.GeneratePrivateKey(t)
	peer1 := test.GeneratePeerID(t)

	req := &proposal.Request{
		Type:      proposal.RemoveNode,
		PeerID:    peer1,
		Challenge: []byte("much secure"),
	}

	t.Run("add-new-vote", func(t *testing.T) {
		s := proposal.NewInMemoryStore()
		err := s.AddRequest(ctx, req)
		require.NoError(t, err, "s.AddRequest()")

		v, err := proposal.NewVote(key, req)
		require.NoError(t, err, "proposal.NewVote()")

		err = s.AddVote(ctx, v)
		require.NoError(t, err, "s.AddVote()")
	})

	t.Run("ignores-duplicate-votes", func(t *testing.T) {
		s := proposal.NewInMemoryStore()
		err := s.AddRequest(ctx, req)
		require.NoError(t, err, "s.AddRequest()")

		v1, err := proposal.NewVote(key, req)
		require.NoError(t, err, "proposal.NewVote()")

		v2, err := proposal.NewVote(key, req)
		require.NoError(t, err, "proposal.NewVote()")

		err = s.AddVote(ctx, v1)
		require.NoError(t, err, "s.AddVote()")

		err = s.AddVote(ctx, v2)
		require.NoError(t, err, "s.AddVote()")

		votes, err := s.GetVotes(ctx, req.PeerID)
		require.NoError(t, err, "s.GetVotes()")
		assert.Len(t, votes, 1)
	})

	t.Run("missing-request", func(t *testing.T) {
		s := proposal.NewInMemoryStore()

		v, err := proposal.NewVote(key, req)
		require.NoError(t, err, "proposal.NewVote()")

		err = s.AddVote(ctx, v)
		require.EqualError(t, err, proposal.ErrMissingRequest.Error())
	})

	t.Run("invalid-vote", func(t *testing.T) {
		s := proposal.NewInMemoryStore()
		err := s.AddRequest(ctx, req)
		require.NoError(t, err, "s.AddRequest()")

		v, err := proposal.NewVote(key, req)
		require.NoError(t, err, "proposal.NewVote()")

		v.Challenge = []byte("challenge mismatch")

		err = s.AddVote(ctx, v)
		require.EqualError(t, err, proposal.ErrInvalidChallenge.Error())
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
		err := s.AddRequest(ctx, &proposal.Request{
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

	t.Run("remove-votes", func(t *testing.T) {
		s := proposal.NewInMemoryStore()
		req := &proposal.Request{
			Type:      proposal.RemoveNode,
			PeerID:    peer1,
			Challenge: []byte("much challenge"),
		}
		err := s.AddRequest(ctx, req)
		require.NoError(t, err)

		v, err := proposal.NewVote(test.GeneratePrivateKey(t), req)
		require.NoError(t, err, "proposal.NewVote()")

		err = s.AddVote(ctx, v)
		require.NoError(t, err, "s.AddVote()")

		err = s.Remove(ctx, peer1)
		require.NoError(t, err)

		votes, err := s.GetVotes(ctx, peer1)
		require.NoError(t, err)
		require.Len(t, votes, 0)
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
		err := s.AddRequest(ctx, &proposal.Request{
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
		err := s.AddRequest(ctx, &proposal.Request{
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
		err := s.AddRequest(ctx, &proposal.Request{
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

func TestStore_GetVotes(t *testing.T) {
	ctx := context.Background()

	key1 := test.GeneratePrivateKey(t)
	key2 := test.GeneratePrivateKey(t)
	peer1 := test.GeneratePeerID(t)

	req := &proposal.Request{
		Type:      proposal.RemoveNode,
		PeerID:    peer1,
		Challenge: []byte("much secure"),
	}

	v1, err := proposal.NewVote(key1, req)
	require.NoError(t, err, "proposal.NewVote()")

	v2, err := proposal.NewVote(key2, req)
	require.NoError(t, err, "proposal.NewVote()")

	t.Run("request-not-voted", func(t *testing.T) {
		s := proposal.NewInMemoryStore()
		err := s.AddRequest(ctx, req)
		require.NoError(t, err, "s.AddRequest()")

		votes, err := s.GetVotes(ctx, req.PeerID)
		require.NoError(t, err, "s.GetVotes()")
		assert.Len(t, votes, 0)
	})

	t.Run("request-expired", func(t *testing.T) {
		expiredReq := &proposal.Request{
			Type:      proposal.RemoveNode,
			PeerID:    test.GeneratePeerID(t),
			Challenge: []byte("much secure"),
		}

		s := proposal.NewInMemoryStore()

		err := s.AddRequest(ctx, expiredReq)
		require.NoError(t, err, "s.AddRequest()")

		v, err := proposal.NewVote(key1, expiredReq)
		require.NoError(t, err, "proposal.NewVote()")

		err = s.AddVote(ctx, v)
		require.NoError(t, err, "s.AddVote()")

		expiredReq.Expires = time.Now().UTC().Add(-5 * time.Minute)

		votes, err := s.GetVotes(ctx, expiredReq.PeerID)
		require.NoError(t, err, "s.GetVotes()")
		assert.Len(t, votes, 0)
	})

	t.Run("get-votes", func(t *testing.T) {
		s := proposal.NewInMemoryStore()
		err := s.AddRequest(ctx, req)
		require.NoError(t, err, "s.AddRequest()")

		err = s.AddVote(ctx, v1)
		require.NoError(t, err, "s.AddVote()")

		err = s.AddVote(ctx, v2)
		require.NoError(t, err, "s.AddVote()")

		votes, err := s.GetVotes(ctx, req.PeerID)
		require.NoError(t, err, "s.GetVotes()")
		assert.ElementsMatch(t, []*proposal.Vote{v1, v2}, votes)
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
			err := s.AddRequest(ctx, r)
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
			err := s.AddRequest(ctx, r)
			require.NoError(t, err)
		}

		rr, err := s.List(ctx)
		require.NoError(t, err)
		require.Len(t, rr, 1)
		assert.Equal(t, requests[2], rr[0])
	})
}

func TestStore_MarshalJSON(t *testing.T) {
	ctx := context.Background()

	r1 := &proposal.Request{
		Type:      proposal.AddNode,
		PeerID:    test.GeneratePeerID(t),
		PeerAddr:  test.GenerateMultiaddr(t),
		Info:      []byte("inf0"),
		Challenge: []byte("ch4ll3ng3"),
		Expires:   time.Now().UTC().Add(10 * time.Minute),
	}

	r2 := &proposal.Request{
		Type:      proposal.RemoveNode,
		PeerID:    test.GeneratePeerID(t),
		Challenge: []byte("such challenge"),
		Expires:   time.Now().UTC().Add(15 * time.Minute),
	}

	v1, _ := proposal.NewVote(test.GeneratePrivateKey(t), r2)
	v2, _ := proposal.NewVote(test.GeneratePrivateKey(t), r2)

	s := proposal.NewInMemoryStore()

	require.NoError(t, s.AddRequest(ctx, r1))
	require.NoError(t, s.AddRequest(ctx, r2))
	require.NoError(t, s.AddVote(ctx, v1))
	require.NoError(t, s.AddVote(ctx, v2))

	b, err := json.Marshal(s)
	require.NoError(t, err)

	var deserialized proposal.InMemoryStore
	err = json.Unmarshal(b, &deserialized)
	require.NoError(t, err)

	rr1, err := deserialized.Get(ctx, r1.PeerID)
	require.NoError(t, err)
	assert.Equal(t, r1, rr1)

	rr2, err := deserialized.Get(ctx, r2.PeerID)
	require.NoError(t, err)
	assert.Equal(t, r2, rr2)

	votes, err := deserialized.GetVotes(ctx, r2.PeerID)
	require.NoError(t, err)
	assert.ElementsMatch(t, []*proposal.Vote{v1, v2}, votes)
}
