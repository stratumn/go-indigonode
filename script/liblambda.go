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
