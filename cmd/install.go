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
	"path/filepath"

	"github.com/spf13/cobra"
)

// installCmd represents the install command.
var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Register the daemon service",
	Run: func(cmd *cobra.Command, args []string) {
		// Make sure the config file exists.
		requireCoreConfigSet()

		cfgPath, err := filepath.Abs(coreCfgFilename())
		fail(err)

		status, err := newDaemon().Install("up", "--core-config", cfgPath)
		fail(err)

		fmt.Println(status)
	},
}

func init() {
	daemonCmd.AddCommand(installCmd)
}
