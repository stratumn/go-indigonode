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

// LibInt contains functions to quote and evaluate expressions.
var LibInt = map[string]InterpreterFuncHandler{
	"+":   LibIntAdd,
	"-":   LibIntSub,
	"*":   LibIntMul,
	"/":   LibIntDiv,
	"mod": LibIntMod,
	"=":   LibIntEq,
	"<":   LibIntLt,
	">":   LibIntGt,
	"<=":  LibIntLte,
	">=":  LibIntGte,
}

// LibIntAdd adds integers to the first expression.
func LibIntAdd(ctx *InterpreterContext) (SExp, error) {
	return libIntOp(ctx, func(acc int64, i int64) (int64, error) {
		return acc + i, nil
	})
}

// LibIntSub substracts integers from the first expression.
func LibIntSub(ctx *InterpreterContext) (SExp, error) {
	return libIntOp(ctx, func(acc int64, i int64) (int64, error) {
		return acc - i, nil
	})
}

// LibIntMul multiplies integers.
func LibIntMul(ctx *InterpreterContext) (SExp, error) {
	return libIntOp(ctx, func(acc int64, i int64) (int64, error) {
		return acc * i, nil
	})
}

// LibIntDiv divides the first expression by the remaining expressions.
func LibIntDiv(ctx *InterpreterContext) (SExp, error) {
	return libIntOp(ctx, func(acc int64, i int64) (int64, error) {
		if i == 0 {
			return 0, ErrDivByZero
		}

		return acc / i, nil
	})
}

// LibIntMod returns the modulo of integers.
func LibIntMod(ctx *InterpreterContext) (SExp, error) {
	return libIntOp(ctx, func(acc int64, i int64) (int64, error) {
		if i == 0 {
			return 0, ErrDivByZero
		}

		return acc % i, nil
	})
}

// LibIntEq returns true if all integers are equal.
func LibIntEq(ctx *InterpreterContext) (SExp, error) {
	return libIntCmp(ctx, func(left int64, right int64) bool {
		return left == right
	})
}

// LibIntLt returns true if all integers are less than the integer to their
// right.
func LibIntLt(ctx *InterpreterContext) (SExp, error) {
	return libIntCmp(ctx, func(left int64, right int64) bool {
		return left < right
	})
}

// LibIntGt returns true if all integers are greater than the integer to their
// right.
func LibIntGt(ctx *InterpreterContext) (SExp, error) {
	return libIntCmp(ctx, func(left int64, right int64) bool {
		return left > right
	})
}

// LibIntLte returns true if all integers are less or equal to the integer to
// their right.
func LibIntLte(ctx *InterpreterContext) (SExp, error) {
	return libIntCmp(ctx, func(left int64, right int64) bool {
		return left <= right
	})
}

// LibIntGte returns true if all integers are greater or equal to the integer
// to their right.
func LibIntGte(ctx *InterpreterContext) (SExp, error) {
	return libIntCmp(ctx, func(left int64, right int64) bool {
		return left >= right
	})
}

func libIntOp(
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
		return nil, Error("not an integer", tail.Meta(), "")
	}

	acc, ok := v.Int64Val()
	if !ok {
		return nil, Error("not an integer", v.Meta(), "")
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
			return nil, Error("not an integer", tail.Meta(), "")
		}

		i, ok := v.Int64Val()
		if !ok {
			return nil, Error("not an integer", v.Meta(), "")
		}

		acc, err = op(acc, i)
		if err != nil {
			return nil, WrapError(err, v.Meta(), "")
		}

		cdr := tail.Cdr()
		if cdr == nil {
			break
		}

		tail = cdr.MustCellVal()
	}

	return Int64(acc, ctx.Meta), nil
}

func libIntCmp(
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
		return nil, Error("not an integer", tail.Meta(), "")
	}

	left, ok := v.Int64Val()
	if !ok {
		return nil, Error("not an integer", v.Meta(), "")
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
			return nil, Error("not an integer", tail.Meta(), "")
		}

		right, ok := v.Int64Val()
		if !ok {
			return nil, Error("not an integer", v.Meta(), "")
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
