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

	if ctx.Args == nil {
		return nil, ctx.Error("missing car")
	}

	car := ctx.Args.Car()

	//	2. cadr (the cell cdr)
	cdr := ctx.Args.Cdr()
	if cdr == nil {
		return nil, ctx.Error("missing cdr")
	}

	cdrCell, ok := cdr.CellVal()
	if !ok {
		return nil, ctx.Error("invalid arguments")
	}

	cadr := cdrCell.Car()

	// Evaluate them.
	carVal, err := ctx.Eval(ctx.Ctx, ctx.Closure, car)
	if err != nil {
		return nil, ctx.WrapError(err)
	}

	cadrVal, err := ctx.Eval(ctx.Ctx, ctx.Closure, cadr)
	if err != nil {
		return nil, ctx.WrapError(err)
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
	if ctx.Args == nil || ctx.Args.Cdr() != nil {
		return nil, ctx.Error("expected a single element")
	}

	v, err := ctx.Eval(ctx.Ctx, ctx.Closure, ctx.Args.Car())
	if err != nil {
		return nil, ctx.WrapError(err)
	}

	if v == nil {
		return nil, nil
	}

	cell, ok := v.CellVal()
	if !ok {
		return nil, ctx.Error("not a cell")
	}

	return cell.Car(), nil
}

// LibCellCdr returns the cdr of a cell.
func LibCellCdr(ctx *InterpreterContext) (SExp, error) {
	// Return the cdr of the evaluated car cell.
	if ctx.Args == nil || ctx.Args.Cdr() != nil {
		return nil, ctx.Error("expected a single element")
	}

	v, err := ctx.Eval(ctx.Ctx, ctx.Closure, ctx.Args.Car())
	if err != nil {
		return nil, ctx.WrapError(err)
	}

	if v == nil {
		return nil, nil
	}

	cell, ok := v.CellVal()
	if !ok {
		return nil, ctx.Error("not a cell")
	}

	return cell.Cdr(), nil
}
