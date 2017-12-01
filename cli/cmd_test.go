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
	"strings"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/pkg/errors"
	"github.com/stratumn/alice/cli"
	"github.com/stratumn/alice/cli/mockcli"
	"github.com/stratumn/alice/cli/script"
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

// Test runs the test against the command.
func (e ExecTest) Test(t *testing.T, cmd cli.Cmd) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	c := mockcli.NewMockCLI(ctrl)

	buf := bytes.NewBuffer(nil)
	cons := cli.NewConsole(buf, false)

	c.EXPECT().Console().Return(cons).AnyTimes()
	if e.Expect != nil {
		e.Expect(c)
	}

	var call script.CallHandler

	call = func(resolve script.ResolveHandler, exp *script.SExp) (*script.SExp, error) {
		if exp.Str == "quote" {
			if exp.Cdr != nil && exp.Cdr.Cdr != nil {
				return nil, cli.NewUseError("expected a single expression")
			}
			return exp.Cdr, nil
		}

		args, err := exp.Cdr.ResolveEvalEach(resolve, call)
		if err != nil {
			return nil, err
		}

		str := args.JoinCars(" ", false)

		switch exp.Str {
		case "echo":
			return &script.SExp{
				Type:   script.TypeStr,
				Str:    str,
				Line:   exp.Line,
				Offset: exp.Offset,
			}, nil
		case "title":
			return &script.SExp{
				Type:   script.TypeStr,
				Str:    strings.Title(str),
				Line:   exp.Line,
				Offset: exp.Offset,
			}, nil
		}

		return nil, script.ErrInvalidOperand
	}

	parser := script.NewParser(script.NewScanner())
	head, err := parser.Parse(e.Command)
	if err != nil {
		t.Fatalf("%s: parser error: %s", e.Command, err)
	}

	closure := script.NewClosure(script.OptResolver(cli.Resolver))

	var exp *script.SExp

	if head != nil {
		exp, err = cmd.Exec(ctx, c, closure, call, head.List)
		err = errors.Cause(err)
	}

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

	got := buf.String() + exp.CarString(false)

	if got != e.Want {
		t.Errorf("%s =>\n%s\nwant\n\n%s", e.Command, got, e.Want)
	}
}
