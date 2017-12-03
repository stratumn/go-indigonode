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

// Let is a command that binds a symbol to a value
var Let = BasicCmdWrapper{BasicCmd{
	Name:     "let",
	Use:      "let <Symbol> <Value>",
	Short:    "Bind a symbol to a value in current closure",
	ExecSExp: letExec,
}}

func letExec(
	ctx context.Context,
	cli CLI,
	closure *script.Closure,
	call script.CallHandler,
	args script.SCell,
	meta script.Meta,
) (script.SExp, error) {
	// Get:
	//
	//	1. car symbol value
	if args == nil {
		return nil, NewUseError("missing symbol")
	}

	car := args.Car()
	if args == nil {
		return nil, NewUseError("missing symbol")
	}

	carSymbol, ok := car.SymbolVal()
	if !ok {
		return nil, NewUseError("not a symbol")
	}

	//	2. cadr, optional (value)
	cdr := args.Cdr()
	var cadr script.SExp

	if cdr != nil {
		cdrCell, ok := cdr.CellVal()
		if !ok {
			return nil, NewUseError("invalid value")
		}

		cadr = cdrCell.Car()
	}

	//	3. a value isn't already bound to symbol in the current
	//	   closure

	if _, ok := closure.Local("$" + carSymbol); ok {
		return nil, NewUseError("a value is already bound to the symbol")
	}

	// Evaluate cadr (the value).
	v, err := script.Eval(closure.Resolve, call, cadr)
	if err != nil {
		return nil, err
	}

	// Bind the value to the symbol and return the value.
	closure.SetLocal("$"+carSymbol, v)

	return v, nil
}
