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

// LibType contains functions to work with types.
var LibType = map[string]InterpreterFuncHandler{
	"nil?":    LibTypeIsNil,
	"atom?":   LibTypeIsAtom,
	"list?":   LibTypeIsList,
	"sym?":    LibTypeIsSym,
	"string?": LibTypeIsString,
	"int64?":  LibTypeIsInt64,
	"bool?":   LibTypeIsBool,
}

// LibTypeIsNil returns whether an expression is nil.
func LibTypeIsNil(ctx *InterpreterContext) (SExp, error) {
	if ctx.Args.IsNil() || !ctx.Args.Cdr().IsNil() {
		return nil, errors.New("expected a single expression")
	}

	v, err := ctx.Eval(ctx, ctx.Args.Car(), false)
	if err != nil {
		return nil, err
	}

	return Bool(v.IsNil(), ctx.Meta), nil
}

// LibTypeIsAtom returns whether an expression is an atom. Nil is an atom.
func LibTypeIsAtom(ctx *InterpreterContext) (SExp, error) {
	if ctx.Args.IsNil() || !ctx.Args.Cdr().IsNil() {
		return nil, errors.New("expected a single expression")
	}

	v, err := ctx.Eval(ctx, ctx.Args.Car(), false)
	if err != nil {
		return nil, err
	}

	return Bool(v.UnderlyingType() != SExpCell || v.IsNil(), ctx.Meta), nil
}

// LibTypeIsList returns whether an expression is a list. Nil is not a list.
func LibTypeIsList(ctx *InterpreterContext) (SExp, error) {
	if ctx.Args.IsNil() || !ctx.Args.Cdr().IsNil() {
		return nil, errors.New("expected a single expression")
	}

	v, err := ctx.Eval(ctx, ctx.Args.Car(), false)
	if err != nil {
		return nil, err
	}

	return Bool(IsList(v), ctx.Meta), nil
}

// LibTypeIsSym returns whether an expression is of type symbol.
func LibTypeIsSym(ctx *InterpreterContext) (SExp, error) {
	return libTypeIsType(ctx, SExpSymbol)
}

// LibTypeIsString returns whether an expression is of type string.
func LibTypeIsString(ctx *InterpreterContext) (SExp, error) {
	return libTypeIsType(ctx, SExpString)
}

// LibTypeIsInt64 returns whether an expression is of type int64.
func LibTypeIsInt64(ctx *InterpreterContext) (SExp, error) {
	return libTypeIsType(ctx, SExpInt64)
}

// LibTypeIsBool returns whether an expression is of type bool.
func LibTypeIsBool(ctx *InterpreterContext) (SExp, error) {
	return libTypeIsType(ctx, SExpBool)
}

func libTypeIsType(ctx *InterpreterContext, typ SExpType) (SExp, error) {
	if ctx.Args.IsNil() || !ctx.Args.Cdr().IsNil() {
		return nil, errors.New("expected a single expression")
	}

	v, err := ctx.Eval(ctx, ctx.Args.Car(), false)
	if err != nil {
		return nil, err
	}

	return Bool(v.UnderlyingType() == typ, ctx.Meta), nil
}
