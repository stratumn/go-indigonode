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
	"strings"

	"github.com/spf13/cobra"
	"github.com/stratumn/go-indigonode/core/cfg"
)

var backup bool

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
		if err := config.Save(filename, 0600, cfg.ConfigSaveOpts{
			Overwrite: true,
			Backup:    backup,
		}); err != nil {
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
	configCmd := &cobra.Command{Use: "config", Short: "Manage Indigo Node configurations"}
	setConfig.Flags().BoolVar(&backup, "backup", true, "save configuration backup when editing")
	configCmd.AddCommand(setConfig)
	configCmd.AddCommand(getConfig)
	RootCmd.AddCommand(configCmd)
}
