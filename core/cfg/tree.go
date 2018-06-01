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

package cfg

import (
	"bytes"
	"fmt"
	"reflect"

	toml "github.com/pelletier/go-toml"
	"github.com/pkg/errors"
)

// Tree is used to modify the configuration tree.
type Tree struct {
	tree *toml.Tree
}

// TreeFromMap creates a tree from a map.
func TreeFromMap(m map[string]interface{}) (*Tree, error) {
	t, err := toml.TreeFromMap(m)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return &Tree{t}, nil
}

// Get returns the value of a key.
//
// The key can be a path such as "core.boot_service".
func (t *Tree) Get(key string) interface{} {
	val := t.tree.Get(key)

	switch v := val.(type) {
	case (*toml.Tree):
		return &Tree{v}
	case ([]*toml.Tree):
		trees := make([]*Tree, len(v))
		for i, child := range v {
			trees[i] = &Tree{child}
		}
		return trees
	default:
		return val
	}
}

// GetDefault returns the value of a key or a default value.
//
// The key can be a path such as "core.boot_service".
func (t *Tree) GetDefault(key string, def interface{}) interface{} {
	return t.tree.GetDefault(key, def)
}

// Set sets the value of a key.
//
// The key can be a path such as "core.boot_service".
func (t *Tree) Set(key string, val interface{}) (err error) {
	switch v := val.(type) {
	case *Tree:
		t.tree.Set(key, "", false, v.tree)
		return
	case []*Tree:
		trees := make([]*toml.Tree, len(v))
		for i, child := range v {
			trees[i] = child.tree
		}
		t.tree.Set(key, "", false, trees)
		return
	}

	// Struct must be converted to a tree before calling tree.Set().
	if reflect.ValueOf(val).Kind() == reflect.Struct {
		val, err = treeFromStruct(val)
		if err != nil {
			return
		}
	}

	t.tree.Set(key, "", false, val)

	return
}

// migrate applies migrations to a tree.
//
// It returns whether migrations were applied.
func (t *Tree) migrate(
	migrations []MigrateHandler,
	versionKey string,
) (bool, error) {
	version, ok := t.GetDefault(versionKey, int64(0)).(int64)
	if !ok {
		return false, errors.WithStack(ErrInvalidVersion)
	}

	if version > int64(len(migrations)) {
		return false, errors.WithStack(ErrOutdatedExec)
	}

	migrations = migrations[version:]

	for _, m := range migrations {
		if err := m(t); err != nil {
			return false, errors.Wrap(err, fmt.Sprintf("migration %d", version))
		}

		version++

		if err := t.Set(versionKey, version); err != nil {
			return false, err
		}
	}

	return len(migrations) > 0, nil
}

// treeFromStruct creates a TOML tree from a struct.
func treeFromStruct(s interface{}) (*toml.Tree, error) {
	b := bytes.NewBuffer(nil)
	enc := toml.NewEncoder(b)
	enc.QuoteMapKeys(true)

	if err := enc.Encode(s); err != nil {
		return nil, errors.WithStack(err)
	}

	tree, err := toml.LoadBytes(b.Bytes())
	return tree, errors.WithStack(err)
}
