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

package trie

import (
	"github.com/pkg/errors"
	"github.com/stratumn/go-node/core/db"
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
