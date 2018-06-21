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

package cli

import (
	"fmt"
	"strings"
	"text/tabwriter"

	"github.com/pkg/errors"
)

// Help is a command that lists all the available commands or displays help for
// specific command.
var Help = BasicCmdWrapper{
	Cmd: BasicCmd{
		Name:  "help",
		Use:   "help [Command]",
		Short: "Get help on commands",
		Exec:  helpExec,
	},
}

func helpExec(ctx *BasicContext) error {
	argc := len(ctx.Args)
	if argc > 1 {
		return NewUseError("unexpected argument(s): " + strings.Join(ctx.Args[1:], " "))
	}

	if argc > 0 {
		// Show help for a specific command.
		cmd := ctx.Args[0]
		for _, v := range ctx.CLI.Commands() {
			if v.Name() == cmd {
				fmt.Fprintln(ctx.Writer, v.Long())
				return nil
			}
		}

		return errors.WithStack(ErrCmdNotFound)
	}

	// List all the available commands using a tab writer.
	tw := new(tabwriter.Writer)
	tw.Init(ctx.Writer, 0, 8, 2, ' ', 0)

	for _, v := range ctx.CLI.Commands() {
		fmt.Fprintln(tw, v.Use()+"\t"+v.Short())
	}

	return errors.WithStack(tw.Flush())
}
