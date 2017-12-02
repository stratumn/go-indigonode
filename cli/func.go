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

// ExecFunc executes a function.
//
// The function arguments are evaluated in the given closure.
func ExecFunc(
	ctx context.Context,
	closure *script.Closure,
	createCall CallerWithClosure,
	exp *script.SExp,
	args *script.SExp,
) (*script.SExp, error) {
	data, ok := exp.UserData.(FuncData)
	if !ok {
		return nil, errors.WithStack(ErrNotFunc)
	}

	fn := exp.List

	if fn == nil {
		return nil, NewUseError("expected function not to be nil")
	}

	execClosure := script.NewClosure(script.OptParent(data.ParentClosure))
	argCall := createCall(ctx, closure)
	arg := args

	for head := fn.Cdr.List; head != nil; head = head.Cdr {
		if arg == nil {
			return nil, NewUseError("missing argument")
		}

		v, err := arg.ResolveEval(closure.Resolve, argCall)
		if err != nil {
			return nil, err
		}

		execClosure.Set("$"+head.Str, v)

		arg = arg.Cdr
	}

	if arg != nil {
		return nil, NewUseError("too many argument")
	}

	fnCall := createCall(ctx, execClosure)

	return evalSExpBody(execClosure.Resolve, fnCall, fn.Cdr.Cdr)
}
