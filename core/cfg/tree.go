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

package cfg

import (
	"bytes"
	"fmt"
	"reflect"

	toml "github.com/pelletier/go-toml"
	"github.com/pkg/errors"
)

var (
	// ErrInvalidGroup is returned when a configuration file contains
	// an invalid service group.
	ErrInvalidGroup = errors.New("the service group is invalid")
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

	return &Tree{tree: t}, nil
}

// Get returns the value of a key.
//
// The key can be a path such as "core.boot_service".
func (t *Tree) Get(key string) interface{} {
	val := t.tree.Get(key)

	switch v := val.(type) {
	case (*toml.Tree):
		return &Tree{tree: v}
	case ([]*toml.Tree):
		trees := make([]*Tree, len(v))
		for i, child := range v {
			trees[i] = &Tree{tree: child}
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
		t.tree.Set(key, v.tree)
		return
	case []*Tree:
		trees := make([]*toml.Tree, len(v))
		for i, child := range v {
			trees[i] = child.tree
		}
		t.tree.Set(key, trees)
		return
	}

	// Struct must be converted to a tree before calling tree.Set().
	if reflect.ValueOf(val).Kind() == reflect.Struct {
		val, err = treeFromStruct(val)
		if err != nil {
			return
		}
	}

	t.tree.Set(key, val)

	return
}

// AddGroup adds a group if it doesn't exist yet.
func (t *Tree) AddGroup(id, name, desc string) error {
	groups, ok := t.Get("core.service_groups").([]*Tree)
	if !ok {
		groups = []*Tree{}
	}

	for _, group := range groups {
		if group.Get("id") == id {
			return nil
		}
	}

	group, err := TreeFromMap(map[string]interface{}{
		"id":          id,
		"name":        name,
		"description": desc,
	})
	if err != nil {
		return err
	}

	return t.Set("core.service_groups", append(groups, group))
}

// AddServiceToGroup adds a service to an existing group if it isn't already
// part of the group.
//
// If the group is not found, it prints a warning and doesn't return an error.
// The user could have intentionally removed the group.
func (t *Tree) AddServiceToGroup(service, group string) error {
	groups, ok := t.Get("core.service_groups").([]*Tree)
	if !ok {
		fmt.Printf(
			"Warning: could not add service %s to group %s because all groups were removed.",
			service, group,
		)
		return nil
	}

	for _, g := range groups {
		if g.Get("id") == group {
			var services []interface{}

			if v := g.Get("services"); v != nil {
				services, ok = v.([]interface{})
				if !ok {
					return errors.Wrap(ErrInvalidGroup, group)
				}
			}

			for _, s := range services {
				sid, ok := s.(string)
				if !ok {
					return errors.Wrap(ErrInvalidGroup, group)
				}

				if sid == service {
					return nil
				}
			}

			return g.Set("services", append(services, service))
		}
	}

	fmt.Printf(
		"Warning: could not add service %s to group %s because the group was removed.",
		service, group,
	)

	return nil
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
