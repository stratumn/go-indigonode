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

package engine_test

import (
	"testing"

	"github.com/stratumn/alice/core/protocol/coin/coinutils"
	"github.com/stratumn/alice/core/protocol/coin/engine"
	"github.com/stratumn/alice/core/protocol/coin/testutil"
	pb "github.com/stratumn/alice/pb/coin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHashEngine_Difficulty(t *testing.T) {
	e := engine.NewHashEngine(42)
	assert.Equal(t, uint64(42), e.Difficulty(), "e.Difficulty()")

	e = engine.NewHashEngine(0)
	assert.Equal(t, uint64(0), e.Difficulty(), "e.Difficulty()")
}

func TestHashEngine_VerifyHeader(t *testing.T) {
	tests := []struct {
		name       string
		difficulty uint64
		run        func(*testing.T, engine.Engine)
	}{{
		"missing-previous-block",
		0,
		func(t *testing.T, e engine.Engine) {
			previousBlock := &pb.Block{Header: &pb.Header{BlockNumber: 41}}
			previousHash, err := coinutils.HashBlock(previousBlock)
			require.NoError(t, err, "coinutils.HashBlock()")

			h := &pb.Header{
				BlockNumber:  42,
				PreviousHash: previousHash,
			}

			err = e.VerifyHeader(&testutil.SimpleChain{}, h)
			assert.EqualError(t, err, engine.ErrInvalidPreviousBlock.Error(), "e.VerifyHeader()")
		},
	}, {
		"invalid-block-number",
		0,
		func(t *testing.T, e engine.Engine) {
			previousBlock := &pb.Block{Header: &pb.Header{BlockNumber: 3}}
			previousHash, err := coinutils.HashBlock(previousBlock)
			require.NoError(t, err, "coinutils.HashBlock()")

			chain := &testutil.SimpleChain{}
			require.NoError(t, chain.AddBlock(previousBlock), "chain.AddBlock()")

			h := &pb.Header{
				BlockNumber:  5,
				PreviousHash: previousHash,
			}

			err = e.VerifyHeader(chain, h)
			assert.EqualError(t, err, engine.ErrInvalidBlockNumber.Error(), "e.VerifyHeader()")
		},
	}, {
		"difficulty-not-met",
		1,
		func(t *testing.T, e engine.Engine) {
			h := &pb.Header{Nonce: 42}

			err := e.VerifyHeader(nil, h)
			assert.EqualError(t, err, engine.ErrDifficultyNotMet.Error(), "e.VerifyHeader()")
		},
	}, {
		"valid-block",
		1,
		func(t *testing.T, e engine.Engine) {
			previousBlock := &pb.Block{Header: &pb.Header{BlockNumber: 3}}
			previousHash, err := coinutils.HashBlock(previousBlock)
			require.NoError(t, err, "coinutils.HashBlock()")

			chain := &testutil.SimpleChain{}
			require.NoError(t, chain.AddBlock(previousBlock), "chain.AddBlock()")

			h := &pb.Header{
				BlockNumber:  4,
				PreviousHash: previousHash,
				Nonce:        212, // Pre-computed nonce that meets a difficulty of 1
			}

			err = e.VerifyHeader(chain, h)
			assert.NoError(t, err, "e.VerifyHeader()")
		},
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := engine.NewHashEngine(tt.difficulty)
			tt.run(t, e)
		})
	}
}
