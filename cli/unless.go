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
	"github.com/stratumn/alice/cli/script"
)

// Unless is a command that does the opposite of the If command.
var Unless = BasicCmdWrapper{BasicCmd{
	Name:     "unless",
	Use:      "unless <Condition> <Then> [Else]",
	Short:    "Opposite of the if command",
	ExecSExp: unlessExec,
}}

func unlessExec(ctx *ExecContext) (script.SExp, error) {
	// Get:
	//
	//	1. car (condition)
	if ctx.Args == nil {
		return nil, NewUseError("missing condition expression")
	}

	car := ctx.Args.Car()
	if car == nil {
		return nil, NewUseError("missing condition expression")
	}

	//	2. cadr (then)

	cdr := ctx.Args.Cdr()
	if cdr == nil {
		return nil, NewUseError("missing then expression")
	}

	cdrCell, ok := cdr.CellVal()
	if !ok {
		return nil, NewUseError("invalid then expression")
	}

	cadr := cdrCell.Car()
	if cadr == nil {
		return nil, NewUseError("missing then expression")
	}

	//	3. caddr, optional (else)

	cddr := cdrCell.Cdr()
	var caddr script.SExp

	if cddr != nil {
		cddrCell, ok := cddr.CellVal()
		if !ok {
			return nil, NewUseError("invalid else expression")
		}
		caddr = cddrCell.Car()

		// Make sure there isn't an extra expression.
		if cddrCell.Cdr() != nil {
			return nil, NewUseError("unexpected expression")
		}
	}

	// Eval car (condition).
	val, err := script.Eval(ctx.Closure.Resolve, ctx.Call, car)

	// Print error, but keep going.
	if err != nil {
		ctx.CLI.PrintError(err)
	}

	if val == nil {
		// If val is nil, eval cadr (then).
		return evalSExpBody(ctx.Closure.Resolve, ctx.Call, cadr)
	}

	if caddr != nil {
		// Otherwise eval caddr (else), if any.
		return evalSExpBody(ctx.Closure.Resolve, ctx.Call, caddr)
	}

	return nil, nil
}
