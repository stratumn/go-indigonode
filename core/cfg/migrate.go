// Copyright © 2017  Stratumn SAS
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
	"context"
	"fmt"
	logging "gx/ipfs/QmSpJByNKFX1sCsHBEp3R73FL4NF6FnQTEGyNAXHm2GS52/go-log"
	"os"
	"reflect"

	toml "github.com/pelletier/go-toml"
	"github.com/pkg/errors"
)

var (
	// ErrInvalidVersion is returned when the config version is invalid.
	ErrInvalidVersion = errors.New("config version is invalid")
)

// Tree is used by migrations to modify the configuration tree.
type Tree struct {
	tree *toml.Tree
}

// Get returns the value of a key.
//
// The key can be a path such as "core.boot_service".
func (t *Tree) Get(key string) interface{} {
	return t.tree.Get(key)
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

// MigrateHandler mutates a TOML tree to update it from one version to the
// next.
type MigrateHandler func(*Tree) error

// Migrate loads a TOML file and sets the configurations of a set of
// configurables, applying migrations if needed.
//
// The version key should point to the tree path that contains the int config
// file version.
//
// The migrations should be a slice of MigrateHandler that upgrade the
// configuration from the version corresponding to its index in the slice to
// the next version.
func Migrate(
	set Set,
	filename string,
	versionKey string,
	migrations []MigrateHandler,
	perms os.FileMode,
) error {
	ctx := logging.ContextWithLoggable(context.Background(), logging.Metadata{
		"filename":   filename,
		"versionKey": versionKey,
	})
	event := log.EventBegin(ctx, "migrate")
	defer event.Done()

	tree, err := toml.LoadFile(filename)
	if err != nil {
		event.SetError(err)
		return errors.WithStack(err)
	}

	t := &Tree{tree}

	version, ok := t.GetDefault(versionKey, int64(0)).(int64)
	if !ok {
		err := errors.WithStack(ErrInvalidVersion)
		event.SetError(err)
		return err
	}

	migrations = migrations[version:]

	for _, m := range migrations {
		if err := m(t); err != nil {
			event.SetError(err)
			return errors.Wrap(err, fmt.Sprintf("migration %d", version))
		}

		version++
		tree.Set(versionKey, "", false, version)
	}

	if err := setValuesFromTree(ctx, set, tree); err != nil {
		event.SetError(err)
		return err
	}

	// Save config file if there were migrations.
	if len(migrations) > 0 {
		return Save(set, filename, perms, true)
	}

	return nil
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
