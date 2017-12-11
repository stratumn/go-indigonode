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

package cli_test

import (
	"bytes"
	"context"
	"io"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/pkg/errors"
	"github.com/stratumn/alice/cli"
	"github.com/stratumn/alice/cli/mockcli"
	"github.com/stratumn/alice/script"
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
	if err != nil {
		t.Fatalf("%s: parser error: %s", e.Command, err)
	}

	if list == nil {
		return nil, nil
	}

	var val script.SExp

	closure := script.NewClosure(script.ClosureOptResolver(cli.Resolver))

	itr := script.NewInterpreter(
		script.InterpreterOptVarPrefix("$"),
		script.InterpreterOptClosure(closure),
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
		if _, ok := err.(*cli.UseError); !ok {
			t.Errorf("%s: error = %v want %v", e.Command, err, e.Err)
		}
	case err != e.Err:
		t.Errorf("%s: error = %v want %v", e.Command, err, e.Err)
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

	if got != e.Want {
		t.Errorf("%s =>\n%s\nwant\n\n%s", e.Command, got, e.Want)
	}
}
