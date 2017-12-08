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
	if ctx.Args.IsNil() {
		return nil, errors.New("missing condition expression")
	}

	car := ctx.Args.Car()

	//	2. cadr (then)
	cdr := ctx.Args.Cdr()
	if cdr.IsNil() {
		return nil, errors.New("missing then expression")
	}

	cadr := cdr.Car()
	otherwise := Nil().(SExp)

	//	3. optional else symbol
	if cddr := cdr.Cdr(); !cddr.IsNil() {
		caddr := cddr.Car()

		if sym, ok := caddr.SymbolVal(); ok && sym == ElseSymbol {
			// The else token is present, so an else expression is
			// expected.
			cdddr := cddr.Cdr()
			if cdddr.IsNil() {
				return nil, errors.New("missing else expression")
			}

			// Make sure there isn't an extra expression.
			if cddddr := cdddr.Cdr(); !cddddr.IsNil() {
				return nil, Error("unexpected expression", cddddr.Meta(), "")
			}

			otherwise = cdddr.Car()
		} else {
			// Make sure there isn't an extra expression.
			if cdddr := cddr.Cdr(); !cdddr.IsNil() {
				return nil, Error("unexpected expression", cdddr.Meta(), "")
			}

			otherwise = caddr
		}
	}

	// Eval car (condition).
	cond, err := ctx.Eval(ctx, car, false)
	if err != nil {
		return nil, err
	}

	// Make sure it's a boolean.
	positive, ok := cond.BoolVal()
	if !ok {
		return nil, Error("not a boolean", cond.Meta(), cond.String())
	}

	// Evaluate cadr or otherwise based on condition.
	if positive == !unless {
		val, err := ctx.EvalBody(ctx, cadr, ctx.IsTail)
		if err != nil {
			return nil, err
		}

		return val, nil
	}

	val, err := ctx.EvalBody(ctx, otherwise, ctx.IsTail)
	if err != nil {
		return nil, err
	}

	return val, nil
}
