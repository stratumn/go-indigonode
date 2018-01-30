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
	ldbopt "github.com/syndtr/goleveldb/leveldb/opt"
)

type levelDB struct {
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

// NewLevelDB creates a new LevelDB key-value database from a LevelDB instance.
func NewLevelDB(ldb *leveldb.DB) DB {
	return &levelDB{ldb: ldb}
}

// OpenLevelDBFile creates a new LevelDB key-value database from a LevelDB
// filename.
//
// Options can be nil. LevelDB has some interesting options to look at to
// improve performance, such as BloomFilter.
func OpenLevelDBFile(filename string, o *ldbopt.Options) (DB, error) {
	ldb, err := leveldb.OpenFile(filename, o)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return NewLevelDB(ldb), nil
}

func (db *levelDB) Get(key []byte) ([]byte, error) {
	v, err := db.ldb.Get(key, nil)

	return v, ldbErr(err)
}

func (db *levelDB) Put(key, value []byte) error {
	return ldbErr(db.ldb.Put(key, value, nil))
}

func (db *levelDB) Delete(key []byte) error {
	tx, err := db.ldb.OpenTransaction()
	if err != nil {
		return ldbErr(err)
	}

	exists, err := tx.Has(key, nil)
	if err != nil {
		tx.Discard()
		return ldbErr(err)
	}

	if !exists {
		tx.Discard()
		return ErrNotFound
	}

	if err := tx.Delete(key, nil); err != nil {
		tx.Discard()
		return ldbErr(err)
	}

	return ldbErr(tx.Commit())
}

func (db *levelDB) Close() error {
	return ldbErr(db.ldb.Close())
}

func (db *levelDB) Batch() Batch {
	return newBatch(db.commit)
}

// commit commits operations from a batch.
func (db *levelDB) commit(ops []interface{}) error {
	tx, err := db.ldb.OpenTransaction()
	if err != nil {
		return ldbErr(err)
	}

	for _, op := range ops {
		switch v := op.(type) {
		case batchPut:
			if err := tx.Put(v.key, v.value, nil); err != nil {
				return ldbErr(err)
			}
		case batchDelete:
			exists, err := tx.Has(v.key, nil)
			if err != nil {
				tx.Discard()
				return ldbErr(err)
			}

			if !exists {
				tx.Discard()
				return ErrNotFound
			}

			if err := tx.Delete(v.key, nil); err != nil {
				tx.Discard()
				return ldbErr(err)
			}
		}
	}

	return ldbErr(tx.Commit())
}
