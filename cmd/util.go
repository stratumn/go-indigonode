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

package cmd

import (
	"fmt"
	"os"

	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"github.com/stratumn/go-node/cli"
	"github.com/stratumn/go-node/core"
	"github.com/stratumn/go-node/core/cfg"
)

// fail prints an error and exits if the error is not nil.
func fail(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s.\n", err)
		os.Exit(1)
	}
}

// coreCfgFilename returns the filename of the core config file.
func coreCfgFilename() string {
	return viper.GetString("core-config")
}

// cliCfgFilename returns the filename of the CLI config file.
func cliCfgFilename() string {
	return viper.GetString("cli-config")
}

// requireCoreConfigSet loads the core's configuration file and exits on failure.
func requireCoreConfigSet() cfg.Set {
	set := core.NewConfigurableSet(services)

	if err := core.LoadConfig(set, coreCfgFilename()); err != nil {
		fmt.Fprintf(os.Stderr, "Could not load the core configuration file %q: %s.\n", coreCfgFilename(), err)

		if os.IsNotExist(errors.Cause(err)) {
			fmt.Fprintln(os.Stderr, "You can create one using `stratumn-node init`.")
		}

		os.Exit(1)
	}

	return set
}

// requireCLIConfig loads the CLI's configuration file and exits on failure.
func requireCLIConfigSet() cfg.Set {
	set := cli.NewConfigurableSet()

	if err := cli.LoadConfig(set, cliCfgFilename()); err != nil {
		fmt.Fprintf(os.Stderr, "Could not load the command line interface configuration file %q: %s.\n", cliCfgFilename(), err)

		if os.IsNotExist(errors.Cause(err)) {
			fmt.Fprintln(os.Stderr, "You can create one using `stratumn-node init`.")
		}

		os.Exit(1)
	}

	return set
}
