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
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stratumn/go-indigonode/core"
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
	RootCmd.PersistentFlags().String("core-config", "alice.core.toml", "core configuration file")
	RootCmd.PersistentFlags().String("cli-config", "alice.cli.toml", "command line interface configuration file")

	err := viper.BindPFlags(RootCmd.PersistentFlags())
	if err != nil {
		panic(err)
	}
}

// initConfig reads in ENV variables if set.
func initConfig() {
	viper.SetEnvPrefix("alice")
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	viper.AutomaticEnv()
}
