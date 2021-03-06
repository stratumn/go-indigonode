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

package trie

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/multiformats/go-multihash"
)

func mustMakeHash(t *testing.T, b58 string) multihash.Multihash {
	h, err := multihash.FromB58String(b58)
	assert.NoErrorf(t, err, "multihash.FromB58String(%s)", b58)

	return h
}

func TestProof(t *testing.T) {
	validProof := Proof{
		ProofNode{
			Key:   []byte{0x01, 0x00},
			Value: []byte("bob"),
		},
		ProofNode{
			ChildHashes: []multihash.Multihash{
				mustMakeHash(t, "QmNqFMMuB5Vy7GFuTzgUJy1uvUVbdEL5xb3vKLUVwhccN3"),
				mustMakeHash(t, "Qmd2U7CRFYZgTRRmeBq3muMUTm9nFJhv85PWfeuSgr7qAm"),
			},
		},
		ProofNode{
			ChildHashes: []multihash.Multihash{
				mustMakeHash(t, "Qmbr6LRmh8NHx4GdnLAZa3LWKCx7Wgfs5ZEuKv3N9skFJo"),
				mustMakeHash(t, "Qmf4Uz9HUk8gTQPR8b2a9fAbuqYHH5DHNwiaHbuoff9Lp8"),
			},
		},
		ProofNode{
			ChildHashes: []multihash.Multihash{
				mustMakeHash(t, "Qmc6BqbcodSzMtsCj5fNxFvrzvoRwvgoZvpJxE4ZBvn413"),
			},
		},
	}

	invalidProof := Proof{
		ProofNode{
			Key:   []byte{0x01, 0x00},
			Value: []byte("bob"),
		},
		ProofNode{
			ChildHashes: []multihash.Multihash{
				mustMakeHash(t, "Qmd2U7CRFYZgTRRmeBq3muMUTm9nFJhv85PWfeuSgr7qAm"),
			},
		},
		ProofNode{
			ChildHashes: []multihash.Multihash{
				mustMakeHash(t, "Qmbr6LRmh8NHx4GdnLAZa3LWKCx7Wgfs5ZEuKv3N9skFJo"),
				mustMakeHash(t, "Qmf4Uz9HUk8gTQPR8b2a9fAbuqYHH5DHNwiaHbuoff9Lp8"),
			},
		},
		ProofNode{
			ChildHashes: []multihash.Multihash{
				mustMakeHash(t, "Qmc6BqbcodSzMtsCj5fNxFvrzvoRwvgoZvpJxE4ZBvn413"),
			},
		},
	}

	tests := []struct {
		name   string
		proof  Proof
		merkle string
		key    string
		val    string
		err    error
	}{{
		"valid",
		validProof,
		"QmVmzmEHEZfBunERTMJU7kVEP3z5EWiJcRS9eqEtF62TTm",
		"0100",
		"bob",
		nil,
	}, {
		"invalid-key",
		validProof,
		"QmVmzmEHEZfBunERTMJU7kVEP3z5EWiJcRS9eqEtF62TTm",
		"0102",
		"bob",
		ErrInvalidKey,
	}, {
		"invalid-value",
		validProof,
		"QmVmzmEHEZfBunERTMJU7kVEP3z5EWiJcRS9eqEtF62TTm",
		"0100",
		"alice",
		ErrInvalidValue,
	}, {
		"invalid-merkle-root",
		validProof,
		"Qmbr6LRmh8NHx4GdnLAZa3LWKCx7Wgfs5ZEuKv3N9skFJo",
		"0100",
		"bob",
		ErrInvalidMerkleRoot,
	}, {
		"invalid",
		invalidProof,
		"QmVmzmEHEZfBunERTMJU7kVEP3z5EWiJcRS9eqEtF62TTm",
		"0100",
		"bob",
		ErrChildNotFound,
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mr := mustMakeHash(t, tt.merkle)

			k, err := hex.DecodeString(tt.key)
			require.NoError(t, err, "hex.DecodeString(tt.key)")

			err = tt.proof.Verify(mr, k, []byte(tt.val))

			if tt.err == nil {
				assert.NoError(t, err, "tt.proof.Verify()")
			} else {
				assert.EqualError(t, err, tt.err.Error())
			}
		})
	}
}

func TestProof_Proto(t *testing.T) {
	want := Proof{
		ProofNode{
			Key:   []byte{0x01, 0x00},
			Value: []byte("bob"),
		},
		ProofNode{
			ChildHashes: []multihash.Multihash{
				mustMakeHash(t, "QmNqFMMuB5Vy7GFuTzgUJy1uvUVbdEL5xb3vKLUVwhccN3"),
				mustMakeHash(t, "Qmd2U7CRFYZgTRRmeBq3muMUTm9nFJhv85PWfeuSgr7qAm"),
			},
		},
		ProofNode{
			ChildHashes: []multihash.Multihash{
				mustMakeHash(t, "Qmbr6LRmh8NHx4GdnLAZa3LWKCx7Wgfs5ZEuKv3N9skFJo"),
				mustMakeHash(t, "Qmf4Uz9HUk8gTQPR8b2a9fAbuqYHH5DHNwiaHbuoff9Lp8"),
			},
		},
		ProofNode{
			ChildHashes: []multihash.Multihash{
				mustMakeHash(t, "Qmc6BqbcodSzMtsCj5fNxFvrzvoRwvgoZvpJxE4ZBvn413"),
			},
		},
	}

	msg := want.ToProto()
	got := NewProofFromProto(msg)

	assert.Equal(t, want, got)
}
