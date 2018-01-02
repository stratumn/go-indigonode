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
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stratumn/alice/cli"
	"github.com/stratumn/alice/core"
)

var (
	coreCfgFilename = "alice.core.toml"
	cliCfgFilename  = "alice.cli.toml"
)

// services are the services that will be used by core.
var services = core.BuiltinServices()

// RootCmd represents the base command when called without any subcommands.
var RootCmd = &cobra.Command{
	Use:   "alice",
	Short: "Alice is the mother of all blockchains",
}

// Execute adds all child commands to the root command and sets flags
// appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	RootCmd.PersistentFlags().StringVar(&coreCfgFilename, "core-config", coreCfgFilename, "core configuration file")
	RootCmd.PersistentFlags().StringVar(&cliCfgFilename, "cli-config", cliCfgFilename, "command line interface configuration file")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	viper.AutomaticEnv() // read in environment variables that match
}

// requireCoreConfigSet loads the core's configuration file and exits on failure.
func requireCoreConfigSet() core.ConfigurableSet {
	set := core.NewConfigurableSet(services)

	if err := core.LoadConfig(set, coreCfgFilename); err != nil {
		fmt.Fprintf(os.Stderr, "Could not load the core configuration file %q: %s.\n", coreCfgFilename, err)

		if os.IsNotExist(errors.Cause(err)) {
			fmt.Fprintln(os.Stderr, "You can create one using `alice init`.")
		}

		os.Exit(1)
	}

	return set
}

// requireCLIConfig loads the CLI's configuration file and exits on failure.
func requireCLIConfigSet() cli.ConfigurableSet {
	set := cli.NewConfigurableSet()

	if err := cli.LoadConfig(set, cliCfgFilename); err != nil {
		fmt.Fprintf(os.Stderr, "Could not the load command line interface configuration file %q: %s.\n", cliCfgFilename, err)

		if os.IsNotExist(errors.Cause(err)) {
			fmt.Fprintln(os.Stderr, "You can create one using `alice init`.")
		}

		os.Exit(1)
	}

	return set
}
