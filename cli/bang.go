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
	return "Execute external commands"
}

// Long returns a long description of the command.
func (Bang) Long() string {
	return `Execute external commands

Usage:
  !<Command> args...

Bang executes the external command immediately following the exclamation point.

If a name doesn't immediately follow the exclamation point, the first element 
is the name of the command, the optional second element are arguments for the
command, and the optional third element will be written to the input stream of 
the command.

Examples:
  !ls -l -a
  ! wc () (echo hello world !) ; use a pipe, notice the space before the command
  ! less () (manager-list)
  ! grep "manager" (manager-list)
`
}

// Use returns a short string showing how to use the command.
func (Bang) Use() string {
	return "!<Command> args..."
}

// LongUse returns a long string showing how to use the command.
func (Bang) LongUse() string {
	return "!<Command> args..."
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
	w io.Writer,
	closure *script.Closure,
	eval script.SExpEvaluator,
	exp *script.SExp,
) error {
	var stdin io.Reader

	args, err := exp.Cdr.ResolveEvalEach(closure.Resolve, eval)
	if err != nil {
		return err
	}

	name := exp.Str[1:]

	if len(name) < 1 {
		argc := len(args)
		if argc < 1 {
			return NewUseError("expected command name")
		}
		if argc > 3 {
			return NewUseError("expected at most three expressions")
		}

		name = args[0]

		if argc > 2 {
			stdin = bytes.NewBuffer([]byte(args[2]))
		}

		if argc > 1 {
			args = strings.Split(args[1], " ")
		} else {
			args = nil
		}
	}

	// Filter out empty arguments.
	tmp, args := args, args[:0]
	for _, a := range tmp {
		if a != "" {
			args = append(args, a)
		}
	}

	if stdin == nil {
		stdin = os.Stdin
	}

	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Stdout = w
	cmd.Stderr = os.Stderr
	cmd.Stdin = stdin

	err = cmd.Start()
	if err != nil {
		return errors.WithStack(err)
	}

	ch := make(chan error, 1)
	go func() {
		ch <- cmd.Wait()
	}()

	return errors.WithStack(<-ch)
}
