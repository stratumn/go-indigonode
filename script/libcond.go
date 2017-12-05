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

package script

import "github.com/pkg/errors"

// LibCond contains functions for conditional statements.
var LibCond = map[string]InterpreterFuncHandler{
	"if":     LibCondIf,
	"unless": LibCondUnless,
}

// LibCondIf evaluates cadr if car evaluates to a non-nil value, otherwise it
// evaluates caddr (if present).
func LibCondIf(ctx *InterpreterContext) (SExp, error) {
	return libCondIf(ctx, false)
}

// LibCondUnless evaluates cadr if car evaluates to nil, otherwise it evaluates
// caddr (if present).
func LibCondUnless(ctx *InterpreterContext) (SExp, error) {
	return libCondIf(ctx, true)
}

func libCondIf(ctx *InterpreterContext, unless bool) (SExp, error) {
	// Get:
	//
	//	1. car (condition)
	if ctx.Args == nil {
		return nil, errors.New("missing condition expression")
	}

	car := ctx.Args.Car()

	//	2. cadr (then)
	cdr := ctx.Args.Cdr()
	if cdr == nil {
		return nil, errors.New("missing then expression")
	}

	cdrCell, ok := cdr.CellVal()
	if !ok {
		return nil, Error("invalid then expression", cdr.Meta(), "")
	}

	cadr := cdrCell.Car()

	//	3. caddr, optional (else)
	cddr := cdrCell.Cdr()
	var caddr SExp

	if cddr != nil {
		cddrCell, ok := cddr.CellVal()
		if !ok {
			return nil, Error("invalid else expression", cddr.Meta(), "")
		}
		caddr = cddrCell.Car()

		// Make sure there isn't an extra expression.
		if cdddr := cddrCell.Cdr(); cdddr != nil {
			return nil, Error("unexpected expression", cdddr.Meta(), "")
		}
	}

	// Eval car (condition).
	cond, err := ctx.Eval(ctx.Ctx, ctx.Closure, car)
	if err != nil {
		return nil, err
	}

	// Evaluate cadr or caddr based on condition.
	if (cond == nil) == unless {
		val, err := ctx.EvalBody(ctx.Ctx, ctx.Closure, cadr)
		if err != nil {
			return nil, err
		}

		return val, nil
	}

	if caddr != nil {
		val, err := ctx.EvalBody(ctx.Ctx, ctx.Closure, caddr)
		if err != nil {
			return nil, err
		}

		return val, nil
	}

	return nil, nil
}
