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
	"fmt"
	"io"

	"github.com/stratumn/alice/cli/script"
)

// Cdr is a command that prints the tail of a list.
var Cdr = BasicCmdWrapper{BasicCmd{
	Name:     "cdr",
	Use:      "cdr (quote (a b c))",
	Short:    "Print the tail a list",
	ExecSExp: cdr,
}}

func cdr(
	ctx context.Context,
	cli CLI, w io.Writer,
	closure *script.Closure,
	eval script.Evaluator,
	exp *script.SExp,
) error {
	if exp.Cdr == nil || exp.Cdr.Cdr != nil {
		return NewUseError("expected a single expression")
	}

	v, err := exp.Cdr.ResolveEval(closure.Resolve, eval)
	if err != nil {
		return err
	}

	scanner := script.NewScanner(script.OptErrorHandler(cli.PrintError))
	parser := script.NewParser(scanner)

	list, err := parser.List(v)
	if err != nil {
		return err
	}

	// TODO: should this be an error?
	if list == nil {
		fmt.Fprintln(w)
		return nil
	}

	fmt.Fprintln(w, list.Cdr.String())

	return nil
}
