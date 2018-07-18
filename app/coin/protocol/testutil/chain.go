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

package testutil

import (
	"bytes"
	"sync"

	"github.com/stratumn/go-indigonode/app/coin/pb"
	"github.com/stratumn/go-indigonode/app/coin/protocol/chain"
	"github.com/stratumn/go-indigonode/app/coin/protocol/coinutil"
)

// SimpleChain can be used in tests to inject arbitrary blocks.
type SimpleChain struct {
	mu           sync.RWMutex
	blocks       []*pb.Block
	currentBlock *pb.Block
}

// Config returns nothing.
func (c *SimpleChain) Config() *chain.Config {
	return nil
}

// CurrentBlock returns the header of the last block added.
func (c *SimpleChain) CurrentBlock() (*pb.Block, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.currentBlock == nil {
		return nil, chain.ErrBlockNotFound
	}

	return c.currentBlock, nil
}

// CurrentHeader returns the header of the last block added.
func (c *SimpleChain) CurrentHeader() (*pb.Header, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.currentBlock == nil {
		return nil, chain.ErrBlockNotFound
	}

	return c.currentBlock.Header, nil
}

// GetHeadersByNumber returns all headers that have the input BlockNumber.
func (c *SimpleChain) GetHeadersByNumber(number uint64) ([]*pb.Header, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var res []*pb.Header
	for _, b := range c.blocks {
		if b.BlockNumber() == number {
			res = append(res, b.Header)
		}
	}

	if res == nil {
		return nil, chain.ErrBlockNotFound
	}

	return res, nil
}

// GetHeaderByNumber retrieves a header from the main branch by number.
func (c *SimpleChain) GetHeaderByNumber(number uint64) (*pb.Header, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	b := c.currentBlock
	if b.BlockNumber() < number {
		return nil, chain.ErrBlockNotFound
	}

	for b.BlockNumber() > number {
		var err error
		b, err = c.getBlock(b.PreviousHash())
		if err != nil {
			return nil, err
		}
	}

	return b.Header, nil
}

// GetHeaderByHash returns the first header of the block with the given hash.
func (c *SimpleChain) GetHeaderByHash(hash []byte) (*pb.Header, error) {
	b, err := c.GetBlock(hash, 0)
	if err != nil {
		return nil, err
	}

	return b.Header, nil
}

// GetBlockByHash returns the first block with the given header hash.
func (c *SimpleChain) GetBlockByHash(hash []byte) (*pb.Block, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.getBlock(hash)
}

// GetBlock returns the first block with the given header hash.
func (c *SimpleChain) GetBlock(hash []byte, _ uint64) (*pb.Block, error) {
	return c.GetBlockByHash(hash)
}

// GetBlockByNumber retrieves a header from the main branch by number.
func (c *SimpleChain) GetBlockByNumber(number uint64) (*pb.Block, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	b := c.currentBlock
	if b.BlockNumber() < number {
		return nil, chain.ErrBlockNotFound
	}

	for b.BlockNumber() > number {
		var err error
		b, err = c.getBlock(b.PreviousHash())
		if err != nil {
			return nil, err
		}
	}

	return b, nil
}

func (c *SimpleChain) getBlock(hash []byte) (*pb.Block, error) {
	for _, b := range c.blocks {
		blockHash, err := coinutil.HashHeader(b.Header)
		if err != nil {
			return nil, err
		}

		if bytes.Equal(hash, blockHash) {
			return b, nil
		}
	}

	return nil, chain.ErrBlockNotFound
}

// GetParentBlock retrieves the block's parent block.
func (c *SimpleChain) GetParentBlock(header *pb.Header) (*pb.Block, error) {
	for _, b := range c.blocks {
		blockHash, err := coinutil.HashHeader(b.Header)
		if err != nil {
			return nil, err
		}

		if bytes.Equal(header.PreviousHash, blockHash) {
			return b, nil
		}
	}

	return nil, chain.ErrInvalidPreviousBlock
}

// AddBlock adds a block to the chain without any validation.
// It also sets the given block as current block.
func (c *SimpleChain) AddBlock(block *pb.Block) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.blocks = append(c.blocks, block)

	return nil
}

// SetHead sets the given block as chain head.
func (c *SimpleChain) SetHead(block *pb.Block) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.currentBlock = block

	return nil
}
