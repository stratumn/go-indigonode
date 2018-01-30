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

// Batch can commit batched write operations to a key-value database.
type Batch interface {
	Writer

	// Commit commits all the write operations to the database. If an error
	// occurs, the state of the database should not have changed.
	Commit() error
}

// batchPut represents a put operation in a batch.
type batchPut struct {
	key   []byte
	value []byte
}

// batchDelete represents a delete operation in a batch.
type batchDelete struct {
	key []byte
}

// batch implements a batch by storing write operations.
type batch struct {
	ops    []interface{}
	commit func([]interface{}) error
}

// newBatch creates a new batch.
//
// It should be given a function to commit the operations.
func newBatch(commit func([]interface{}) error) *batch {
	return &batch{nil, commit}
}

func (b *batch) Put(key, value []byte) error {
	b.ops = append(b.ops, batchPut{key, value})

	return nil
}

func (b *batch) Delete(key []byte) error {
	b.ops = append(b.ops, batchDelete{key})

	return nil
}

func (b *batch) Commit() error {
	return b.commit(b.ops)
}
