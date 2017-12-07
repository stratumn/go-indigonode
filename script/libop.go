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

import (
	"github.com/pkg/errors"
)

// LibOp contains functions for primitive operations.
var LibOp = map[string]InterpreterFuncHandler{
	"+":   LibOpAdd,
	"-":   LibOpSub,
	"*":   LibOpMul,
	"/":   LibOpDiv,
	"mod": LibOpMod,
	"=":   LibOpEq,
	"<":   LibOpLt,
	">":   LibOpGt,
	"<=":  LibOpLte,
	">=":  LibOpGte,
	"not": LibOpNot,
	"and": LibOpAnd,
	"or":  LibOpOr,
}

// LibOpAdd adds integers to the first expression.
func LibOpAdd(ctx *InterpreterContext) (SExp, error) {
	return libOpInt(ctx, func(acc int64, i int64) (int64, error) {
		return acc + i, nil
	})
}

// LibOpSub substracts integers from the first expression.
func LibOpSub(ctx *InterpreterContext) (SExp, error) {
	return libOpInt(ctx, func(acc int64, i int64) (int64, error) {
		return acc - i, nil
	})
}

// LibOpMul multiplies integers.
func LibOpMul(ctx *InterpreterContext) (SExp, error) {
	return libOpInt(ctx, func(acc int64, i int64) (int64, error) {
		return acc * i, nil
	})
}

// LibOpDiv divides the first expression by the remaining expressions.
func LibOpDiv(ctx *InterpreterContext) (SExp, error) {
	return libOpInt(ctx, func(acc int64, i int64) (int64, error) {
		if i == 0 {
			return 0, ErrDivByZero
		}

		return acc / i, nil
	})
}

// LibOpMod returns the modulo of integers.
func LibOpMod(ctx *InterpreterContext) (SExp, error) {
	return libOpInt(ctx, func(acc int64, i int64) (int64, error) {
		if i == 0 {
			return 0, ErrDivByZero
		}

		return acc % i, nil
	})
}

// LibOpEq returns true if all arguments are equal.
func LibOpEq(ctx *InterpreterContext) (SExp, error) {
	tail := ctx.Args
	if tail == nil {
		return nil, errors.New("missing argument")
	}

	left, err := ctx.Eval(ctx, tail.Car(), false)
	if err != nil {
		return nil, err
	}

	cdr := tail.Cdr()
	if cdr == nil {
		return Bool(true, ctx.Meta), nil
	}

	tail = cdr.MustCellVal()

	for {
		right, err := ctx.Eval(ctx, tail.Car(), false)
		if err != nil {
			return nil, err
		}

		switch {
		case left == nil && right != nil:
			return Bool(false, ctx.Meta), nil
		case left != nil && right == nil:
			return Bool(false, ctx.Meta), nil
		case left == nil && right == nil:
		case !left.Equals(right):
			return Bool(false, ctx.Meta), nil
		}

		cdr = tail.Cdr()
		if cdr == nil {
			break
		}

		tail = cdr.MustCellVal()
		left = right
	}

	return Bool(true, ctx.Meta), nil
}

// LibOpLt returns true if all integers are less than the integer to their
// right.
func LibOpLt(ctx *InterpreterContext) (SExp, error) {
	return libOpIntCmp(ctx, func(left int64, right int64) bool {
		return left < right
	})
}

// LibOpGt returns true if all integers are greater than the integer to their
// right.
func LibOpGt(ctx *InterpreterContext) (SExp, error) {
	return libOpIntCmp(ctx, func(left int64, right int64) bool {
		return left > right
	})
}

// LibOpLte returns true if all integers are less or equal to the integer to
// their right.
func LibOpLte(ctx *InterpreterContext) (SExp, error) {
	return libOpIntCmp(ctx, func(left int64, right int64) bool {
		return left <= right
	})
}

// LibOpGte returns true if all integers are greater or equal to the integer
// to their right.
func LibOpGte(ctx *InterpreterContext) (SExp, error) {
	return libOpIntCmp(ctx, func(left int64, right int64) bool {
		return left >= right
	})
}

func libOpInt(
	ctx *InterpreterContext,
	op func(int64, int64) (int64, error),
) (SExp, error) {
	tail := ctx.Args
	if tail == nil {
		return nil, errors.New("missing argument")
	}

	v, err := ctx.Eval(ctx, tail.Car(), false)
	if err != nil {
		return nil, err
	}

	if v == nil {
		return nil, Error("not an integer", tail.Meta(), "<nil>")
	}

	acc, ok := v.Int64Val()
	if !ok {
		return nil, Error("not an integer", v.Meta(), v.String())
	}

	cdr := tail.Cdr()
	if cdr == nil {
		return Int64(acc, ctx.Meta), nil
	}

	tail = cdr.MustCellVal()

	for {
		v, err := ctx.Eval(ctx, tail.Car(), false)
		if err != nil {
			return nil, err
		}

		if v == nil {
			return nil, Error("not an integer", tail.Meta(), "<nil>")
		}

		i, ok := v.Int64Val()
		if !ok {
			return nil, Error("not an integer", v.Meta(), v.String())
		}

		acc, err = op(acc, i)
		if err != nil {
			return nil, WrapError(err, v.Meta(), v.String())
		}

		cdr := tail.Cdr()
		if cdr == nil {
			break
		}

		tail = cdr.MustCellVal()
	}

	return Int64(acc, ctx.Meta), nil
}

func libOpIntCmp(
	ctx *InterpreterContext,
	cmp func(int64, int64) bool,
) (SExp, error) {
	tail := ctx.Args
	if tail == nil {
		return nil, errors.New("missing argument")
	}

	v, err := ctx.Eval(ctx, tail.Car(), false)
	if err != nil {
		return nil, err
	}

	if v == nil {
		return nil, Error("not an integer", tail.Meta(), "<nil>")
	}

	left, ok := v.Int64Val()
	if !ok {
		return nil, Error("not an integer", v.Meta(), v.String())
	}

	cdr := tail.Cdr()
	if cdr == nil {
		return Bool(true, ctx.Meta), nil
	}

	tail = cdr.MustCellVal()

	for {
		v, err := ctx.Eval(ctx, tail.Car(), false)
		if err != nil {
			return nil, err
		}

		if v == nil {
			return nil, Error("not an integer", tail.Meta(), "<nil>")
		}

		right, ok := v.Int64Val()
		if !ok {
			return nil, Error("not an integer", v.Meta(), v.String())
		}

		if !cmp(left, right) {
			return Bool(false, ctx.Meta), nil
		}

		cdr = tail.Cdr()
		if cdr == nil {
			break
		}

		tail = cdr.MustCellVal()
		left = right
	}

	return Bool(true, ctx.Meta), nil
}

// LibOpNot returns true if the expression is false and true if the expression
// is false.
func LibOpNot(ctx *InterpreterContext) (SExp, error) {
	if ctx.Args == nil || ctx.Args.Cdr() != nil {
		return nil, errors.New("expected a single argument")
	}

	v, err := ctx.Eval(ctx, ctx.Args.Car(), false)
	if err != nil {
		return nil, err
	}

	if v == nil {
		return nil, Error("not a boolean", ctx.Args.Meta(), "<nil>")
	}

	b, ok := v.BoolVal()
	if !ok {
		return nil, Error("not a boolean", v.Meta(), v.String())
	}

	return Bool(!b, ctx.Meta), nil
}

// LibOpAnd returns true if all booleans are true.
func LibOpAnd(ctx *InterpreterContext) (SExp, error) {
	tail := ctx.Args
	if tail == nil {
		return nil, errors.New("missing argument")
	}

	for {
		v, err := ctx.Eval(ctx, tail.Car(), false)
		if err != nil {
			return nil, err
		}

		if v == nil {
			return nil, Error("not a boolean", tail.Meta(), "<nil>")
		}

		b, ok := v.BoolVal()
		if !ok {
			return nil, Error("not a boolean", v.Meta(), v.String())
		}

		if !b {
			return Bool(false, ctx.Meta), nil
		}

		cdr := tail.Cdr()
		if cdr == nil {
			break
		}

		tail = cdr.MustCellVal()
	}

	return Bool(true, ctx.Meta), nil
}

// LibOpOr returns true if at least one boolean is true.
func LibOpOr(ctx *InterpreterContext) (SExp, error) {
	tail := ctx.Args
	if tail == nil {
		return nil, errors.New("missing argument")
	}

	for {
		v, err := ctx.Eval(ctx, tail.Car(), false)
		if err != nil {
			return nil, err
		}

		if v == nil {
			return nil, Error("not a boolean", tail.Meta(), "<nil>")
		}

		b, ok := v.BoolVal()
		if !ok {
			return nil, Error("not a boolean", v.Meta(), v.String())
		}

		if b {
			return Bool(true, ctx.Meta), nil
		}

		cdr := tail.Cdr()
		if cdr == nil {
			break
		}

		tail = cdr.MustCellVal()
	}

	return Bool(false, ctx.Meta), nil
}
