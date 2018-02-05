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

	ptypes "github.com/gogo/protobuf/types"
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
			previousHeader := &pb.Header{BlockNumber: 41}
			previousHash, err := coinutils.HashHeader(previousHeader)
			require.NoError(t, err, "coinutils.HashHeader()")

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
			previousHash, err := coinutils.HashHeader(previousBlock.Header)
			require.NoError(t, err, "coinutils.HashHeader()")

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
			previousHash, err := coinutils.HashHeader(previousBlock.Header)
			require.NoError(t, err, "coinutils.HashHeader()")

			chain := &testutil.SimpleChain{}
			require.NoError(t, chain.AddBlock(previousBlock), "chain.AddBlock()")

			h := &pb.Header{
				BlockNumber:  4,
				PreviousHash: previousHash,
				Nonce:        226, // Pre-computed nonce that meets a difficulty of 1
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

func TestHashEngine_Prepare(t *testing.T) {
	tests := []struct {
		name string
		run  func(*testing.T, engine.Engine)
	}{{
		"chain-error",
		func(t *testing.T, e engine.Engine) {
			h := &pb.Header{}
			// An empty chain will have no current header so it should fail.
			err := e.Prepare(&testutil.SimpleChain{}, h)
			assert.EqualError(t, err, engine.ErrInvalidChain.Error(), "e.Prepare()")
		},
	}, {
		"valid-header",
		func(t *testing.T, e engine.Engine) {
			chain := &testutil.SimpleChain{}
			genesis := &pb.Block{Header: &pb.Header{
				BlockNumber: 0,
				Timestamp:   ptypes.TimestampNow(),
				Nonce:       0,
			}}

			genesisHash, err := coinutils.HashHeader(genesis.Header)
			assert.NoError(t, err, "coinutils.HashHeader()")

			chain.AddBlock(genesis)

			h := &pb.Header{}
			err = e.Prepare(chain, h)
			assert.NoError(t, err, "e.Prepare()")

			assert.Equal(t, int32(1), h.Version, "h.Version")
			assert.Equal(t, uint64(1), h.BlockNumber, "h.BlockNumber")
			assert.EqualValues(t, genesisHash, h.PreviousHash, "h.PreviousHash")
			assert.InDelta(
				t,
				genesis.Header.Timestamp.GetSeconds(),
				h.Timestamp.GetSeconds(),
				1.0,
				"h.Timestamp",
			)
			assert.Nil(t, h.MerkleRoot, "h.MerkleRoot")
			assert.Equal(t, uint64(0), h.Nonce, "h.Nonce")
		},
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := engine.NewHashEngine(42)
			tt.run(t, e)
		})
	}
}
