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

// LibLambda contains functions to work with lambda functions.
var LibLambda = map[string]InterpreterFuncHandler{
	"lambda": LibLambdaLambda,
	"λ":      LibLambdaLambda,
}

// LibLambdaLambda creates a lambda function.
func LibLambdaLambda(ctx *InterpreterContext) (SExp, error) {
	// Make sure that:
	//
	//	1. car is nil or a list of symbols (function arguments)
	if ctx.Args.IsNil() {
		return nil, errors.New("missing function arguments")
	}

	car := ctx.Args.Car()
	if !car.IsNil() {
		if !IsList(car) {
			return nil, Error(
				"function arguments are not a list",
				car.Meta(),
				"",
			)
		}

		for _, exp := range ToSlice(car) {
			if exp.UnderlyingType() != SExpSymbol {
				return nil, Error(
					"function argument is not a symbol",
					car.Meta(),
					"",
				)
			}
		}
	}

	//	2. cadr (function body) is there but can be anything including
	//	   nil
	cdr := ctx.Args.Cdr()
	if cdr.IsNil() {
		return nil, errors.New("missing function body")
	}

	if !cdr.Cdr().IsNil() {
		return nil, Error(
			"extra expressions after function body",
			cdr.Meta(),
			"",
		)
	}

	// To differenciate lambda functions from other expressions, FuncData
	// is stored in the meta, which is also used to store the parent
	// closure of the function.
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

	return lambda, nil
}
