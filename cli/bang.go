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
	"context"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/pkg/errors"
	"github.com/stratumn/alice/cli/script"
)

// Bang is a command that executes external commands.
type Bang struct{}

// Name returns the name of the command (used by `help command`
// to find the command).
func (Bang) Name() string {
	return "!"
}

// Short returns a short description of the command.
func (Bang) Short() string {
	return "Execute and output external commands"
}

// Long returns a long description of the command.
func (Bang) Long() string {
	return `Execute and output external commands

Usage:
  ! <Command> [Args] [Input]

The first expression is the name of the executable, the optional second
expression are is a list of command arguments, and the optional third expression will
be written to the input stream of the command.

Examples:
  ! ls '-l -a'
  ! wc () (echo hello world !)
  ! less () (manager-list)
  ! grep manager (manager-list)
`
}

// Use returns a short string showing how to use the command.
func (Bang) Use() string {
	return "! <Command> [Args] [Input]"
}

// LongUse returns a long string showing how to use the command.
func (Bang) LongUse() string {
	return "! <Command> [Args] [Input]"
}

// Suggest gives a chance for the command to add auto-complete
// suggestions for the current content.
func (Bang) Suggest(Content) []Suggest {
	return nil
}

// Match returns whether the command can execute against the given
// command name.
func (Bang) Match(name string) bool {
	return strings.HasPrefix(name, "!")
}

// Exec executes the given S-Expression.
//
func (Bang) Exec(
	ctx context.Context,
	c CLI,
	closure *script.Closure,
	call script.CallHandler,
	args script.SCell,
	meta script.Meta,
) (script.SExp, error) {
	// Evaluate all the arguments to strings.
	argv, err := script.EvalListToStrings(closure.Resolve, call, args)
	if err != nil {
		return nil, err
	}

	// Check number of arguments.
	argc := len(argv)

	if argc < 1 {
		return nil, NewUseError("missing executable name")
	}

	if argc > 3 {
		return nil, NewUseError("unexpected expression")
	}

	// Executable name is the first argument.
	execName := argv[0]

	// Find executable arguments.
	var execArgs []string

	if argc > 1 {
		execArgs = strings.Split(argv[1], " ")

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
		stdin = bytes.NewBuffer([]byte(argv[2]))
	}

	// Use a buffer to capture executable output.
	buf := bytes.NewBuffer(nil)

	// Set up the command.
	cmd := exec.CommandContext(ctx, execName, execArgs...)
	cmd.Stdout = buf
	cmd.Stderr = os.Stderr
	cmd.Stdin = stdin

	// Start the process.
	err = cmd.Start()
	if err != nil {
		return nil, errors.WithStack(err)
	}

	// Wait for it to complete async.
	ch := make(chan error, 1)
	go func() {
		ch <- cmd.Wait()
	}()

	if err := <-ch; err != nil {
		return nil, errors.WithStack(err)
	}

	return script.String(buf.String(), meta), nil
}
