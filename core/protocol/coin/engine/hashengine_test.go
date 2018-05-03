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
	"context"
	"testing"

	ptypes "github.com/gogo/protobuf/types"
	"github.com/stratumn/alice/core/protocol/coin/coinutil"
	"github.com/stratumn/alice/core/protocol/coin/engine"
	"github.com/stratumn/alice/core/protocol/coin/testutil"
	txtest "github.com/stratumn/alice/core/protocol/coin/testutil/transaction"
	pb "github.com/stratumn/alice/pb/coin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHashEngine_Difficulty(t *testing.T) {
	e := engine.NewHashEngine("alice", 42, 5)
	assert.Equal(t, uint64(42), e.Difficulty(), "e.Difficulty()")

	e = engine.NewHashEngine("alice", 0, 5)
	assert.Equal(t, uint64(0), e.Difficulty(), "e.Difficulty()")
}

func TestHashEngine_Reward(t *testing.T) {
	e := engine.NewHashEngine("alice", 3, 42)
	assert.Equal(t, uint64(42), e.Reward(), "e.Reward()")
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
			previousHash, err := coinutil.HashHeader(previousHeader)
			require.NoError(t, err, "coinutil.HashHeader()")

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
			previousHash, err := coinutil.HashHeader(previousBlock.Header)
			require.NoError(t, err, "coinutil.HashHeader()")

			chain := &testutil.SimpleChain{}
			require.NoError(t, chain.AddBlock(previousBlock), "chain.AddBlock()")
			require.NoError(t, chain.SetHead(previousBlock), "chain.SetHead()")

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
			previousHash, err := coinutil.HashHeader(previousBlock.Header)
			require.NoError(t, err, "coinutil.HashHeader()")

			chain := &testutil.SimpleChain{}
			require.NoError(t, chain.AddBlock(previousBlock), "chain.AddBlock()")
			require.NoError(t, chain.SetHead(previousBlock), "chain.SetHead()")

			h := &pb.Header{
				BlockNumber:  4,
				PreviousHash: previousHash,
				Nonce:        394, // Pre-computed nonce that meets a difficulty of 1
			}

			err = e.VerifyHeader(chain, h)
			assert.NoError(t, err, "e.VerifyHeader()")
		},
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := engine.NewHashEngine("", tt.difficulty, 5)
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

			genesisHash, err := coinutil.HashHeader(genesis.Header)
			assert.NoError(t, err, "coinutil.HashHeader()")

			chain.AddBlock(genesis)
			chain.SetHead(genesis)

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
			e := engine.NewHashEngine("alice", 42, 5)
			tt.run(t, e)
		})
	}
}

func TestHashEngine_Finalize(t *testing.T) {
	ctx := context.Background()
	chain := &testutil.SimpleChain{}

	genesis := &pb.Block{Header: &pb.Header{Version: 1}}
	genesisHash, err := coinutil.HashHeader(genesis.Header)
	require.NoError(t, err, "coinutil.HashHeader()")
	require.NoError(t, chain.AddBlock(genesis), "chain.AddBlock()")
	require.NoError(t, chain.SetHead(genesis), "chain.SetHead()")

	firstBlock := &pb.Block{
		Header:       &pb.Header{Version: 1, BlockNumber: 1, PreviousHash: genesisHash},
		Transactions: []*pb.Transaction{txtest.NewTransaction(t, 4, 1, 2)},
	}
	require.NoError(t, chain.AddBlock(firstBlock), "chain.AddBlock()")
	require.NoError(t, chain.SetHead(firstBlock), "chain.SetHead()")

	// Static reward to miner in reward transaction.
	testReward := uint64(5)

	tests := []struct {
		name       string
		difficulty uint64
		run        func(*testing.T, engine.Engine)
	}{{
		"invalid-block-number",
		0,
		func(t *testing.T, e engine.Engine) {
			invalidBlockNumber := &pb.Header{Version: 1, BlockNumber: 2, PreviousHash: genesisHash}

			_, err := e.Finalize(ctx, chain, invalidBlockNumber, nil, nil)
			assert.EqualError(t, err, engine.ErrInvalidBlockNumber.Error(), "e.Finalize()")
		},
	}, {
		"missing-previous-block",
		0,
		func(t *testing.T, e engine.Engine) {
			invalidPrevious := &pb.Header{Version: 1, BlockNumber: 2, PreviousHash: []byte("hello")}

			_, err := e.Finalize(ctx, chain, invalidPrevious, nil, nil)
			assert.EqualError(t, err, engine.ErrInvalidPreviousBlock.Error(), "e.Finalize()")
		},
	}, {
		"block-reward",
		0,
		func(t *testing.T, e engine.Engine) {
			block, err := e.Finalize(ctx, chain, firstBlock.Header, nil, firstBlock.Transactions)
			assert.NoError(t, err, "e.Finalize()")
			assert.NotNil(t, block.Header.MerkleRoot, "block.Header.MerkleRoot")
			assert.Equal(t, uint64(1), block.Header.BlockNumber, "block.Header.BlockNumber")
			assert.EqualValues(t, genesisHash, block.Header.PreviousHash, "block.Header.PreviousHash")

			// A block reward tx should be added.
			assert.Len(t, block.Transactions, 2, "block.Transactions")

			var blockReward *pb.Transaction
			for _, tx := range block.Transactions {
				if tx.From == nil {
					blockReward = tx
					break
				}
			}

			txFees := uint64(0)
			for _, tx := range firstBlock.Transactions {
				txFees += tx.Fee
			}

			assert.NotNil(t, blockReward, "blockReward")
			assert.Equal(t, testReward+txFees, blockReward.Value, "blockReward.Value")
			assert.Nil(t, blockReward.Signature, "blockReward.Signature")

			assert.EqualValues(t, []byte("alice"), blockReward.To, "blockReward.To")
		},
	}, {
		"block-nonce-pow",
		10,
		func(t *testing.T, e engine.Engine) {
			block, err := e.Finalize(ctx, chain, firstBlock.Header, nil, firstBlock.Transactions)
			assert.NoError(t, err, "e.Finalize()")
			assert.True(t, block.Header.Nonce > 0, "block.Header.Nonce > 0")
		},
	}, {
		"block-pow-cancel",
		42,
		func(t *testing.T, e engine.Engine) {
			childCtx, cancel := context.WithCancel(ctx)
			errChan := make(chan error)
			go func() {
				_, err := e.Finalize(childCtx, chain, firstBlock.Header, nil, firstBlock.Transactions)
				errChan <- err
			}()

			cancel()
			err := <-errChan

			assert.EqualError(t, err, context.Canceled.Error(), "cancel()")
		},
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := engine.NewHashEngine("alice", tt.difficulty, testReward)
			tt.run(t, e)
		})
	}
}
