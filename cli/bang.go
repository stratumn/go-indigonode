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
	"bytes"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/pkg/errors"
)

// Bang is a command that executes external commands.
var Bang = BasicCmdWrapper{
	Cmd: BasicCmd{
		Name:    "!",
		Use:     "! <Command> [args] [input]",
		Short:   "Executes external commands",
		Exec:    bangExec,
		NoFlags: true,
	},
}

func bangExec(ctx *BasicContext) error {
	// Check number of arguments.
	argc := len(ctx.Args)

	if argc < 1 {
		return NewUseError("missing executable name")
	}

	if argc > 3 {
		return NewUseError("unexpected expression")
	}

	// Executable name is the first argument.
	execName := ctx.Args[0]

	// Find executable arguments.
	var execArgs []string

	if argc > 1 {
		execArgs = strings.Split(ctx.Args[1], " ")

		// Filter out empty arguments.
		tmp := execArgs
		execArgs = execArgs[:0]

		for _, a := range tmp {
			if a != "" {
				execArgs = append(execArgs, a)
			}
		}
	}

	// Set executable input
	var stdin io.Reader = os.Stdin

	if argc > 2 {
		stdin = bytes.NewBuffer([]byte(ctx.Args[2]))
	}

	// Set up the command.
	cmd := exec.CommandContext(ctx.Ctx, execName, execArgs...)
	cmd.Stdout = ctx.Writer
	cmd.Stderr = os.Stderr
	cmd.Stdin = stdin

	// Start the process.
	err := cmd.Start()
	if err != nil {
		return errors.WithStack(err)
	}

	// Wait for it to complete async.
	ch := make(chan error, 1)
	go func() {
		ch <- cmd.Wait()
	}()

	if err := <-ch; err != nil {
		return errors.WithStack(err)
	}

	return nil
}
