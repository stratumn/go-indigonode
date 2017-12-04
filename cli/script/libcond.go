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
		return nil, ctx.Error("missing condition expression")
	}

	car := ctx.Args.Car()

	//	2. cadr (then)
	cdr := ctx.Args.Cdr()
	if cdr == nil {
		return nil, ctx.Error("missing then expression")
	}

	cdrCell, ok := cdr.CellVal()
	if !ok {
		return nil, ctx.Error("invalid then expression")
	}

	cadr := cdrCell.Car()

	//	3. caddr, optional (else)
	cddr := cdrCell.Cdr()
	var caddr SExp

	if cddr != nil {
		cddrCell, ok := cddr.CellVal()
		if !ok {
			return nil, ctx.Error("invalid else expression")
		}
		caddr = cddrCell.Car()

		// Make sure there isn't an extra expression.
		if cddrCell.Cdr() != nil {
			return nil, ctx.Error("unexpected expression")
		}
	}

	// Eval car (condition).
	cond, err := ctx.Eval(ctx.Ctx, ctx.Closure, car)
	if err != nil {
		return nil, ctx.WrapError(err)
	}

	// Evaluate cadr or caddr based on condition.
	if (cond == nil) == unless {
		val, err := ctx.EvalBody(ctx.Ctx, ctx.Closure, cadr)
		if err != nil {
			return nil, ctx.WrapError(err)
		}

		return val, nil
	}

	if caddr != nil {
		val, err := ctx.EvalBody(ctx.Ctx, ctx.Closure, caddr)
		if err != nil {
			return nil, ctx.WrapError(err)
		}

		return val, nil
	}

	return nil, nil
}
