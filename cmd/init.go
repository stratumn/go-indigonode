// Copyright © 2017 Stratumn SAS
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
	"github.com/spf13/cobra"
	"github.com/stratumn/alice/cli"
	"github.com/stratumn/alice/core"
)

// initRecreate is whether to load and save the config file, adding any missing
// settings since it was last saved.
var initRecreate = false

// initCmd represents the init command.
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Create configuration file",
	Run: func(cmd *cobra.Command, args []string) {
		if err := core.InitConfig(core.NewConfigSet(), coreCfgFilename, initRecreate); err != nil {
			fmt.Fprintf(os.Stderr, "Could not save the core configuration file: %s.\n", err)
			cfgExists(err)
			os.Exit(1)
		}

		fmt.Printf("Created configuration file %q.\n", coreCfgFilename)
		fmt.Println("Keep this file private!!!")

		if err := cli.InitConfig(cli.NewConfigSet(), cliCfgFilename, initRecreate); err != nil {
			fmt.Fprintf(os.Stderr, "Could not save the command line interface configuration file: %s.\n", err)
			cfgExists(err)
			os.Exit(1)
		}

		fmt.Printf("Created configuration file %q.\n", cliCfgFilename)
		fmt.Println("Keep this file private!!!")
	},
}

func init() {
	RootCmd.AddCommand(initCmd)
	initCmd.Flags().BoolVar(&initRecreate, "recreate", false, "load and save configuration files, adding missing settings")
}

// cfgExists handle an error that occured because a configuration file already
// exists.
func cfgExists(err error) {
	if os.IsExist(errors.Cause(err)) {
		fmt.Println("You can load and save the configuration files using `alice init --recreate`.")
		fmt.Println("This will keep your existing settings but add any missing or new settings.")
	}
}
