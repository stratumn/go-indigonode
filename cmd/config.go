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
	"strings"

	"github.com/spf13/cobra"
	"github.com/stratumn/alice/core/cfg"
)

const (
	keySeparator = "."
)

var setConfig = &cobra.Command{
	Use:   "set",
	Short: "Edit values from configuration files",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		// We need to check the first element of the key
		// to know which configuration file should be edited.
		key := args[0]
		svcKey := strings.Split(key, keySeparator)[0]
		value := args[1]

		var config cfg.Set
		var filename string
		if svcKey == "cli" {
			config = requireCLIConfigSet()
			filename = cliCfgFilename()
		} else {
			config = requireCoreConfigSet()
			filename = coreCfgFilename()
		}

		if err := config.Set(key, value); err != nil {
			fail(err)
		}
		if err := config.Save(filename, 0600, true); err != nil {
			fail(err)
		}
	},
}

var getConfig = &cobra.Command{
	Use:   "get",
	Short: "Get values from configuration files",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		key := args[0]
		svcKey := strings.Split(key, keySeparator)[0]

		var config cfg.Set
		if svcKey == "cli" {
			config = requireCLIConfigSet()
		} else {
			config = requireCoreConfigSet()
		}
		val, err := config.Get(key)
		if err != nil {
			fail(err)
		}

		fmt.Printf("%v\n", val)
	},
}

func init() {
	configCmd := &cobra.Command{Use: "config", Short: "Manage alice configurations"}
	configCmd.AddCommand(setConfig)
	configCmd.AddCommand(getConfig)
	RootCmd.AddCommand(configCmd)
}
