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
func evalCellUnquoted(ctx *InterpreterContext, cell SExp, isTail bool) (SExp, error) {
	if cell.IsNil() {
		return Nil(), nil
	}

	cdr := cell.Cdr()

	carVal, err := evalUnquoted(ctx, cell.Car(), isTail && cdr.IsNil())
	if err != nil {
		return nil, err
	}

	var cdrVal SExp

	if cdr.UnderlyingType() == SExpCell {
		cdrVal, err = evalUnquoted(ctx, cdr, ctx.IsTail)
	} else {
		cdrVal, err = evalUnquoted(ctx, cdr, ctx.IsTail)
	}

	if err != nil {
		return nil, err
	}

	return Cons(carVal, cdrVal, cell.Meta()), nil
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
