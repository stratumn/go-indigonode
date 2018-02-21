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
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"

	"github.com/pkg/errors"
	"github.com/stratumn/alice/core/protocol/coin/coinutil"
	"github.com/stratumn/alice/core/protocol/coin/db"
	pb "github.com/stratumn/alice/pb/coin"

	logging "gx/ipfs/QmSpJByNKFX1sCsHBEp3R73FL4NF6FnQTEGyNAXHm2GS52/go-log"
)

// prefixes for db keys.
var (
	lastBlockKey    = []byte("LastBlock")
	blockPrefix     = []byte("b") // blockPrefix + hash -> block
	numToHashPrefix = []byte("n") // numToHashPrefix + num -> []hash
)

// log is the logger for the chain.
var log = logging.Logger("coin.chain")

/*
chainDB implements the Chain interface with a given DB.
We store 3 types of values for now (see prefixes):
	- serialized blocks indexed by hash.
	- mapping between block numbers and corresponding header hashes (1 to many).
	- last block.
*/
type chainDB struct {
	db     db.DB
	prefix []byte

	// mainChain is a lazy loaded cache that will contain refs to main branch's blocks
	// We generate it only once requested.
	// If there is a reorg, we just erase it and wait for next request.
	mainBranch map[uint64][]byte
}

// Opt is an option for Chain
type Opt func(*chainDB)

// OptPrefix sets a prefix for all the database keys.
var OptPrefix = func(prefix []byte) Opt {
	return func(c *chainDB) {
		c.prefix = prefix
	}
}

// OptGenesisBlock sets the genesis block for the chain.
var OptGenesisBlock = func(genesis *pb.Block) Opt {
	if genesis.BlockNumber() != 0 {
		panic("Genesis block should have number 0")
	}

	return func(c *chainDB) {
		if err := c.AddBlock(genesis); err != nil {
			panic(err)
		}

		if err := c.SetHead(genesis); err != nil {
			panic(err)
		}
	}
}

// NewChainDB returns a new blockchain using a given DB instance.
func NewChainDB(db db.DB, opts ...Opt) Chain {
	c := &chainDB{db: db}

	for _, o := range opts {
		o(c)
	}

	return c
}

// Config retrieves the blockchain's chain configuration.
func (c *chainDB) Config() *Config {
	return &Config{}
}

// GetBlock retrieves a block from the database by header hash and number.
func (c *chainDB) GetBlock(hash []byte, number uint64) (*pb.Block, error) {
	// Get the block from the hash
	block, err := c.GetBlockByHash(hash)
	if err != nil {
		return nil, err
	}

	// Check that the block number is correct
	if block.BlockNumber() != number {
		return nil, ErrBlockNumberIncorrect
	}

	return block, nil
}

// GetBlock retrieves a block from the database by header hash and number.
func (c *chainDB) GetBlockByHash(hash []byte) (*pb.Block, error) {
	// Get the block from the hash
	block, err := c.dbGetBlock(c.blockKey(hash))
	if err != nil {
		return nil, err
	}

	return block, nil
}

// CurrentBlock retrieves the current block from the local chain.
func (c *chainDB) CurrentBlock() (*pb.Block, error) {
	return c.dbGetBlock(lastBlockKey)
}

// CurrentHeader retrieves the current header from the local chain.
func (c *chainDB) CurrentHeader() (*pb.Header, error) {
	block, err := c.dbGetBlock(lastBlockKey)
	if err != nil {
		return nil, err
	}

	return block.Header, nil
}

// GetHeadersByNumber retrieves block headers from the database by number.
// In case of forks there might be multiple headers with the same number.
func (c *chainDB) GetHeadersByNumber(number uint64) ([]*pb.Header, error) {
	hashes, err := c.dbGetHashes(number)
	if err != nil {
		return nil, err
	}

	res := make([]*pb.Header, len(hashes))

	for i, h := range hashes {
		block, err := c.dbGetBlock(c.blockKey(h))
		if err != nil {
			return nil, err
		}
		res[i] = block.Header
	}

	return res, nil
}

// GetHeaderByNumber retrieves a header from the main branch by number.
func (c *chainDB) GetHeaderByNumber(number uint64) (*pb.Header, error) {
	if c.mainBranch == nil {
		if err := c.generateMainBranch(); err != nil {
			return nil, err
		}
	}

	hash, ok := c.mainBranch[number]
	if !ok {
		return nil, ErrBlockNumberNotFound
	}

	block, err := c.dbGetBlock(c.blockKey(hash))
	if err != nil {
		return nil, err
	}

	return block.Header, nil
}

// GetHeaderByHash retrieves a block header from the database by its hash.
func (c *chainDB) GetHeaderByHash(hash []byte) (*pb.Header, error) {
	block, err := c.dbGetBlock(c.blockKey(hash))
	if err != nil {
		return nil, err
	}

	return block.Header, nil
}

// GetParentBlock retrieves the parent block.
func (c *chainDB) GetParentBlock(header *pb.Header) (*pb.Block, error) {
	return c.GetBlock(header.PreviousHash, header.BlockNumber-1)
}

// AddBlock adds a block to the chain.
// It assumes that the block has been validated.
// We still check that previous hash points to the block before this one.
func (c *chainDB) AddBlock(block *pb.Block) error {

	// Check previous block
	if err := c.checkAddBlock(block.Header); err != nil {
		return err
	}

	// Add the block
	tx, err := c.db.Transaction()
	if err != nil {
		return err
	}

	if err = c.doAddBlock(tx, block); err != nil {
		tx.Discard()
		return err
	}

	return tx.Commit()
}

// checkAddBlock checks that the previous block exists
// and that the block number is correct.
func (c *chainDB) checkAddBlock(h *pb.Header) error {
	// Is this the first block ?
	if h.PreviousHash == nil && h.BlockNumber == 0 {
		return nil
	}

	// Check previous block
	prevBlock, err := c.dbGetBlock(c.blockKey(h.PreviousHash))
	if errors.Cause(err) == ErrBlockNumberIncorrect {
		return ErrInvalidPreviousBlock
	}
	if err != nil {
		return err
	}
	if prevBlock.BlockNumber() != h.BlockNumber-1 {
		return ErrInvalidPreviousBlock
	}

	return nil
}

// doAddBlock actually prepares the transaction to add a block.
func (c *chainDB) doAddBlock(tx db.Transaction, block *pb.Block) error {
	b, err := block.Marshal()
	if err != nil {
		return err
	}

	h, err := coinutil.HashHeader(block.Header)
	if err != nil {
		return err
	}

	// Add block to the chain.
	if err = tx.Put(c.blockKey(h), b); err != nil {
		return err
	}

	// Add header hash to the mapping.
	n := block.BlockNumber()
	hashes, err := c.dbGetHashes(n)
	if errors.Cause(err) == ErrBlockNumberNotFound {
		hashes = make([][]byte, 0)
	} else if err != nil {
		return err
	}

	hashes = append(hashes, h)
	hs, err := serializeHashes(hashes)
	if err != nil {
		return err
	}

	return tx.Put(c.numToHashKey(n), hs)
}

// SetHead sets the head of the chain.
func (c *chainDB) SetHead(block *pb.Block) (err error) {
	e := log.EventBegin(context.Background(), "SetHead", &logging.Metadata{
		"BlockNumber":  block.BlockNumber(),
		"PreviousHash": block.PreviousHash(),
		"Nonce":        block.Nonce(),
		"TxCount":      len(block.Transactions),
	})
	defer func() {
		if err != nil {
			e.SetError(err)
		}

		e.Done()
	}()

	b, err := block.Marshal()
	if err != nil {
		return err
	}

	h, err := coinutil.HashHeader(block.Header)
	if err != nil {
		return err
	}

	// Check that block is in the chain.
	_, err = c.db.Get(c.blockKey(h))
	if errors.Cause(err) == db.ErrNotFound {
		return ErrBlockHashNotFound
	}

	if err != nil {
		return err
	}

	// If mainBranch is not empty and `block` is not the genesis block,
	// check if we switched branches and update mainBranch cache accordingly.
	if len(c.mainBranch) > 0 && block.BlockNumber() > 0 {
		cur, err := c.CurrentHeader()
		if err != nil {
			return err
		}
		hb, err := coinutil.HashHeader(cur)
		if err != nil {
			return err
		}

		if !bytes.Equal(hb, block.PreviousHash()) {
			// If we switched branches, reset the main branch cache
			c.mainBranch = nil
		}
	}

	// Update LastBlock.
	return c.db.Put(lastBlockKey, b)
}

// Get a value from the DB and deserialize it into a pb.Block.
func (c *chainDB) dbGetBlock(idx []byte) (*pb.Block, error) {
	b, err := c.db.Get(idx)
	if errors.Cause(err) == db.ErrNotFound {
		return nil, ErrBlockHashNotFound
	}

	if err != nil {
		return nil, err
	}

	block := &pb.Block{}
	err = block.Unmarshal(b)
	if err != nil {
		return nil, err
	}

	return block, nil
}

// Get the list of block hashes for a block height.
func (c *chainDB) dbGetHashes(number uint64) ([][]byte, error) {
	b, err := c.db.Get(c.numToHashKey(number))
	if errors.Cause(err) == db.ErrNotFound {
		return nil, ErrBlockNumberNotFound
	}

	if err != nil {
		return nil, err
	}

	return deserializeHashes(b)
}

// generateMainBranch creates the map that contains
// the refs to the main branch's blocks.
func (c *chainDB) generateMainBranch() error {
	mainBranch := map[uint64][]byte{}

	header, err := c.CurrentHeader()
	if err != nil {
		return err
	}
	hash, err := coinutil.HashHeader(header)
	if err != nil {
		return err
	}
	mainBranch[header.BlockNumber] = hash

	for header.BlockNumber > 0 {
		hash = header.PreviousHash
		header, err = c.GetHeaderByHash(hash)
		if err != nil {
			return err
		}
		mainBranch[header.BlockNumber] = hash
	}

	c.mainBranch = mainBranch
	return nil
}

// blockKey returns the key corresponding to a block in th DB.
func (c *chainDB) blockKey(hash []byte) []byte {
	return append(append(c.prefix, blockPrefix...), hash...)
}

// numToHashKey returns the key corresponding to
// a number to hash mapping in the DB.
func (c *chainDB) numToHashKey(num uint64) []byte {
	return append(append(c.prefix, numToHashPrefix...), encodeUint64(num)...)
}

// encodeUint64 encodes an uint64 to a buffer.
func encodeUint64(value uint64) []byte {
	v := make([]byte, 8)
	binary.LittleEndian.PutUint64(v, value)

	return v
}

// Serialize a list of hashes into a byte array.
func serializeHashes(h [][]byte) ([]byte, error) {
	return json.Marshal(h)
}

// Deserialize a byte array into a list of hashes.
func deserializeHashes(b []byte) ([][]byte, error) {
	hashes := [][]byte{}
	err := json.Unmarshal(b, &hashes)
	if err != nil {
		return nil, err
	}

	return hashes, nil
}
