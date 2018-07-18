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
	"fmt"

	"github.com/pkg/errors"
)

// LibCell contains functions to work with cons cells.
var LibCell = map[string]InterpreterFuncHandler{
	"cons": LibCellCons,
	"car":  LibCellCar,
	"cdr":  LibCellCdr,
}

// LibCellCons constructs a cell.
func LibCellCons(ctx *InterpreterContext) (SExp, error) {
	// Get:
	//
	//	1. car (the cell car)

	if ctx.Args.IsNil() {
		return nil, errors.New("missing car")
	}

	car := ctx.Args.Car()

	//	2. cadr (the cell cdr)
	cdr := ctx.Args.Cdr()
	if cdr.IsNil() {
		return nil, errors.New("missing cdr")
	}

	cadr := cdr.Car()

	// Evaluate them.
	carVal, err := ctx.Eval(ctx, car, false)
	if err != nil {
		return nil, err
	}

	cadrVal, err := ctx.Eval(ctx, cadr, false)
	if err != nil {
		return nil, err
	}

	// Construct the cell.
	return Cons(carVal, cadrVal, Meta{
		Line:   ctx.Meta.Line,
		Offset: ctx.Meta.Offset,
	}), nil
}

// LibCellCar returns the car of a cell.
func LibCellCar(ctx *InterpreterContext) (SExp, error) {
	// Return the car of the evaluated car cell.
	if ctx.Args.IsNil() || !ctx.Args.Cdr().IsNil() {
		return nil, errors.New("expected a single element")
	}

	v, err := ctx.Eval(ctx, ctx.Args.Car(), false)
	if err != nil {
		return nil, err
	}

	if v.UnderlyingType() != SExpCell {
		return nil, Error("not a cell", v.Meta(), fmt.Sprint(v))
	}

	return v.Car(), nil
}

// LibCellCdr returns the cdr of a cell.
func LibCellCdr(ctx *InterpreterContext) (SExp, error) {
	// Return the cdr of the evaluated car cell.
	if ctx.Args.IsNil() || !ctx.Args.Cdr().IsNil() {
		return nil, errors.New("expected a single element")
	}

	v, err := ctx.Eval(ctx, ctx.Args.Car(), false)
	if err != nil {
		return nil, err
	}

	if v.UnderlyingType() != SExpCell {
		return nil, Error("not a cell", v.Meta(), fmt.Sprint(v))
	}

	return v.Cdr(), nil
}
