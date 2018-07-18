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

package script

import (
	"github.com/pkg/errors"
)

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
