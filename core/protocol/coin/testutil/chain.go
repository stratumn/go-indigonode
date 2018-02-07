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

	"github.com/stratumn/alice/core/protocol/coin/chain"
	"github.com/stratumn/alice/core/protocol/coin/coinutil"
	pb "github.com/stratumn/alice/pb/coin"
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

	return c.currentBlock, nil
}

// CurrentHeader returns the header of the last block added.
func (c *SimpleChain) CurrentHeader() (*pb.Header, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.currentBlock == nil {
		return nil, nil
	}

	return c.currentBlock.Header, nil
}

// GetHeaderByNumber returns all headers that have the input BlockNumber.
func (c *SimpleChain) GetHeaderByNumber(number uint64) ([]*pb.Header, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var res []*pb.Header
	for _, b := range c.blocks {
		if b.Header.BlockNumber == number {
			res = append(res, b.Header)
		}
	}

	if res == nil {
		return nil, chain.ErrBlockNumberNotFound
	}

	return res, nil
}

// GetHeaderByHash returns the first header of the block with the given hash.
func (c *SimpleChain) GetHeaderByHash(hash []byte) (*pb.Header, error) {
	b, err := c.GetBlock(hash, 0)
	if err != nil {
		return nil, err
	}

	return b.Header, nil
}

// GetBlock returns the first block with the given header hash.
func (c *SimpleChain) GetBlock(hash []byte, _ uint64) (*pb.Block, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	for _, b := range c.blocks {
		blockHash, err := coinutil.HashHeader(b.Header)
		if err != nil {
			return nil, err
		}

		if bytes.Equal(hash, blockHash) {
			return b, nil
		}
	}

	return nil, chain.ErrBlockHashNotFound
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
