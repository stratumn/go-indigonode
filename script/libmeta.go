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

import (
	"github.com/pkg/errors"
)

// LibMeta contains functions to quote and evaluate expressions.
//
// Quasiquote and unquote nesting rules seem to differ amongst Lisp dialects.
// This implementation follows the rules of MIT Scheme:
//
//	https://www.gnu.org/software/mit-scheme/documentation/mit-scheme-ref/Quoting.html
var LibMeta = map[string]InterpreterFuncHandler{
	"quote":      LibMetaQuote,
	"quasiquote": LibMetaQuasiquote,
	"unquote":    LibMetaUnquote,
	"eval":       LibMetaEval,
}

// LibMetaQuote returns an expression without evaluating it.
func LibMetaQuote(ctx *InterpreterContext) (SExp, error) {
	// Quote is the simplest command, it just returns the car as is.
	if ctx.Args.IsNil() || !ctx.Args.Cdr().IsNil() {
		return nil, errors.New("expected a single expression")
	}

	return ctx.Args.Car(), nil
}

// LibMetaQuasiquote returns an expression evaluating only its unquoted
// nested expressions.
func LibMetaQuasiquote(ctx *InterpreterContext) (SExp, error) {
	if ctx.Args.IsNil() || !ctx.Args.Cdr().IsNil() {
		return nil, errors.New("expected a single expression")
	}

	// Only the outermost quasiquote should be processed. This error should
	// never occur because this function is never called unless the quasiquote
	// level is zero.
	if ctx.QuasiquoteLevel != 0 {
		return nil, errors.New("unexpected quasiquote")
	}

	ctx.QuasiquoteLevel = 1
	defer func() { ctx.QuasiquoteLevel = 0 }()

	return evalUnquoted(ctx, ctx.Args.Car(), ctx.IsTail)
}

// LibMetaUnquote evaluates unquoted expressions within a quasiquoted
// expression.
func LibMetaUnquote(ctx *InterpreterContext) (SExp, error) {
	if ctx.Args.IsNil() || !ctx.Args.Cdr().IsNil() {
		return nil, errors.New("expected a single expression")
	}

	// Only unquotes at the same level as the outermost quote should be
	// processed. This error only happens if unquote was used too many times.
	// This function is never called for unquotes with a level greater than the
	// outermost quote.
	if ctx.QuasiquoteLevel != 1 {
		return nil, errors.New("outside of quasiquote")
	}

	ctx.QuasiquoteLevel = 0
	defer func() { ctx.QuasiquoteLevel = 1 }()

	return ctx.Eval(ctx, ctx.Args.Car(), ctx.IsTail)
}

// evalUnquoted evaluates unquoted expressions within a quasiquoted expression.
func evalUnquoted(ctx *InterpreterContext, exp SExp, isTail bool) (SExp, error) {
	if exp.UnderlyingType() == SExpCell {
		name, ok := exp.Car().SymbolVal()
		if !ok {
			return evalCellUnquoted(ctx, exp, isTail)
		}

		// Nested quasiquotes are not processed, but the level is incremented.
		if name == QuasiquoteSymbol {
			ctx.QuasiquoteLevel++
			defer func() { ctx.QuasiquoteLevel-- }()

			val, err := evalCellUnquoted(ctx, exp, isTail)
			if err != nil {
				return nil, err
			}

			// If there were no unquotes at the same level, transform to a
			// regular quote.

			return val, nil
		}

		if name == UnquoteSymbol {
			if ctx.QuasiquoteLevel == 1 {
				// Same level as the outermost quasiquote, so evaluate it.
				return ctx.Eval(ctx, exp, isTail)
			}

			// Otherwise just pop a level.
			ctx.QuasiquoteLevel--
			defer func() { ctx.QuasiquoteLevel++ }()
		}

		return evalCellUnquoted(ctx, exp, isTail)
	}

	return exp, nil
}

// evalCellUnquoted evaluates unquoted expressions within the car and cdr of
// a cell.
func evalCellUnquoted(ctx *InterpreterContext, exp SExp, isTail bool) (SExp, error) {
	if exp.IsNil() {
		return Nil(), nil
	}

	cdr := exp.Cdr()

	isTail = isTail && cdr.IsNil()

	carVal, err := evalUnquoted(ctx, exp.Car(), isTail)
	if err != nil {
		return nil, err
	}

	cdrVal, err := evalUnquoted(ctx, cdr, isTail)
	if err != nil {
		return nil, err
	}

	return Cons(carVal, cdrVal, exp.Meta()), nil
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
