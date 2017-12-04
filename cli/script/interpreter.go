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
	"context"
	"fmt"
	"os"

	"github.com/pkg/errors"
)

// InterpreterOpt is a closure options.
type InterpreterOpt func(*Interpreter)

// InterpreterOptClosure sets the top closure.
func InterpreterOptClosure(c *Closure) InterpreterOpt {
	return func(itr *Interpreter) {
		itr.closure = c
	}
}

// InterpreterOptErrorHandler sets the error handler.
func InterpreterOptErrorHandler(h func(error)) InterpreterOpt {
	return func(itr *Interpreter) {
		itr.errHandler = h
	}
}

// InterpreterOptValueHandler sets the value handler.
//
// The handler receives evaluated expressions.
func InterpreterOptValueHandler(h func(SExp)) InterpreterOpt {
	return func(itr *Interpreter) {
		itr.valHandler = h
	}
}

// InterpreterOptFuncHandlers adds function handlers.
func InterpreterOptFuncHandlers(m map[string]InterpreterFuncHandler) InterpreterOpt {
	return func(itr *Interpreter) {
		for k, h := range m {
			itr.funcHandlers[k] = h
		}
	}
}

// InterpreterFuncHandler handle a call to a function.
type InterpreterFuncHandler func(*InterpreterContext) (SExp, error)

// InterpreterContext is passed to a function handler.
type InterpreterContext struct {
	Ctx context.Context

	Name    string
	Closure *Closure
	Args    SCell
	Meta    Meta

	// These can be used to evaluate arguments.
	Eval              func(context.Context, *Closure, SExp) (SExp, error)
	EvalList          func(context.Context, *Closure, SCell) (SCell, error)
	EvalListToSlice   func(context.Context, *Closure, SCell) (SExpSlice, error)
	EvalListToStrings func(context.Context, *Closure, SCell) ([]string, error)
}

// Interpreter evaluates S-Expressions.
type Interpreter struct {
	closure *Closure

	funcHandlers map[string]InterpreterFuncHandler
	errHandler   func(error)
	valHandler   func(SExp)
}

// NewInterpreter creates a new interpreter.
func NewInterpreter(opts ...InterpreterOpt) *Interpreter {
	itr := &Interpreter{
		funcHandlers: map[string]InterpreterFuncHandler{},
	}

	for _, o := range opts {
		o(itr)
	}

	if itr.closure == nil {
		itr.closure = NewClosure()
	}

	if itr.errHandler == nil {
		itr.errHandler = func(err error) {
			fmt.Fprintln(os.Stderr, err)
		}
	}

	if itr.valHandler == nil {
		itr.valHandler = func(exp SExp) {
			fmt.Println(exp)
		}
	}

	return itr
}

// EvalInput evaluates the given input.
func (itr *Interpreter) EvalInput(ctx context.Context, in string) error {
	var scanError error

	scanner := NewScanner(ScannerOptErrorHandler(func(err error) {
		itr.errHandler(err)
		scanError = err
	}))

	parser := NewParser(scanner)

	instrs, err := parser.Parse(in)
	if err != nil {
		return err
	}

	if scanError != nil {
		return err
	}

	return itr.evalInstrs(ctx, instrs)
}

// evalInstrs evaluates a list of instructions.
func (itr *Interpreter) evalInstrs(ctx context.Context, instrs SCell) error {
	if instrs == nil {
		return nil
	}

	car := instrs.Car()

	val, err := itr.eval(ctx, itr.closure, car)
	if err != nil {
		return err
	}

	itr.valHandler(val)

	cdr := instrs.Cdr()
	if cdr == nil {
		return nil
	}

	return itr.evalInstrs(ctx, cdr.MustCellVal())
}

// eval evaluates an expression.
func (itr *Interpreter) eval(
	ctx context.Context,
	closure *Closure,
	exp SExp,
) (SExp, error) {
	if exp == nil {
		return nil, nil
	}

	switch exp.UnderlyingType() {
	case TypeCell:
		return itr.evalCell(ctx, closure, exp.MustCellVal())

	case TypeSymbol:
		return closure.Resolve(exp)

	default:
		return exp, nil
	}
}

// evalCell evaluates a cell.
func (itr *Interpreter) evalCell(
	ctx context.Context,
	closure *Closure,
	cell SCell,
) (SExp, error) {
	meta := Meta{}

	wrapError := func(err error) error {
		if cell == nil {
			return errors.WithStack(err)
		}

		return errors.Wrapf(
			err,
			"%d:%d: %v",
			meta.Line,
			meta.Offset,
			cell.Car(),
		)
	}

	if cell != nil && cell.IsList() {
		meta = cell.Meta()
		car := cell.Car()

		if car.UnderlyingType() == TypeSymbol {
			name := car.MustSymbolVal()
			handler, ok := itr.funcHandlers[name]
			if !ok {
				err := wrapError(ErrUnknownFunc)
				itr.errHandler(err)
				return nil, err
			}

			var args SCell

			if cdr := cell.Cdr(); cdr != nil {
				args = cdr.MustCellVal()
			}

			handlerCtx := &InterpreterContext{
				Ctx:               ctx,
				Name:              name,
				Closure:           closure,
				Args:              args,
				Meta:              meta,
				Eval:              itr.eval,
				EvalList:          itr.evalList,
				EvalListToSlice:   itr.evalListToSlice,
				EvalListToStrings: itr.evalListToStrings,
			}

			val, err := handler(handlerCtx)
			if err != nil {
				err = wrapError(err)
				itr.errHandler(err)
			}

			return val, err
		}
	}

	err := wrapError(ErrFuncName)
	itr.errHandler(err)
	return nil, err
}

// evalList evaluates each expression in a list and returns a list.
//
// It assumes a valid list is given.
func (itr *Interpreter) evalList(
	ctx context.Context,
	closure *Closure,
	list SCell,
) (SCell, error) {
	if list == nil {
		return nil, nil
	}

	car := list.Car()

	carVal, err := itr.eval(ctx, closure, car)
	if err != nil {
		return nil, err
	}

	var meta Meta

	if carVal != nil {
		carValMeta := carVal.Meta()
		meta = Meta{
			Line:   carValMeta.Line,
			Offset: carValMeta.Offset,
		}
	} else if car != nil {
		carMeta := car.Meta()
		meta = Meta{
			Line:   carMeta.Line,
			Offset: carMeta.Offset,
		}
	}

	cdr := list.Cdr()
	if cdr == nil {
		return Cons(carVal, nil, meta), nil
	}

	cdrVal, err := itr.evalList(ctx, closure, cdr.MustCellVal())
	if err != nil {
		return nil, err
	}

	return Cons(carVal, cdrVal, meta), nil
}

// evalListToSlice evaluates each expression in a list and returns a slice of
// expressions.
//
// It assumes a valid list is given.
func (itr *Interpreter) evalListToSlice(
	ctx context.Context,
	closure *Closure,
	list SCell,
) (SExpSlice, error) {
	vals, err := itr.evalList(ctx, closure, list)
	if err != nil {
		return nil, err
	}

	if vals == nil {
		return nil, nil
	}

	return vals.MustToSlice(), nil
}

// evalListToStrings evaluates each expression in a list and returns a slice
// containing the string values of each result.
//
// Nil values will be empty strings.
//
// It assumes a valid list is given.
func (itr *Interpreter) evalListToStrings(
	ctx context.Context,
	closure *Closure,
	list SCell,
) ([]string, error) {
	vals, err := itr.evalListToSlice(ctx, closure, list)
	if err != nil {
		return nil, err
	}

	strings := vals.Strings(false)

	// Replace <nil> with empty strings.
	for i, v := range vals {
		if v == nil {
			strings[i] = ""
		}
	}

	return strings, nil
}
