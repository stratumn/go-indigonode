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

// LibMacro contains preprocessor functions to work with macros.
var LibMacro = map[string]InterpreterFuncHandler{
	"def-macro": LibMacroDefMacro,
}

// LibMacroDefMacro binds a preprocessor symbol to a macro function.
func LibMacroDefMacro(ctx *InterpreterContext) (SExp, error) {
	// Make sure that:
	//
	//	1. car is a symbol
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

	//	2. cadr is nil or a list of symbols (macro arguments)
	cdr := ctx.Args.Cdr()
	if cdr.IsNil() {
		return nil, errors.New("missing macro arguments")
	}

	cadr := cdr.Car()

	if !cadr.IsNil() {
		if !IsList(cadr) {
			return nil, Error(
				"macro arguments are not a list",
				cadr.Meta(),
				"",
			)
		}

		for _, exp := range ToSlice(cadr) {
			if exp.UnderlyingType() != SExpSymbol {
				return nil, Error(
					"macro argument is not a symbol",
					cadr.Meta(),
					"",
				)
			}
		}
	}

	//	3. caddr (macro body) is there but can be anything including
	//	   nil
	cddr := ctx.Args.Cdr()
	if cddr.IsNil() {
		return nil, errors.New("missing macro body")
	}

	if !cddr.Cdr().IsNil() {
		return nil, Error(
			"extra expressions after macro body",
			cddr.Meta(),
			"",
		)
	}

	//	4. a value isn't already bound to symbol in the current closure
	if _, ok := ctx.Closure.Local(carSymbol); ok {
		return nil, Error(
			"a value is already bound to the symbol",
			car.Meta(),
			carSymbol,
		)
	}

	// To differenciate lambda functions from other expressions, FuncData
	// is stored in the meta, which is also used to store the parent
	// closure of the function. The difference between a macro and a regular
	// function is that the arguments are not evaluated.
	lambda := Cons(
		Symbol(LambdaSymbol, ctx.Meta),
		ctx.Args,
		Meta{
			Line:   ctx.Meta.Line,
			Offset: ctx.Meta.Offset,
			UserData: FuncData{
				ParentClosure: ctx.Closure,
			},
		},
	)

	// Bind the lambda to the preprocessor symbol and return it.
	ctx.Closure.SetLocal(carSymbol, lambda)

	return lambda, nil
}
