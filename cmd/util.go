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

package cmd

import (
	"fmt"
	"os"

	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"github.com/stratumn/go-indigonode/cli"
	"github.com/stratumn/go-indigonode/core"
	"github.com/stratumn/go-indigonode/core/cfg"
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
			fmt.Fprintln(os.Stderr, "You can create one using `indigo-node init`.")
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
			fmt.Fprintln(os.Stderr, "You can create one using `indigo-node init`.")
		}

		os.Exit(1)
	}

	return set
}
