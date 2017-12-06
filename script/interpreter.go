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
	"sync/atomic"

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

// InterpreterOptTailOptimizations sets whether to enable tail call
// optimizations (on by default).
func InterpreterOptTailOptimizations(e bool) InterpreterOpt {
	return func(itr *Interpreter) {
		itr.tailOptimize = e
	}
}

// InterpreterOptVarPrefix sets a prefix for variables.
func InterpreterOptVarPrefix(p string) InterpreterOpt {
	return func(itr *Interpreter) {
		itr.varPrefix = p
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

var libs = []map[string]InterpreterFuncHandler{
	LibCell,
	LibClosure,
	LibCond,
	LibLambda,
	LibMeta,
}

// InterpreterOptBuiltinLibs adds all the builtin function libraries.
func InterpreterOptBuiltinLibs(itr *Interpreter) {
	for _, l := range libs {
		for k, h := range l {
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

	VarPrefix string

	IsTail bool

	// These can be used to evaluate arguments. The last argument should
	// be set to IsTail if it is the value being returned.
	Eval              func(*InterpreterContext, SExp, bool) (SExp, error)
	EvalList          func(*InterpreterContext, SCell, bool) (SCell, error)
	EvalListToSlice   func(*InterpreterContext, SCell, bool) (SExpSlice, error)
	EvalListToStrings func(*InterpreterContext, SCell, bool) ([]string, error)

	// This one evaluates a "body", such as a function body that may
	// contain multiple expressions.
	EvalBody func(*InterpreterContext, SExp, bool) (SExp, error)

	// Creates an error with meta information.
	Error func(string) error

	// Wraps an error with meta information.
	WrapError func(error) error

	// funcIDs are the IDs of functions that lead to that point. It is used
	// to track recursive calls.
	funcIDs []uint64
}

// FuncData contains information about a lambda function.
type FuncData struct {
	// ID is a unique ID for the function.
	ID uint64

	// ParentClosure is the closure the function was defined within.
	ParentClosure *Closure
}

var funcID = uint64(0)

// FuncID generates a unique function ID.
func FuncID() uint64 {
	return atomic.AddUint64(&funcID, 1)
}

// LazyCall contains information about a call to a lambda function. It is
// used for tail call optimization.
type LazyCall struct {
	Ctx    *InterpreterContext
	Lambda SExp
	FuncID uint64
}

// Interpreter evaluates S-Expressions.
type Interpreter struct {
	closure *Closure

	tailOptimize bool
	varPrefix    string

	funcHandlers map[string]InterpreterFuncHandler
	errHandler   func(error)
	valHandler   func(SExp)
}

// NewInterpreter creates a new interpreter.
func NewInterpreter(opts ...InterpreterOpt) *Interpreter {
	itr := &Interpreter{
		tailOptimize: true,
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

// AddFuncHandler adds a function handler.
func (itr *Interpreter) AddFuncHandler(name string, handler InterpreterFuncHandler) {
	itr.funcHandlers[name] = handler
}

// DeleteFuncHandler removes a function handler.
func (itr *Interpreter) DeleteFuncHandler(name string) {
	delete(itr.funcHandlers, name)
}

// EvalInput evaluates the given input.
//
// All errors are sent the the error handler, but the last one is returned for
// convenience.
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

	if err := itr.evalInstrs(ctx, instrs); err != nil {
		itr.errHandler(err)
		return err
	}

	return nil
}

// evalInstrs evaluates a list of instructions.
func (itr *Interpreter) evalInstrs(ctx context.Context, instrs SCell) error {
	if instrs == nil {
		return nil
	}

	car := instrs.Car()

	val, err := itr.eval(&InterpreterContext{
		Ctx:     ctx,
		Closure: itr.closure,
	}, car, false)

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
	ctx *InterpreterContext,
	exp SExp,
	isTail bool,
) (SExp, error) {
	if exp == nil {
		return nil, nil
	}

	switch exp.UnderlyingType() {
	case TypeCell:
		// Do not evaluate lazy calls created by tail call
		// optimizations.
		_, lazy := exp.Meta().UserData.(LazyCall)
		if lazy {
			return exp, nil
		}

		return itr.evalCall(ctx, exp.MustCellVal(), isTail)

	case TypeSymbol:
		return ctx.Closure.Resolve(exp)

	default:
		return exp, nil
	}
}

// evalCall evaluates a function call.
func (itr *Interpreter) evalCall(
	ctx *InterpreterContext,
	call SCell,
	isTail bool,
) (SExp, error) {
	meta := call.Meta()
	name := ""

	wrapError := func(err error) error {
		return WrapError(err, meta, name)
	}

	if call == nil {
		return nil, wrapError(ErrInvalidCall)
	}

	if !call.IsList() {
		return nil, wrapError(ErrInvalidCall)
	}

	car := call.Car()
	if car == nil {
		return nil, wrapError(ErrInvalidCall)
	}

	meta = car.Meta()

	if car.UnderlyingType() != TypeSymbol {
		return nil, wrapError(ErrFuncName)
	}

	name = car.MustSymbolVal()

	var args SCell
	if cdr := call.Cdr(); cdr != nil {
		var ok bool
		args, ok = cdr.CellVal()
		if !ok {
			meta = cdr.Meta()
			return nil, wrapError(ErrInvalidCall)
		}
	}

	funcCtx := &InterpreterContext{
		Ctx:               ctx.Ctx,
		Name:              name,
		Closure:           ctx.Closure,
		Args:              args,
		Meta:              call.Meta(),
		VarPrefix:         itr.varPrefix,
		IsTail:            isTail,
		Eval:              itr.eval,
		EvalList:          itr.evalList,
		EvalListToSlice:   itr.evalListToSlice,
		EvalListToStrings: itr.evalListToStrings,
		EvalBody:          itr.evalBody,
		Error: func(s string) error {
			return wrapError(errors.New(s))
		},
		WrapError: wrapError,
		funcIDs:   ctx.funcIDs,
	}

	lambda, ok := ctx.Closure.Get(itr.varPrefix + name)
	if ok {
		return itr.execFunc(funcCtx, lambda)
	}

	handler, ok := itr.funcHandlers[name]
	if !ok {
		return nil, wrapError(ErrUnknownFunc)
	}

	val, err := handler(funcCtx)

	if err != nil {
		return nil, wrapError(err)
	}

	return val, nil
}

// evalList evaluates each expression in a list and returns a list.
//
// It assumes a valid list is given.
func (itr *Interpreter) evalList(
	ctx *InterpreterContext,
	list SCell,
	isTail bool,
) (SCell, error) {
	if list == nil {
		return nil, nil
	}

	car := list.Car()
	cdr := list.Cdr()

	carVal, err := itr.eval(ctx, car, isTail && cdr == nil)
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

	if cdr == nil {
		return Cons(carVal, nil, meta), nil
	}

	cdrVal, err := itr.evalList(ctx, cdr.MustCellVal(), isTail)
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
	ctx *InterpreterContext,
	list SCell,
	isTail bool,
) (SExpSlice, error) {
	vals, err := itr.evalList(ctx, list, isTail)
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
	ctx *InterpreterContext,
	list SCell,
	isTail bool,
) ([]string, error) {
	vals, err := itr.evalListToSlice(ctx, list, isTail)
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

// evalBody evaluates a "body" such as a function body.
func (itr *Interpreter) evalBody(
	ctx *InterpreterContext,
	body SExp,
	isTail bool,
) (SExp, error) {
	evalDirectly := func() (SExp, error) {
		val, err := itr.eval(ctx, body, isTail)
		return val, err
	}

	// Eval directly unless body is a list and car is a list.
	if body == nil {
		return evalDirectly()
	}

	cell, ok := body.CellVal()
	if !ok || !cell.IsList() {
		return evalDirectly()
	}

	car := cell.Car()
	if car == nil {
		return evalDirectly()
	}

	carCell, ok := car.CellVal()
	if !ok || !carCell.IsList() {
		return evalDirectly()
	}

	// Eval each expression in the body list.
	vals, err := itr.evalListToSlice(ctx, cell, isTail)
	if err != nil {
		return nil, err
	}

	// Return the last evaluated expression if there is one.
	l := len(vals)
	if l > 0 {
		return vals[l-1], nil
	}

	return nil, nil
}

// execFunc executes a script function.
//
// The function arguments are evaluated in the given closure. A new closure
// is created for the function body. The parent of the function body closure
// is the one stored in the FuncData.
//
// It assumes that the function was properly created, using Lambda for
// instance.
func (itr *Interpreter) execFunc(ctx *InterpreterContext, lambda SExp) (SExp, error) {
	// Make sure that is a function (has FuncData in meta).
	if lambda == nil {
		return nil, errors.WithStack(ErrNotFunc)
	}

	data, ok := lambda.Meta().UserData.(FuncData)
	if !ok {
		return nil, errors.WithStack(ErrNotFunc)
	}

	// If this is a tail call and optimizations are enabled, check if this
	// is a recursive call. If it is the case, return a lazy lambda
	// instead of evaluating.
	if itr.tailOptimize && ctx.IsTail {
		for _, id := range ctx.funcIDs {
			if id == data.ID {
				meta := ctx.Meta
				meta.UserData = LazyCall{
					Ctx:    ctx,
					Lambda: lambda,
					FuncID: data.ID,
				}
				return Cons(nil, nil, meta), nil
			}
		}
	}

	// Assumes the function has already been checked when created.
	//
	// Ignore car of lambda which normally contains the symbol 'lambda'.
	//
	// Get:
	//
	//	1. cadr of lambda (the argument symbols)
	lambdaCell := lambda.MustCellVal()
	lambdaCdr := lambdaCell.Cdr()
	lambdaCdrCell := lambdaCdr.MustCellVal()
	lambdaCadr := lambdaCdrCell.Car()

	var lambdaCadrCell SCell
	if lambdaCadr != nil {
		lambdaCadrCell = lambdaCadr.MustCellVal()
	}

	//	2. caddr of lambda (the function body)
	lambdaCddr := lambdaCdrCell.Cdr()
	lambdaCddrCell := lambdaCddr.MustCellVal()
	lambdaCaddr := lambdaCddrCell.Car()

	// Transform cadr of lambda (the argument symbols) to a slice if it
	// isn't nil.
	var lambdaCadrSlice SExpSlice
	if lambdaCadrCell != nil {
		lambdaCadrSlice = lambdaCadrCell.MustToSlice()
	}

	// Create a closure for the function body.
	bodyClosure := NewClosure(ClosureOptParent(data.ParentClosure))

	// Set the closure values from the arguments.
	if err := itr.setArgs(ctx, bodyClosure, lambdaCadrSlice); err != nil {
		return nil, err
	}

	// Create the context for executing the body.
	bodyCtx := &InterpreterContext{
		Ctx:     ctx.Ctx,
		Closure: bodyClosure,
		funcIDs: append(ctx.funcIDs, data.ID),
	}

	// Finally, evaluate the function body.
	val, err := itr.evalBody(bodyCtx, lambdaCaddr, true)

	if itr.tailOptimize {
		// Handle lazy functions from tail call optimizations.
		for {
			if val == nil || err != nil {
				return val, err
			}

			lazy, ok := val.Meta().UserData.(LazyCall)
			if !ok || lazy.FuncID != data.ID {
				return val, nil
			}

			bodyClosure = NewClosure(ClosureOptParent(data.ParentClosure))

			err = itr.setArgs(lazy.Ctx, bodyClosure, lambdaCadrSlice)
			if err != nil {
				return nil, err
			}

			bodyCtx = &InterpreterContext{
				Ctx:     ctx.Ctx,
				Closure: bodyClosure,
				funcIDs: append(ctx.funcIDs, data.ID),
			}

			val, err = itr.evalBody(bodyCtx, lambdaCaddr, true)
		}
	}

	return val, err
}

// setArgs sets the values of a closure from funciton arguments.
func (itr *Interpreter) setArgs(
	callerCtx *InterpreterContext,
	closure *Closure,
	argSyms SExpSlice,
) error {
	argv, err := itr.evalListToSlice(callerCtx, callerCtx.Args, false)
	if err != nil {
		return callerCtx.WrapError(err)
	}

	if len(argv) != len(argSyms) {
		return callerCtx.Error("unexpected number of arguments")
	}

	for i, symbol := range argSyms {
		closure.Set(itr.varPrefix+symbol.MustSymbolVal(), argv[i])
	}

	return nil
}

// Error creates an error with S-Expression meta.
func Error(msg string, meta Meta, ident string) error {
	return WrapError(errors.New(msg), meta, ident)
}

// WrapError wraps an error with S-Expression meta.
func WrapError(err error, meta Meta, ident string) error {
	if err == nil {
		return nil
	}

	if ident == "" {
		return errors.Wrapf(
			err,
			"%d:%d",
			meta.Line,
			meta.Offset,
		)
	}

	return errors.Wrapf(
		err,
		"%d:%d: %v",
		meta.Line,
		meta.Offset,
		ident,
	)
}
