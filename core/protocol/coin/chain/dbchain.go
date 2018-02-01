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

package chain

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"

	"github.com/pkg/errors"
	"github.com/stratumn/alice/core/protocol/coin/db"
	pb "github.com/stratumn/alice/pb/coin"
)

// prefixes for db keys
var (
	lastBlockKey    = []byte("LastBlock")
	blockPrefix     = []byte("b") // blockPrefix + num -> block
	numToHashPrefix = []byte("n") // numToHashPrefix + num -> []hash
)

/*
dbChain implements the Chain interface with a given DB
We store 3 types of values for now (see prefixes):
	- serialized blocks indexed by hash
	- mapping between block numbers and corresponding hashes (1 to many)
	- last block
*/
type dbChain struct {
	db db.DB
}

// NewDBChain returns a new blockchain using a given DB instance
func NewDBChain(db db.DB) Chain {
	return &dbChain{db: db}
}

func (dbChain) Config() *Config {
	return &Config{}
}

func (c *dbChain) GetBlock(hash []byte, number uint64) (*pb.Block, error) {
	// We don't really need both hash and number for now
	return c.dbGetBlock(append(blockPrefix, hash...))
}

func (c *dbChain) CurrentHeader() (*pb.Header, error) {
	block, err := c.dbGetBlock(lastBlockKey)
	if err != nil {
		return nil, err
	}

	return block.Header, nil
}

func (c *dbChain) GetHeaderByNumber(number uint64) ([]*pb.Header, error) {
	hashes, err := c.dbGetHashes(number)
	if err != nil {
		return nil, err
	}

	res := make([]*pb.Header, len(hashes))

	for i, h := range hashes {
		block, err := c.dbGetBlock(append(blockPrefix, h...))
		if err != nil {
			return nil, err
		}
		res[i] = block.Header
	}

	return res, nil
}

func (c *dbChain) GetHeaderByHash(hash []byte) (*pb.Header, error) {
	block, err := c.dbGetBlock(append(blockPrefix, hash...))
	if err != nil {
		return nil, err
	}

	return block.Header, nil
}

func (c *dbChain) AddBlock(block *pb.Block) error {
	// Once here, we consider the block to valid and legit
	// We just add it to the chain

	tx, err := c.db.Transaction()
	if err != nil {
		return err
	}

	b, err := json.Marshal(block)
	h := sha256.Sum256(b)
	n := block.Header.BlockNumber

	// Add block to the chain
	if err != nil {
		return err
	}
	if err = tx.Put(append(blockPrefix, h[:]...), b); err != nil {
		return err
	}

	// Add block hash to the mapping
	hashes, err := c.dbGetHashes(n)
	if errors.Cause(err) == ErrBlockNumberNotFound {
		hashes = make([][]byte, 0)
	} else if err != nil {
		return err
	}

	hashes = append(hashes, h[:])
	hs, err := json.Marshal(hashes)
	if err != nil {
		return err
	}

	if err = tx.Put(append(numToHashPrefix, encodeUint64(n)...), hs); err != nil {
		return err
	}

	// Update LastBlock
	if err = tx.Put(lastBlockKey, b); err != nil {
		return err
	}

	return tx.Commit()
}

// Get a value from the DB and deserialize it into a pb.Block
func (c *dbChain) dbGetBlock(idx []byte) (*pb.Block, error) {
	b, err := c.db.Get(idx)
	if errors.Cause(err) == db.ErrNotFound {
		return nil, ErrBlockHashNotFound
	}

	if err != nil {
		return nil, err
	}

	block := &pb.Block{}
	err = json.Unmarshal(b, block)
	if err != nil {
		return nil, err
	}

	return block, nil
}

// Get a value from the DB and deserialize it into a pb.Block
func (c *dbChain) dbGetHashes(number uint64) ([][]byte, error) {
	b, err := c.db.Get(append(numToHashPrefix, encodeUint64(number)...))
	if errors.Cause(err) == db.ErrNotFound {
		return nil, ErrBlockNumberNotFound
	}

	if err != nil {
		return nil, err
	}

	hashes := [][]byte{}
	err = json.Unmarshal(b, &hashes)
	if err != nil {
		return nil, err
	}

	return hashes, nil
}

// encodeUint64 encodes an uint64 to a buffer.
func encodeUint64(value uint64) []byte {
	v := make([]byte, 8)
	binary.LittleEndian.PutUint64(v, value)

	return v
}
