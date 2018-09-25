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
	"github.com/spf13/cobra"
	"github.com/takama/daemon"
)

// daemonCmd represents the daemon command.
var daemonCmd = &cobra.Command{
	Use:   "daemon",
	Short: "Subcommands to run Indigo Node as a daemon",
}

func init() {
	RootCmd.AddCommand(daemonCmd)
}

// newDaemon creates a new daemon service.
func newDaemon() daemon.Daemon {
	d, err := daemon.New("stratumn-node", "Indigo Node")
	fail(err)

	return d
}
