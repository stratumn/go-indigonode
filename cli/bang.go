// Copyright © 2017  Stratumn SAS
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
	"bytes"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/pkg/errors"
)

// Bang is a command that executes external commands.
var Bang = BasicCmdWrapper{BasicCmd{
	Name:        "!",
	Use:         "! <Command> [args] [input]",
	Short:       "Executes external commands",
	ExecStrings: bangExec,
}}

func bangExec(ctx *StringsContext, cli CLI) error {
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
