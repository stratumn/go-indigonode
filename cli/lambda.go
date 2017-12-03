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

package cli

import (
	"context"

	"github.com/stratumn/alice/cli/script"
)

// Lambda is a command that creates an anomymous function.
var Lambda = BasicCmdWrapper{BasicCmd{
	Name:     "lambda",
	Use:      "lambda <Arguments> <Body>",
	Short:    "Create anonymous function",
	ExecSExp: lambdaExec,
}}

func lambdaExec(
	ctx context.Context,
	cli CLI,
	closure *script.Closure,
	call script.CallHandler,
	args script.SCell,
	meta script.Meta,
) (script.SExp, error) {
	// Make sure that:
	//
	//	1. car is nil or a list of symbols (function arguments)

	if args == nil {
		return nil, NewUseError("missing function arguments")
	}

	car := args.Car()
	if car != nil {
		carCell, ok := car.CellVal()
		if !ok || !carCell.IsList() {
			return nil, NewUseError("function arguments are not a list")
		}

		for _, exp := range carCell.MustToSlice() {
			if exp.UnderlyingType() != script.TypeSymbol {
				return nil, NewUseError("function argument is not a symbol")
			}
		}
	}

	//	2. cadr (function body) is there but can be anything including
	//	   nil

	cdr := args.Cdr()
	if cdr == nil {
		return nil, NewUseError("missing function body")
	}

	cdrCell, ok := cdr.CellVal()
	if !ok {
		return nil, NewUseError("invalid function body")
	}

	if cdrCell.Cdr() != nil {
		return nil, NewUseError("extra expressions after function body")
	}

	// To differenciate lambda functions from other expressions, FuncData
	// is stored in the meta, which is also used to store the parent
	// closure of the function.
	lambda := script.Cons(
		script.Symbol("lambda", meta),
		args,
		script.Meta{
			Line:     meta.Line,
			Offset:   meta.Offset,
			UserData: FuncData{closure},
		},
	)

	return lambda, nil
}
