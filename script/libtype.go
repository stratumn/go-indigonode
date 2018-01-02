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
