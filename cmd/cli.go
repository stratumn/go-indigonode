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
	"context"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/stratumn/go-node/cli"

	"google.golang.org/grpc/status"
)

var (
	cliCommand = ""
	cliOpen    = ""
)

// cliCmd represents the cli command.
var cliCmd = &cobra.Command{
	Use:   "cli",
	Short: "Launch command line interface client",
	Run: func(cmd *cobra.Command, args []string) {
		config := requireCLIConfigSet().Configs()

		c, err := cli.New(config)
		fail(err)

		ctx := context.Background()

		if cliCommand != "" || cliOpen != "" {
			script := cliCommand

			if cliOpen != "" {
				b, err := ioutil.ReadFile(cliOpen)
				fail(err)
				script = string(b)
			}

			if err := c.Connect(ctx, ""); err != nil {
				fail(err)
			}

			// We don't use Run so we can handle the errors here.
			if err := c.Exec(ctx, script); err != nil {
				cause := errors.Cause(err)
				stack := cli.StackTrace(err)

				if s, ok := status.FromError(err); ok {
					fmt.Fprintf(os.Stderr, "Error: %s.\n", s.Message())
				} else {
					fmt.Fprintf(os.Stderr, "Error: %s.\n", cause)
				}

				if c.Config().EnableDebugOutput && len(stack) > 0 {
					fmt.Fprintf(os.Stderr, "%+v\n", stack)
				}

				// If it is a usage error, print the usage
				// message.
				if useError, ok := cause.(*cli.UseError); ok {
					fmt.Fprintln(os.Stderr, "\n"+useError.Use())
				}

				os.Exit(1)
			}

			return
		}

		c.Start(ctx)
	},
}

func init() {
	RootCmd.AddCommand(cliCmd)

	cliCmd.Flags().StringVarP(&cliCommand, "command", "c", "", "execute command and exit")
	cliCmd.Flags().StringVarP(&cliOpen, "open", "o", "", "execute file and exit")
}
