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
  !<Command> args...

Bang executes the external command immediately following the exclamation point.

If there isn't space between the exclamation point and the command name, the
remaining expressions will be evaluated as the command's arguments.

If there is whitespace before the name, the first expression is the name of the
command, the optional second expression are the command's arguments, and the
optional third expression will be written to the input stream of the command.

Examples:
  !ls -l -a
  ! wc () (echo hello world !) ; use a pipe, notice the space before the command
  ! less () (manager-list)
  ! grep manager (manager-list)
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
	closure *script.Closure,
	call script.CallHandler,
	exp *script.SExp,
) (*script.SExp, error) {
	var stdin io.Reader

	vals, err := exp.Cdr.ResolveEvalEach(closure.Resolve, call)
	if err != nil {
		return nil, err
	}

	args := vals.Strings(false)
	name := exp.Str[1:]

	if len(name) < 1 {
		argc := len(args)
		if argc < 1 {
			return nil, NewUseError("expected command name")
		}
		if argc > 3 {
			return nil, NewUseError("expected at most three expressions")
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

	buf := bytes.NewBuffer(nil)

	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Stdout = buf
	cmd.Stderr = os.Stderr
	cmd.Stdin = stdin

	err = cmd.Start()
	if err != nil {
		return nil, errors.WithStack(err)
	}

	ch := make(chan error, 1)
	go func() {
		ch <- cmd.Wait()
	}()

	if err := <-ch; err != nil {
		return nil, errors.WithStack(err)
	}

	return &script.SExp{
		Type:   script.TypeStr,
		Str:    strings.TrimSuffix(buf.String(), "\n"),
		Line:   exp.Line,
		Offset: exp.Offset,
	}, nil
}
