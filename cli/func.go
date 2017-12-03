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

	"github.com/pkg/errors"
	"github.com/stratumn/alice/cli/script"
)

// FuncData contains information about a funciton.
type FuncData struct {
	// ParentClosure is the closure the function was defined within.
	ParentClosure *script.Closure
}

// CallerWithClosure create a call handler for a closure.
type CallerWithClosure func(context.Context, *script.Closure) script.CallHandler

// ExecFunc executes a script function.
//
// The function arguments are evaluated in the given closure. A new closure
// is created for the function body. The parent of the function body closure
// is the one stored in the FuncData.
//
// It is assumed that the function was properly created, using Lambda for
// instance.
func ExecFunc(
	ctx context.Context,
	closure *script.Closure,
	createCall CallerWithClosure,
	name string,
	def script.SExp,
	args script.SCell,
) (script.SExp, error) {
	// Make sure that is a function (has FuncData in meta).
	data, ok := def.Meta().UserData.(FuncData)
	if !ok {
		return nil, errors.WithStack(ErrNotFunc)
	}

	// Assumes the function has already been checked when created.
	//
	// Ignore car or def which normally contains the symbol 'lambda'.
	//
	// Get:
	//
	//	1. cadr of def (the argument symbols)
	defCell := def.MustCellVal()
	defCdr := defCell.Cdr()
	defCdrCell := defCdr.MustCellVal()
	defCadr := defCdrCell.Car()

	var defCadrCell script.SCell
	if defCadr != nil {
		defCadrCell = defCadr.MustCellVal()
	}

	//	2. caddr of def (the function body)
	defCddr := defCdrCell.Cdr()
	defCddrCell := defCddr.MustCellVal()
	defCaddr := defCddrCell.Car()

	// Now evaluate the function arguments in the given context.
	argv, err := script.EvalListToSlice(
		closure.Resolve,
		createCall(ctx, closure),
		args,
	)
	if err != nil {
		return nil, err
	}

	// Transform cadr of def (the argument symbols) to a slice if it isn't
	// nil.
	var defCadrSlice script.SExpSlice
	if defCadrCell != nil {
		defCadrSlice = defCadrCell.MustToSlice()
	}

	// Make sure the number of arguments is correct.
	if len(argv) != len(defCadrSlice) {
		return nil, NewUseError("unexpected number of arguments")
	}

	// Create a closure for the function body.
	bodyClosure := script.NewClosure(script.OptParent(data.ParentClosure))

	// Bind the argument vector to symbol values.
	for i, symbol := range defCadrSlice {
		bodyClosure.Set("$"+symbol.MustSymbolVal(), argv[i])
	}

	// Finally, evaluate the function body.
	bodyCall := createCall(ctx, bodyClosure)

	return evalSExpBody(bodyClosure.Resolve, bodyCall, defCaddr)
}
