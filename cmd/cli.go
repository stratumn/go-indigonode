// Copyright © 2017 Stratumn SAS
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
	"context"
	"fmt"
	"os"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/stratumn/alice/cli"
	"google.golang.org/grpc"
)

var (
	cliCommand = ""
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

		if cliCommand != "" {
			// Execute command line argument.
			if err := c.Connect(ctx, ""); err != nil {
				fail(err)
			}

			// We don't use Exec so we can handle the errors here.
			if err := c.Eval(ctx, cliCommand); err != nil {
				// If it is a usage error, print the usage
				// message.
				cause := errors.Cause(err)
				stack := cli.StackTrace(err)

				if desc := grpc.ErrorDesc(cause); desc != "" {
					fmt.Fprintf(os.Stderr, "Error: %s.\n", desc)
				} else {
					fmt.Fprintf(os.Stderr, "Error: %s.\n", cause)
				}

				if c.Config().EnableDebugOutput && len(stack) > 0 {
					fmt.Fprintf(os.Stderr, "%+v\n", stack)
				}

				if userr, ok := cause.(*cli.UseError); ok {
					fmt.Fprintln(os.Stderr, "\n"+userr.Use())
				}

				os.Exit(1)
			}

			return
		}

		// Launch the shell.
		c.Run(ctx)
	},
}

func init() {
	RootCmd.AddCommand(cliCmd)

	cliCmd.Flags().StringVarP(&cliCommand, "command", "c", "", "execute command and exit")
}
