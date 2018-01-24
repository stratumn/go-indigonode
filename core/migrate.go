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

package core

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/stratumn/alice/core/cfg"
)

var (
	// ErrInvalidGroup is returned when a configuration file contains
	// an invalid service group.
	ErrInvalidGroup = errors.New("the service group is invalid")
)

var migrations = []cfg.MigrateHandler{
	func(tree *cfg.Tree) error {
		return tree.Set(ConfigVersionKey, 1)
	},
	func(tree *cfg.Tree) error {
		if err := tree.Set("core.boot_screen_host", "host"); err != nil {
			return err
		}

		return tree.Set("core.boot_screen_metrics", "metrics")
	},
	func(tree *cfg.Tree) error {
		return tree.Set("chat.host", "host")
	},
	func(tree *cfg.Tree) error {
		cwd, err := os.Getwd()
		if err != nil {
			return errors.WithStack(err)
		}

		filename, err := filepath.Abs(filepath.Join(cwd, "data", "contacts.toml"))
		if err != nil {
			return errors.WithStack(err)
		}

		return tree.Set("contacts.filename", filename)
	},
	func(tree *cfg.Tree) error {
		err := addGroup(tree, "util", "Utility Services", "Starts utility services.")
		if err != nil {
			return err
		}

		if err := addServiceToGroup(tree, "contacts", "util"); err != nil {
			return nil
		}

		return addServiceToGroup(tree, "util", "boot")
	},
	func(tree *cfg.Tree) error {
		err := tree.Set("chat.event", "event")
		if err != nil {
			return err
		}

		if err = tree.Set("event.write_timeout", "100ms"); err != nil {
			return err
		}

		return addServiceToGroup(tree, "event", "util")
	},
	func(tree *cfg.Tree) error {
		return tree.Set("coin.host", "host")
	},
}

// addGroup adds a group if it doesn't exist yet.
func addGroup(tree *cfg.Tree, id, name, desc string) error {
	groups, ok := tree.Get("core.service_groups").([]*cfg.Tree)
	if !ok {
		groups = []*cfg.Tree{}
	}

	for _, group := range groups {
		if group.Get("id") == id {
			return nil
		}
	}

	group, err := cfg.TreeFromMap(map[string]interface{}{
		"id":          id,
		"name":        name,
		"description": desc,
	})
	if err != nil {
		return err
	}

	return tree.Set("core.service_groups", append(groups, group))
}

// addServiceToGroup adds a services to an existing group if it isn't already
// part of the group.
//
// If the group is not found, it prints a warning and doesn't return an error.
// The user could have intentionally removed the group.
func addServiceToGroup(tree *cfg.Tree, service, group string) error {
	groups, ok := tree.Get("core.service_groups").([]*cfg.Tree)
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
