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

//go:generate mockgen -package mockchain -destination mockchain/mockchainreader.go github.com/stratumn/alice/core/protocol/coin/chain Reader
//go:generate mockgen -package mockchain -destination mockchain/mockchainwriter.go github.com/stratumn/alice/core/protocol/coin/chain Writer
//go:generate mockgen -package mockchain -destination mockchain/mockchain.go github.com/stratumn/alice/core/protocol/coin/chain Chain

package chain

import (
	"errors"

	pb "github.com/stratumn/alice/pb/coin"
)

var (
	// ErrBlockHashNotFound is returned when looking for a block that is not in the chain.
	ErrBlockHashNotFound = errors.New("block hash not found in the chain")

	// ErrBlockNumberNotFound is returned when looking for a block that is not in the chain.
	ErrBlockNumberNotFound = errors.New("block number not found in the chain")

	// ErrBlockNumberIncorrect is returned when adding a block with a bad number.
	ErrBlockNumberIncorrect = errors.New("block number does not correspond to hash")

	// ErrInvalidPreviousBlock is returned when adding a block with a bad number or previous hash.
	ErrInvalidPreviousBlock = errors.New("link to previous block is invalid")
)

// Config describes the blockchain's chain configuration.
type Config struct {
}

// Reader defines a small collection of methods needed to access the local
// blockchain.
type Reader interface {
	// Config retrieves the blockchain's chain configuration.
	Config() *Config

	// CurrentHeader retrieves the current header from the local chain.
	CurrentHeader() (*pb.Header, error)

	// GetHeadersByNumber retrieves block headers from the database by number.
	// In case of forks there might be multiple headers with the same number,
	GetHeadersByNumber(number uint64) ([]*pb.Header, error)

	// GetHeaderByNumber retrieves a header from the main branch by number.
	GetHeaderByNumber(number uint64) (*pb.Header, error)

	// GetHeaderByHash retrieves a block header from the database by its hash.
	GetHeaderByHash(hash []byte) (*pb.Header, error)

	// CurrentBlock retrieves the current block from the local chain.
	CurrentBlock() (*pb.Block, error)

	// GetBlock retrieves a block from the database by header hash and number.
	GetBlock(hash []byte, number uint64) (*pb.Block, error)

	// GetBlockByHash retrieves a block from the database by header hash.
	GetBlockByHash(hash []byte) (*pb.Block, error)
}

// Writer defines methods needed to write to the local blockchain.
type Writer interface {
	// AddBlock adds a block to the chain.
	// It assumes that the block has been validated.
	AddBlock(block *pb.Block) error

	// SetHead sets the head of the chain.
	SetHead(block *pb.Block) error
}

// Chain defines methods to interact with the local blockchain.
type Chain interface {
	Reader
	Writer
}
