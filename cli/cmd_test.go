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

package cli_test

import (
	"bytes"
	"context"
	"io"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/pkg/errors"
	"github.com/stratumn/go-node/cli"
	"github.com/stratumn/go-node/cli/mockcli"
	"github.com/stratumn/go-node/script"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	// ErrAny means any error type is expected.
	ErrAny = errors.New("any error")

	// ErrUse means a usage error is expected.
	ErrUse = errors.New("usage error")
)

// ExecTest helps testing commands.
type ExecTest struct {
	// Command is the command to run.
	Command string

	// Want is the expected output.
	Want string

	// Err is the expected error.
	Err error

	// If the command expects anything other than the console from the CLI,
	// add expectations in this function.
	Expect func(*mockcli.MockCLI)
}

// Exec executes the test command.
func (e ExecTest) Exec(t *testing.T, w io.Writer, cmd cli.Cmd) (script.SExp, error) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	c := mockcli.NewMockCLI(ctrl)
	cons := cli.NewConsole(w, false)

	c.EXPECT().Console().Return(cons).AnyTimes()
	if e.Expect != nil {
		e.Expect(c)
	}

	parser := script.NewParser(script.NewScanner())
	list, err := parser.Parse(e.Command)
	require.NoError(t, err, "parser error")

	if list == nil {
		return nil, nil
	}

	var val script.SExp

	closure := script.NewClosure(script.ClosureOptResolver(cli.Resolver))

	itr := script.NewInterpreter(
		script.InterpreterOptClosure(closure),
		script.InterpreterOptBuiltinLibs,
		script.InterpreterOptErrorHandler(func(error) {}),
		script.InterpreterOptValueHandler(func(v script.SExp) {
			val = v
		}),
	)

	// Find command name.
	name := list.Car().Car().MustSymbolVal()
	itr.AddFuncHandler(name, func(ctx *script.InterpreterContext) (script.SExp, error) {
		return cmd.Exec(ctx, c)
	})

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	err = itr.EvalInput(ctx, e.Command)

	return val, err
}

// Test runs the test against the command.
func (e ExecTest) Test(t *testing.T, cmd cli.Cmd) {
	buf := bytes.NewBuffer(nil)
	val, err := e.Exec(t, buf, cmd)
	err = errors.Cause(err)

	switch {
	case e.Err == ErrAny && err != nil:
		// Pass.
	case e.Err == ErrUse:
		_, ok := err.(*cli.UseError)
		assert.True(t, ok, "err.(*cli.UseError)")
	case err != e.Err:
		assert.Fail(t, "err != e.Err")
	}

	got := buf.String()

	if val != nil {
		if val.UnderlyingType() == script.SExpString {
			// So we don't get a quoted string.
			got += val.MustStringVal()
		} else {
			got += val.String()
		}
	}

	assert.Equal(t, e.Want, got)
}
