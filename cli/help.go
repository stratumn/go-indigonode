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

package cli

import (
	"context"
	"fmt"
	"strings"
	"text/tabwriter"

	"github.com/pkg/errors"
	"github.com/spf13/pflag"
)

// Help is a command that lists all the available commands or displays help for
// specific command.
var Help = BasicCmdWrapper{BasicCmd{
	Name:  "help",
	Use:   "help [Command]",
	Short: "Get help on commands",
	Exec:  helpExec,
}}

func helpExec(ctx context.Context, cli CLI, args []string, flags *pflag.FlagSet) error {
	if len(args) > 1 {
		return NewUseError("unexpected argument(s): " + strings.Join(args[1:], " "))
	}

	c := cli.Console()

	if len(args) > 0 {
		// Show help for a specific command.
		cmd := args[0]
		for _, v := range cli.Commands() {
			if v.Name() == cmd {
				c.Print(v.Long())
				return nil
			}
		}

		return errors.WithStack(ErrCmdNotFound)
	}

	// List all the available commands using a tab writer.
	w := new(tabwriter.Writer)
	w.Init(c, 0, 8, 2, ' ', 0)

	for _, v := range cli.Commands() {
		fmt.Fprintln(w, v.Use()+"\t"+v.Short())
	}

	return errors.WithStack(w.Flush())
}
