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
