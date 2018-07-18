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

	"github.com/spf13/cobra"
)

// removeCmd represents the remove command.
var removeCmd = &cobra.Command{
	Use:   "remove",
	Short: "Unregister the daemon service",
	Run: func(cmd *cobra.Command, args []string) {
		status, err := newDaemon().Remove()
		fail(err)

		fmt.Println(status)
	},
}

func init() {
	daemonCmd.AddCommand(removeCmd)
}
