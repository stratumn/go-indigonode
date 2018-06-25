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
