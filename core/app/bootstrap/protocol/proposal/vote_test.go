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

package proposal_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stratumn/go-node/core/app/bootstrap/pb"
	"github.com/stratumn/go-node/core/app/bootstrap/protocol/proposal"
	"github.com/stratumn/go-node/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	ic "github.com/libp2p/go-libp2p-crypto"
)

func TestVote_New(t *testing.T) {
	testCases := []struct {
		name    string
		key     func(*testing.T) ic.PrivKey
		request func(*testing.T) *proposal.Request
		err     error
	}{{
		"missing-request-challenge",
		test.GeneratePrivateKey,
		func(t *testing.T) *proposal.Request {
			return &proposal.Request{
				Type:   proposal.RemoveNode,
				PeerID: test.GeneratePeerID(t),
			}
		},
		proposal.ErrMissingChallenge,
	}, {
		"missing-private-key",
		func(*testing.T) ic.PrivKey { return nil },
		func(t *testing.T) *proposal.Request {
			return &proposal.Request{
				Type:      proposal.RemoveNode,
				PeerID:    test.GeneratePeerID(t),
				Challenge: []byte("such challenge very crypto"),
			}
		},
		proposal.ErrMissingPrivateKey,
	}, {
		"invalid-type",
		test.GeneratePrivateKey,
		func(t *testing.T) *proposal.Request {
			return &proposal.Request{
				Type:      42,
				PeerID:    test.GeneratePeerID(t),
				Challenge: []byte("such challenge very crypto"),
			}
		},
		proposal.ErrInvalidRequestType,
	}, {
		"valid-vote",
		test.GeneratePrivateKey,
		func(t *testing.T) *proposal.Request {
			return &proposal.Request{
				Type:      proposal.RemoveNode,
				PeerID:    test.GeneratePeerID(t),
				Challenge: []byte("such challenge very crypto"),
			}
		},
		nil,
	}}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			r := tt.request(t)
			v, err := proposal.NewVote(context.Background(), tt.key(t), r)
			if tt.err != nil {
				assert.EqualError(t, err, tt.err.Error())
			} else {
				require.NoError(t, err)
				require.NotNil(t, v)
				assert.Equal(t, r.Type, v.Type)
				assert.Equal(t, r.PeerID, v.PeerID)
				assert.Equal(t, r.Challenge, v.Challenge)
				assert.NoError(t, v.Verify(context.Background(), r))
			}
		})
	}
}

func TestVote_Verify(t *testing.T) {
	peer1 := test.GeneratePeerID(t)
	peer2 := test.GeneratePeerID(t)

	testCases := []struct {
		name    string
		request func(*testing.T) *proposal.Request
		vote    func(*testing.T) *proposal.Vote
		err     error
	}{{
		"peer-id-mismatch",
		func(t *testing.T) *proposal.Request {
			return &proposal.Request{
				Type:      proposal.RemoveNode,
				PeerID:    peer1,
				Challenge: []byte("such challenge very crypto"),
			}
		},
		func(t *testing.T) *proposal.Vote {
			v, err := proposal.NewVote(
				context.Background(),
				test.GeneratePrivateKey(t),
				&proposal.Request{
					Type:      proposal.RemoveNode,
					PeerID:    peer2,
					Challenge: []byte("such challenge very crypto"),
				},
			)
			require.NoError(t, err, "proposal.NewVote()")
			return v
		},
		proposal.ErrInvalidPeerID,
	}, {
		"type-mismatch",
		func(t *testing.T) *proposal.Request {
			return &proposal.Request{
				Type:      proposal.RemoveNode,
				PeerID:    peer1,
				Challenge: []byte("such challenge very crypto"),
			}
		},
		func(t *testing.T) *proposal.Vote {
			v, err := proposal.NewVote(
				context.Background(),
				test.GeneratePrivateKey(t),
				&proposal.Request{
					Type:      proposal.AddNode,
					PeerID:    peer1,
					Challenge: []byte("such challenge very crypto"),
				},
			)
			require.NoError(t, err, "proposal.NewVote()")
			return v
		},
		proposal.ErrInvalidRequestType,
	}, {
		"challenge-mismatch",
		func(t *testing.T) *proposal.Request {
			return &proposal.Request{
				Type:      proposal.RemoveNode,
				PeerID:    peer1,
				Challenge: []byte("such challenge very crypto"),
			}
		},
		func(t *testing.T) *proposal.Vote {
			v, err := proposal.NewVote(
				context.Background(),
				test.GeneratePrivateKey(t),
				&proposal.Request{
					Type:      proposal.RemoveNode,
					PeerID:    peer1,
					Challenge: []byte("much crypto very challenge"),
				},
			)
			require.NoError(t, err, "proposal.NewVote()")
			return v
		},
		proposal.ErrInvalidChallenge,
	}, {
		"invalid-signature",
		func(t *testing.T) *proposal.Request {
			return &proposal.Request{
				Type:      proposal.RemoveNode,
				PeerID:    peer1,
				Challenge: []byte("such challenge very crypto"),
			}
		},
		func(t *testing.T) *proposal.Vote {
			v, err := proposal.NewVote(
				context.Background(),
				test.GeneratePrivateKey(t),
				&proposal.Request{
					Type:      proposal.RemoveNode,
					PeerID:    peer1,
					Challenge: []byte("such challenge very crypto"),
				},
			)
			require.NoError(t, err, "proposal.NewVote()")
			require.NotNil(t, v)
			require.NotNil(t, v.Signature)
			require.NotNil(t, v.Signature.Signature)

			v.Signature.Signature[4] = 0
			v.Signature.Signature[7] = 0

			return v
		},
		proposal.ErrInvalidSignature,
	}, {
		"valid-vote",
		func(t *testing.T) *proposal.Request {
			return &proposal.Request{
				Type:      proposal.RemoveNode,
				PeerID:    peer1,
				Challenge: []byte("such challenge very crypto"),
			}
		},
		func(t *testing.T) *proposal.Vote {
			v, err := proposal.NewVote(
				context.Background(),
				test.GeneratePrivateKey(t),
				&proposal.Request{
					Type:      proposal.RemoveNode,
					PeerID:    peer1,
					Challenge: []byte("such challenge very crypto"),
				},
			)
			require.NoError(t, err, "proposal.NewVote()")
			return v
		},
		nil,
	}}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			r := tt.request(t)
			v := tt.vote(t)
			err := v.Verify(context.Background(), r)
			if tt.err != nil {
				assert.EqualError(t, err, tt.err.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestVote_ToProtoVote(t *testing.T) {
	sk := test.GeneratePrivateKey(t)

	req := &proposal.Request{
		Type:      proposal.RemoveNode,
		PeerID:    test.GeneratePeerID(t),
		Challenge: []byte("such challenge"),
	}

	vote, err := proposal.NewVote(context.Background(), sk, req)
	require.NoError(t, err)

	proto := vote.ToProtoVote()
	require.NotNil(t, proto)

	assert.Equal(t, pb.UpdateType_RemoveNode, proto.UpdateType)
	assert.Equal(t, req.Challenge, proto.Challenge)
	assert.Equal(t, []byte(req.PeerID), proto.PeerId)
	require.NotNil(t, proto.Signature)
	assert.Equal(t, vote.Signature.Signature, proto.Signature.Signature)

	vote2 := &proposal.Vote{}
	err = vote2.FromProtoVote(proto)
	require.NoError(t, err)

	assert.Equal(t, vote.Type, vote2.Type)
	assert.Equal(t, vote.PeerID, vote2.PeerID)
	assert.Equal(t, vote.Challenge, vote2.Challenge)
	assert.NoError(t, vote2.Verify(context.Background(), req))
}

func TestVote_MarshalJSON(t *testing.T) {
	req := &proposal.Request{
		Type:      proposal.RemoveNode,
		PeerID:    test.GeneratePeerID(t),
		Challenge: []byte("much ch4ll3ng3"),
	}

	vote, err := proposal.NewVote(context.Background(), test.GeneratePrivateKey(t), req)
	require.NoError(t, err)

	b, err := json.Marshal(vote)
	require.NoError(t, err)

	var deserialized proposal.Vote
	err = json.Unmarshal(b, &deserialized)
	require.NoError(t, err)

	assert.Equal(t, vote.Type, deserialized.Type)
	assert.Equal(t, vote.PeerID, deserialized.PeerID)
	assert.Equal(t, vote.Challenge, deserialized.Challenge)
	assert.NoError(t, deserialized.Verify(context.Background(), req))
}
