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
	"os"

	"github.com/spf13/cobra"
	"github.com/stratumn/go-indigonode/core"
)

var (
	// logLevel filters entry with the given level.
	logLevel string

	// logSystem filters entry from a log system.
	logSystem string

	// logFollow is whether to follow the log file.
	logFollow bool

	// logJSON is whether to output JSON.
	logJSON bool

	// logNoColor is whether to disable color output.
	logNoColor bool
)

// logCmd represents the log command.
var logCmd = &cobra.Command{
	Use:   "log <filename>",
	Short: "Pretty print log files",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			if err := RootCmd.Usage(); err != nil {
				fail(err)
			}

			os.Exit(1)
		}

		err := core.PrettyLog(args[0], logLevel, logSystem, logFollow, logJSON, !logNoColor)
		fail(err)
	},
}

func init() {
	RootCmd.AddCommand(logCmd)
	logCmd.Flags().StringVarP(&logLevel, "level", "l", "all", "only show entries with this level (info, error, all)")
	logCmd.Flags().StringVarP(&logSystem, "system", "s", "", "only show entries from this system")
	logCmd.Flags().BoolVarP(&logFollow, "follow", "f", false, "follow log file")
	logCmd.Flags().BoolVar(&logJSON, "json", false, "output line separated JSON")
	logCmd.Flags().BoolVar(&logNoColor, "no-color", false, "disable color output")
}
