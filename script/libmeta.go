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

// LibMeta contains functions to quote and evaluate expressions.
var LibMeta = map[string]InterpreterFuncHandler{
	"quote": LibMetaQuote,
	"eval":  LibMetaEval,
}

// LibMetaQuote returns an expression without evaluating it.
func LibMetaQuote(ctx *InterpreterContext) (SExp, error) {
	// Quote is the simplest command, it just returns the car as is.
	if ctx.Args.IsNil() || !ctx.Args.Cdr().IsNil() {
		return nil, errors.New("expected a single expression")
	}

	return ctx.Args.Car(), nil
}

// LibMetaEval evaluates an expression.
func LibMetaEval(ctx *InterpreterContext) (SExp, error) {
	// Eval is a simple command, all it does is evaluate the car twice.
	if ctx.Args.IsNil() || !ctx.Args.Cdr().IsNil() {
		return nil, errors.New("expected a single expression")
	}

	v, err := ctx.Eval(ctx, ctx.Args.Car(), false)
	if err != nil {
		return nil, err
	}

	v, err = ctx.Eval(ctx, v, ctx.IsTail)
	if err != nil {
		return nil, err
	}

	return v, nil
}
