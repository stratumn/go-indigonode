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

	"github.com/spf13/cobra"
	"github.com/stratumn/alice/cli"
	"github.com/stratumn/alice/core"
	"github.com/stratumn/alice/core/cfg"
	"github.com/stratumn/alice/core/service/swarm"
)

var (
	initPrivateWithCoordinator = false
)

// initCmd represents the init command.
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Create configuration file",
	Run: func(cmd *cobra.Command, args []string) {
		configSet := core.NewConfigurableSet(services)
		coreConfigFilename := coreCfgFilename()

		if err := core.InitConfig(configSet, coreConfigFilename); err != nil {
			fmt.Fprintf(os.Stderr, "Could not save the core configuration file: %s.\n", err)
			os.Exit(1)
		}

		if initPrivateWithCoordinator {
			configurePrivateCoordinatorMode(configSet, coreConfigFilename)
		}

		fmt.Printf("Created configuration file %q.\n", coreCfgFilename())
		fmt.Println("Keep this file private!!!")

		if err := cli.InitConfig(cli.NewConfigurableSet(), cliCfgFilename()); err != nil {
			fmt.Fprintf(os.Stderr, "Could not save the command line interface configuration file: %s.\n", err)
			os.Exit(1)
		}

		fmt.Printf("Created configuration file %q.\n", cliCfgFilename())
		fmt.Println("Keep this file private!!!")
	},
}

func configurePrivateCoordinatorMode(configSet cfg.Set, coreConfigFilename string) {
	if err := cfg.Update(
		configSet,
		coreConfigFilename,
		[]cfg.UpdateHandler{
			func(tree *cfg.Tree) error {
				if err := tree.Set("swarm.protection_mode", swarm.PrivateCoordinatorMode); err != nil {
					return err
				}
				if err := tree.Set("bootstrap.addresses", nil); err != nil {
					return err
				}
				if err := tree.Set("bootstrap.min_peer_threshold", int64(0)); err != nil {
					return err
				}

				return nil
			},
		},
		0600,
	); err != nil {
		fmt.Fprintf(os.Stderr, "Could not configure private network: %s.\n", err)
		os.Exit(1)
	}
}

func init() {
	RootCmd.AddCommand(initCmd)

	initCmd.Flags().BoolVar(
		&initPrivateWithCoordinator,
		"private-coordinator-mode",
		false,
		"initialize a private network using a coordinator node",
	)
}
