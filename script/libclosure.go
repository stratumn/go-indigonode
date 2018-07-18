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

import "github.com/pkg/errors"

// LibClosure contains functions to work with closures.
var LibClosure = map[string]InterpreterFuncHandler{
	"let": LibClosureLet,
}

// LibClosureLet binds a value to a symbol in the current closure.
func LibClosureLet(ctx *InterpreterContext) (SExp, error) {
	// Get:
	//
	//	1. car symbol value
	if ctx.Args.IsNil() {
		return nil, errors.New("missing symbol")
	}

	car := ctx.Args.Car()
	if car.IsNil() {
		return nil, errors.New("missing symbol")
	}

	carSymbol, ok := car.SymbolVal()
	if !ok {
		return nil, Error("not a symbol", car.Meta(), "")
	}

	//	2. cadr, optional (value)
	cdr := ctx.Args.Cdr()
	cadr := Nil().(SExp)

	if !cdr.IsNil() {
		cadr = cdr.Car()
	}

	//	3. a value isn't already bound to symbol in the current
	//	   closure

	if _, ok := ctx.Closure.Local(carSymbol); ok {
		return nil, Error(
			"a value is already bound to the symbol",
			car.Meta(),
			carSymbol,
		)
	}

	// Evaluate cadr (the value).
	v, err := ctx.Eval(ctx, cadr, ctx.IsTail)
	if err != nil {
		return nil, err
	}

	// Bind the value to the symbol and return the value.
	ctx.Closure.SetLocal(carSymbol, v)

	return v, nil
}
