// Copyright © 2017-2018 Stratumn SAS
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
