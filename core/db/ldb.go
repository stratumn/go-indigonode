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

package db

import (
	"github.com/pkg/errors"
	"github.com/syndtr/goleveldb/leveldb"
	ldberrors "github.com/syndtr/goleveldb/leveldb/errors"
	ldbiterator "github.com/syndtr/goleveldb/leveldb/iterator"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"github.com/syndtr/goleveldb/leveldb/storage"
	"github.com/syndtr/goleveldb/leveldb/util"
)

// levelDB makes a LevelDB database compatible with the DB interface.
type levelDB struct {
	// We could just use `type levelDB leveldb.DB` but a struct makes it
	// easier to add more functionality in the future.
	ldb *leveldb.DB
}

// ldbErr should be used to wrap LevelDB errors so that a stack is added and
// not-found errors are handled properly.
func ldbErr(err error) error {
	if err == ldberrors.ErrNotFound {
		return ErrNotFound
	}

	return errors.WithStack(err)
}

// NewLevelDB creates a new LevelDB key-value database from an existing LevelDB
// database instance.
func NewLevelDB(ldb *leveldb.DB) DB {
	return levelDB{ldb: ldb}
}

// NewMemDB creates a new in-memory key-value database.
//
// It can be used for testing or for storing small amounts of data, but it
// will run out of memory quickly otherwise.
func NewMemDB(o *opt.Options) (DB, error) {
	ldb, err := leveldb.Open(storage.NewMemStorage(), o)
	if err != nil {
		return nil, ldbErr(err)
	}

	return NewLevelDB(ldb), nil
}

// NewFileDB creates a new LevelDB key-value database from a LevelDB
// file.
//
// Options can be nil. LevelDB has some interesting options to look at to
// improve performance, such as BloomFilter.
func NewFileDB(filename string, o *opt.Options) (DB, error) {
	ldb, err := leveldb.OpenFile(filename, o)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return NewLevelDB(ldb), nil
}

func (db levelDB) Get(key []byte) ([]byte, error) {
	v, err := db.ldb.Get(key, nil)

	return v, ldbErr(err)
}

func (db levelDB) IterateRange(start, stop []byte) Iterator {
	return levelDBIter{
		ldbiter: db.ldb.NewIterator(&util.Range{Start: start, Limit: stop}, nil),
	}
}

func (db levelDB) IteratePrefix(prefix []byte) Iterator {
	return levelDBIter{
		ldbiter: db.ldb.NewIterator(util.BytesPrefix(prefix), nil),
	}
}

func (db levelDB) Put(key, value []byte) error {
	return ldbErr(db.ldb.Put(key, value, nil))
}

func (db levelDB) Delete(key []byte) error {
	return ldbErr(db.ldb.Delete(key, nil))
}

func (db levelDB) Batch() Batch {
	return &leveldb.Batch{}
}

func (db levelDB) Write(batch Batch) error {
	b, ok := batch.(*leveldb.Batch)
	if !ok {
		return ErrInvalidBatch
	}

	return ldbErr(db.ldb.Write(b, nil))
}

func (db levelDB) Transaction() (Transaction, error) {
	tx, err := db.ldb.OpenTransaction()

	return levelDBTx{ldbtx: tx}, errors.WithStack(err)
}

func (db levelDB) Close() error {
	return ldbErr(db.ldb.Close())
}

// levelDBTx makes a LevelDB transaction compatible with the Transaction
// interface.
type levelDBTx struct {
	ldbtx *leveldb.Transaction
}

func (tx levelDBTx) Get(key []byte) ([]byte, error) {
	v, err := tx.ldbtx.Get(key, nil)

	return v, ldbErr(err)
}

func (tx levelDBTx) IterateRange(start, stop []byte) Iterator {
	return levelDBIter{
		ldbiter: tx.ldbtx.NewIterator(&util.Range{Start: start, Limit: stop}, nil),
	}
}

func (tx levelDBTx) IteratePrefix(prefix []byte) Iterator {
	return levelDBIter{
		ldbiter: tx.ldbtx.NewIterator(util.BytesPrefix(prefix), nil),
	}
}

func (tx levelDBTx) Put(key, value []byte) error {
	return ldbErr(tx.ldbtx.Put(key, value, nil))
}

func (tx levelDBTx) Delete(key []byte) error {
	return ldbErr(tx.ldbtx.Delete(key, nil))
}

func (tx levelDBTx) Batch() Batch {
	return &leveldb.Batch{}
}

func (tx levelDBTx) Write(batch Batch) error {
	b, ok := batch.(*leveldb.Batch)
	if !ok {
		return ErrInvalidBatch
	}

	return ldbErr(tx.ldbtx.Write(b, nil))
}

func (tx levelDBTx) Commit() error {
	return ldbErr(tx.ldbtx.Commit())
}

func (tx levelDBTx) Discard() {
	tx.ldbtx.Discard()
}

// levelDBIter makes a LevelDB iterator compatible with the Iterator interface.
type levelDBIter struct {
	ldbiter ldbiterator.Iterator
}

// Next returns whether there are keys left.
func (i levelDBIter) Next() (bool, error) {
	return i.ldbiter.Next(), nil
}

// Key returns the key of the current entry.
func (i levelDBIter) Key() []byte {
	return i.ldbiter.Key()
}

// Value returns the value of the current entry.
func (i levelDBIter) Value() []byte {
	return i.ldbiter.Value()
}

// Release needs to be called to free the iterator.
func (i levelDBIter) Release() {
	i.ldbiter.Release()
}
