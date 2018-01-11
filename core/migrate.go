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
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/stratumn/alice/core/cfg"
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
}
