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

	"github.com/stratumn/alice/cli/script"
)

// Cons is a command that constructs a cell.
var Cons = BasicCmdWrapper{BasicCmd{
	Name:     "cons",
	Use:      "cons <Car> <Cdr>",
	Short:    "Construct a cons cell",
	ExecSExp: consExec,
}}

func consExec(
	ctx context.Context,
	cli CLI,
	closure *script.Closure,
	call script.CallHandler,
	args script.SCell,
	meta script.Meta,
) (script.SExp, error) {
	// Get:
	//
	//	1. car (the cell car)

	if args == nil {
		return nil, NewUseError("missing car")
	}

	car := args.Car()

	//	2. cadr (the cell cdr)
	cdr := args.Cdr()
	if cdr == nil {
		return nil, NewUseError("missing cdr")
	}

	cdrCell, ok := cdr.CellVal()
	if !ok {
		return nil, NewUseError("invalid cdr")
	}

	cadr := cdrCell.Car()

	// Evaluate them.
	carVal, err := script.Eval(closure.Resolve, call, car)
	if err != nil {
		return nil, err
	}

	cadrVal, err := script.Eval(closure.Resolve, call, cadr)
	if err != nil {
		return nil, err
	}

	// Construct the cell.
	return script.Cons(carVal, cadrVal, script.Meta{
		Line:   meta.Line,
		Offset: meta.Offset,
	}), nil
}
