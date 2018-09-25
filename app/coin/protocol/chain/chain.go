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

package chain

import (
	"bytes"
	"errors"

	"github.com/stratumn/go-node/app/coin/pb"
	"github.com/stratumn/go-node/app/coin/protocol/coinutil"
)

var (
	// ErrBlockNotFound is returned when looking for a block that is not in the chain.
	ErrBlockNotFound = errors.New("block not found in the chain")

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

	// GetBlockByNumber retrieves a block from the main branch by number.
	GetBlockByNumber(number uint64) (*pb.Block, error)

	// GetParentBlock retrieves the header's parent block.
	GetParentBlock(header *pb.Header) (*pb.Block, error)
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

// GetPath returns the path from current header
// to a given block (excluding that block).
func GetPath(c Reader, from *pb.Block, to *pb.Block) (rollbacks []*pb.Block, replays []*pb.Block, err error) {

	toParent, err := c.GetParentBlock(to.Header)
	if err != nil {
		return nil, nil, ErrInvalidPreviousBlock
	}

	fromHash, err := coinutil.HashHeader(from.Header)
	if err != nil {
		return nil, nil, err
	}

	fromParent, err := c.GetBlock(fromHash, from.Header.BlockNumber)
	if err != nil {
		return nil, nil, ErrBlockNotFound
	}

	if bytes.Equal(to.Header.PreviousHash, fromHash) {
		return
	}

	// Rewind the to branch until we are at from block height.
	for toParent.Header.BlockNumber > fromParent.Header.BlockNumber {
		replays = append([]*pb.Block{toParent}, replays...)
		toParent, err = c.GetParentBlock(toParent.Header)
		if err != nil {
			return nil, nil, err
		}
	}

	// Rewind the from branch until we are at to block height.
	for fromParent.Header.BlockNumber > toParent.Header.BlockNumber {
		rollbacks = append(rollbacks, fromParent)
		fromParent, err = c.GetParentBlock(fromParent.Header)
		if err != nil {
			return nil, nil, err
		}
	}

	// Rewind both branches until we found the common ancestor.
	for !bytes.Equal(toParent.Header.PreviousHash, fromParent.Header.PreviousHash) {
		replays = append([]*pb.Block{toParent}, replays...)
		toParent, err = c.GetParentBlock(toParent.Header)
		if err != nil {
			return nil, nil, err
		}

		rollbacks = append(rollbacks, fromParent)
		fromParent, err = c.GetParentBlock(fromParent.Header)
		if err != nil {
			return nil, nil, err
		}
	}

	fromHash, err = coinutil.HashHeader(fromParent.Header)
	if err != nil {
		return nil, nil, err
	}
	toHash, err := coinutil.HashHeader(toParent.Header)
	if err != nil {
		return nil, nil, err
	}

	if !bytes.Equal(fromHash, toHash) {
		rollbacks = append(rollbacks, fromParent)
		replays = append([]*pb.Block{toParent}, replays...)
	}

	return
}
