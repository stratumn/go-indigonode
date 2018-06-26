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
	"github.com/stratumn/go-indigonode/core/cfg"
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
	// Set contacts default DB
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
	// Set coin default DB
	func(tree *cfg.Tree) error {
		cwd, err := os.Getwd()
		if err != nil {
			return errors.WithStack(err)
		}

		folder, err := filepath.Abs(filepath.Join(cwd, "data", "coin", "db"))
		if err != nil {
			return errors.WithStack(err)
		}

		return tree.Set("coin.db_path", folder)
	},
	// Set chat history default DB
	func(tree *cfg.Tree) error {
		cwd, err := os.Getwd()
		if err != nil {
			return errors.WithStack(err)
		}

		filename, err := filepath.Abs(filepath.Join(cwd, "data", "chat-history.db"))
		if err != nil {
			return err
		}

		return tree.Set("chat.history_db_path", filename)
	},
	// Set storage default paths
	func(tree *cfg.Tree) error {
		cwd, err := os.Getwd()
		if err != nil {
			return errors.WithStack(err)
		}

		localStoragePath, err := filepath.Abs(filepath.Join(cwd, "data", "storage", "files"))
		if err != nil {
			return errors.WithStack(err)
		}

		storageDB, err := filepath.Abs(filepath.Join(cwd, "data", "storage", "db"))
		if err != nil {
			return errors.WithStack(err)
		}

		if err := tree.Set("storage.local_storage", localStoragePath); err != nil {
			return err
		}

		return tree.Set("storage.db_path", storageDB)
	},
	// Set public bootstrap seeds
	func(tree *cfg.Tree) error {
		seeds := append(
			tree.Get("bootstrap.addresses").([]interface{}),
			"/ip4/138.197.77.223/tcp/8903/ipfs/Qmc1QbSba7RtPgxEw4NqXNeDpB5CpCTwv9dvdZRdTkche1",
			"/ip4/165.227.14.175/tcp/8903/ipfs/QmQVdocY8ZbYxrKRSrff2Vxmm27Mhu6DgXyWXQwmuz1b6P",
			"/ip6/2a03:b0c0:1:d0::c9:9001/tcp/8903/ipfs/QmQJib6mnEMgdCe3bGH1YP7JswHbQQejyNucvW9BjFqmWr",
			"/ip6/2400:6180:0:d0::b1:b001/tcp/8903/ipfs/Qmc1rLFp5stHrjtq4duFg6KakBcDCpB3bTjjMZVSAdnHLj",
		)
		return tree.Set("bootstrap.addresses", seeds)
	},
	// New migrations should be added here for backwards-compatibility.
	// Existing migrations should not be edited as they won't be re-run.
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
