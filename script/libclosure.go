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

// LibClosure contains functions to work with closures.
var LibClosure = map[string]InterpreterFuncHandler{
	"let": LibClosureLet,
}

// LibClosureLet binds a value to a symbol in the current closure.
func LibClosureLet(ctx *InterpreterContext) (SExp, error) {
	// Get:
	//
	//	1. car symbol value
	if ctx.Args == nil {
		return nil, errors.New("missing symbol")
	}

	car := ctx.Args.Car()
	if ctx.Args == nil {
		return nil, errors.New("missing symbol")
	}

	carSymbol, ok := car.SymbolVal()
	if !ok {
		return nil, Error("not a symbol", car.Meta(), "")
	}

	//	2. cadr, optional (value)
	cdr := ctx.Args.Cdr()
	var cadr SExp

	if cdr != nil {
		cdrCell, ok := cdr.CellVal()
		if !ok {
			return nil, Error("invalid value", cdr.Meta(), "")
		}

		cadr = cdrCell.Car()
	}

	//	3. a value isn't already bound to symbol in the current
	//	   closure

	if _, ok := ctx.Closure.Local(ctx.VarPrefix + carSymbol); ok {
		return nil, errors.New("a value is already bound to the symbol")
	}

	// Evaluate cadr (the value).
	v, err := ctx.Eval(ctx.Ctx, ctx.Closure, cadr)
	if err != nil {
		return nil, err
	}

	// Bind the value to the symbol and return the value.
	ctx.Closure.SetLocal(ctx.VarPrefix+carSymbol, v)

	return v, nil
}
