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
  !<Command>

Bang executes the external command immediatelly following the exclamation point.

Example:
  !ls -l
`
}

// Use returns a short string showing how to use the command.
func (Bang) Use() string {
	return "!<Command>"
}

// LongUse returns a long string showing how to use the command.
func (Bang) LongUse() string {
	return "!<Command>"
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
	resolve script.SExpResolver,
	eval script.SExpEvaluator,
	exp *script.SExp,
) error {
	name := exp.Str[1:]
	if len(name) == 0 {
		return NewUseError("command is blank")
	}

	args, err := exp.Cdr.ResolveEvalEach(resolve, eval)
	if err != nil {
		return err
	}

	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Stdout = w
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

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
