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
	"testing"

	"github.com/stratumn/alice/core/protocol/bootstrap/proposal"
	"github.com/stratumn/alice/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	ic "gx/ipfs/Qme1knMqwt1hKZbc1BmQFmnm9f36nyQGwXxPGVpVJ9rMK5/go-libp2p-crypto"
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
			v, err := proposal.NewVote(tt.key(t), r)
			if tt.err != nil {
				assert.EqualError(t, err, tt.err.Error())
			} else {
				require.NoError(t, err)
				require.NotNil(t, v)
				assert.Equal(t, r.Type, v.Type)
				assert.Equal(t, r.PeerID, v.PeerID)
				assert.Equal(t, r.Challenge, v.Challenge)
				assert.NoError(t, v.Verify(r))
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
			err := v.Verify(r)
			if tt.err != nil {
				assert.EqualError(t, err, tt.err.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
