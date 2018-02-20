// Copyright © 2017-2018  Stratumn SAS
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

package trie

import (
	"github.com/pkg/errors"
	"github.com/stratumn/alice/core/protocol/coin/db"
)

// nodeDB deals with storing trie nodes.
type nodeDB interface {
	// Get gets a node from its key from the database. If they key is not
	// found, it should return a null node.
	Get(key []uint8) (node, error)

	// Put inserts the node with the given key in the database.
	Put(key []uint8, n node) error

	// Delete removes the node with the given key from the database.
	Delete(key []uint8) error
}

// dbrwNodeDB implements nodeDB using a db.ReadWritter.
type dbrwNodeDB struct {
	dbrw   db.ReadWriter
	prefix []byte
}

// newDbrwNodeDB creates a new nodeDB from a db.ReadWriter and database key
// prefix.
func newDbrwNodeDB(dbrw db.ReadWriter, prefix []byte) nodeDB {
	l := len(prefix)

	return dbrwNodeDB{dbrw: dbrw, prefix: prefix[:l:l]}
}

func (ndb dbrwNodeDB) Get(key []uint8) (node, error) {
	k, err := ndb.key(key)
	if err != nil {
		return null{}, err
	}

	buf, err := ndb.dbrw.Get(k)
	if err != nil {
		if errors.Cause(err) == db.ErrNotFound {
			return null{}, nil
		}

		return nil, err
	}

	n, _, err := unmarshalNode(buf)

	return n, err
}

func (ndb dbrwNodeDB) Put(key []uint8, n node) error {
	k, err := ndb.key(key)
	if err != nil {
		return err
	}

	v, err := n.MarshalBinary()
	if err != nil {
		return err
	}

	return ndb.dbrw.Put(k, v)
}

func (ndb dbrwNodeDB) Delete(key []uint8) error {
	k, err := ndb.key(key)
	if err != nil {
		return err
	}

	return ndb.dbrw.Delete(k)
}

// key returns the key of a node given its key in the trie.
func (ndb dbrwNodeDB) key(key []uint8) ([]byte, error) {
	k, err := path(newNibsFromNibs(key...)).MarshalBinary()
	if err != nil {
		return nil, err
	}

	return append(ndb.prefix, k...), nil
}

// mapNodeDB implements nodeDB using a map.
type mapNodeDB struct {
	nodes map[string]node
}

// newMapNodeDB creates a new nodeDB using a map.
func newMapNodeDB() *mapNodeDB {
	return &mapNodeDB{nodes: map[string]node{}}
}

func (ndb *mapNodeDB) Get(key []uint8) (node, error) {
	n, ok := ndb.nodes[string(key)]
	if !ok {
		return null{}, nil
	}

	return n.Clone(), nil
}

func (ndb *mapNodeDB) Put(key []uint8, n node) error {
	ndb.nodes[string(key)] = n.Clone()

	return nil
}

func (ndb *mapNodeDB) Delete(key []uint8) error {
	delete(ndb.nodes, string(key))

	return nil
}
