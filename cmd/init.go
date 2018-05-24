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
	"github.com/stratumn/alice/core/service/bootstrap"
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
		if initPrivateWithCoordinator {
			configurePrivateWithCoordinatorMode(configSet)
		}

		if err := core.InitConfig(configSet, coreCfgFilename()); err != nil {
			fmt.Fprintf(os.Stderr, "Could not save the core configuration file: %s.\n", err)
			os.Exit(1)
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

func configurePrivateWithCoordinatorMode(configSet cfg.Set) {
	swarmConfig, ok := configSet[swarm.ServiceID].Config().(swarm.Config)
	if !ok {
		fmt.Fprintln(os.Stderr, "Invalid or missing swarm configuration.")
		os.Exit(1)
	}

	bootstrapConfig, ok := configSet[bootstrap.ServiceID].Config().(bootstrap.Config)
	if !ok {
		fmt.Fprintln(os.Stderr, "Invalid or missing bootstrap configuration.")
		os.Exit(1)
	}

	swarmConfig.ProtectionMode = swarm.PrivateWithCoordinatorMode
	bootstrapConfig.Addresses = nil
	bootstrapConfig.MinPeerThreshold = 0

	if err := configSet[swarm.ServiceID].SetConfig(swarmConfig); err != nil {
		fmt.Fprintf(os.Stderr, "Could not set swarm configuration: %s.\n", err)
		os.Exit(1)
	}

	if err := configSet[bootstrap.ServiceID].SetConfig(bootstrapConfig); err != nil {
		fmt.Fprintf(os.Stderr, "Could not set bootstrap configuration: %s.\n", err)
		os.Exit(1)
	}
}

func init() {
	RootCmd.AddCommand(initCmd)

	initCmd.Flags().BoolVar(
		&initPrivateWithCoordinator,
		"private-with-coordinator",
		false,
		"initialize a private network using a coordinator node",
	)
}
