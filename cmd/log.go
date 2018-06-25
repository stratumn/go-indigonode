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
