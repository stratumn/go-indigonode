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
	LibOp,
	LibType,
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
	Args    SExp
	Meta    Meta

	VarPrefix string

	IsTail bool

	// These can be used to evaluate arguments. The last argument should
	// be set to IsTail if it is the value being returned.
	Eval              func(*InterpreterContext, SExp, bool) (SExp, error)
	EvalList          func(*InterpreterContext, SExp, bool) (SExp, error)
	EvalListToSlice   func(*InterpreterContext, SExp, bool) (SExpSlice, error)
	EvalListToStrings func(*InterpreterContext, SExp, bool) ([]string, error)

	// This one evaluates a "body", such as a function body that may
	// contain multiple expressions.
	EvalBody func(*InterpreterContext, SExp, bool) (SExp, error)

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
	FuncID uint64
	Args   SExpSlice
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

	if err := itr.EvalInstrs(ctx, instrs); err != nil {
		itr.errHandler(err)
		return err
	}

	return nil
}

// EvalInstrs evaluates a list of instructions.
func (itr *Interpreter) EvalInstrs(ctx context.Context, instrs SExp) error {
	if instrs == Nil() {
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
	if cdr.IsNil() {
		return nil
	}

	return itr.EvalInstrs(ctx, cdr)
}

// eval evaluates an expression.
func (itr *Interpreter) eval(
	ctx *InterpreterContext,
	exp SExp,
	isTail bool,
) (SExp, error) {
	switch exp.UnderlyingType() {
	case TypeCell:
		if exp.IsNil() {
			return Nil(), nil
		}

		return itr.evalCall(ctx, exp, isTail)

	case TypeSymbol:
		return ctx.Closure.Resolve(exp)

	default:
		return exp, nil
	}
}

// evalCall evaluates a function call.
func (itr *Interpreter) evalCall(
	ctx *InterpreterContext,
	call SExp,
	isTail bool,
) (SExp, error) {
	meta := call.Meta()
	name := ""

	wrapError := func(err error) error {
		return WrapError(err, meta, name)
	}

	if call.IsNil() {
		return nil, wrapError(ErrInvalidCall)
	}

	if !IsList(call) {
		return nil, wrapError(ErrInvalidCall)
	}

	car := call.Car()
	if car.IsNil() {
		return nil, wrapError(ErrInvalidCall)
	}

	meta = car.Meta()

	name, ok := car.SymbolVal()
	if !ok {
		return nil, wrapError(ErrFuncName)
	}

	args := call.Cdr()

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
		funcIDs:           ctx.funcIDs,
	}

	lambda, ok := ctx.Closure.Get(itr.varPrefix + name)
	if ok {
		val, err := itr.execFunc(funcCtx, lambda)
		if err != nil {
			return nil, wrapError(err)
		}

		return val, nil
	}

	handler, ok := itr.funcHandlers[name]
	if !ok {
		return nil, wrapError(ErrUnknownFunc)
	}

	val, err := handler(funcCtx)
	if err != nil {
		return nil, wrapError(err)
	}

	if val == nil {
		return nil, wrapError(errors.New("handler returned nil"))
	}

	return val, nil
}

// evalList evaluates each expression in a list and returns a list.
//
// It assumes a valid list is given.
func (itr *Interpreter) evalList(
	ctx *InterpreterContext,
	list SExp,
	isTail bool,
) (SExp, error) {
	if list.IsNil() {
		return Nil(), nil
	}

	car, cdr := list.Car(), list.Cdr()

	carVal, err := itr.eval(ctx, car, isTail && cdr.IsNil())
	if err != nil {
		return nil, err
	}

	var meta Meta

	if !carVal.IsNil() {
		carValMeta := carVal.Meta()
		meta = Meta{
			Line:   carValMeta.Line,
			Offset: carValMeta.Offset,
		}
	} else if !car.IsNil() {
		carMeta := car.Meta()
		meta = Meta{
			Line:   carMeta.Line,
			Offset: carMeta.Offset,
		}
	}

	if cdr.IsNil() {
		return Cons(carVal, nil, meta), nil
	}

	cdrVal, err := itr.evalList(ctx, cdr, isTail)
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
	list SExp,
	isTail bool,
) (SExpSlice, error) {
	vals, err := itr.evalList(ctx, list, isTail)
	if err != nil {
		return nil, err
	}

	return ToSlice(vals), nil
}

// evalListToStrings evaluates each expression in a list and returns a slice
// containing the string values of each result.
//
// Nil values will be empty strings.
//
// It assumes a valid list is given.
func (itr *Interpreter) evalListToStrings(
	ctx *InterpreterContext,
	list SExp,
	isTail bool,
) ([]string, error) {
	vals, err := itr.evalListToSlice(ctx, list, isTail)
	if err != nil {
		return nil, err
	}

	strings := vals.Strings(false)

	// Replace <nil> with empty strings.
	for i, v := range vals {
		if v.IsNil() {
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
		v, err := itr.eval(ctx, body, isTail)
		return v, err
	}

	// Eval directly unless body is a list and car is a list.
	if body.IsNil() || !IsList(body) {
		return evalDirectly()
	}

	car := body.Car()
	if car.IsNil() || !IsList(car) {
		return evalDirectly()
	}

	// Eval each expression in the body list.
	vals, err := itr.evalListToSlice(ctx, body, isTail)
	if err != nil {
		return nil, err
	}

	// Return the last evaluated expression if there is one.
	l := len(vals)
	if l > 0 {
		return vals[l-1], nil
	}

	return Nil(), nil
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
	data, ok := lambda.Meta().UserData.(FuncData)
	if !ok {
		return nil, errors.WithStack(ErrNotFunc)
	}

	// Assumes the function has already been checked when created.
	//
	// Ignore car of lambda which normally contains the symbol 'lambda'.
	//
	// Get:
	//
	//	1. cadr of lambda (the argument symbols)
	cdr := lambda.Cdr()
	cadr := cdr.Car()

	//	2. caddr of lambda (the function body)
	cddr := cdr.Cdr()
	caddr := cddr.Car()
	cadrSlice := ToSlice(cadr)

	// Evalute the arguments.
	args, err := itr.evalListToSlice(ctx, ctx.Args, false)
	if err != nil {
		return nil, err
	}

	// If this is a tail call and optimizations are enabled, check if this
	// is a recursive call. If it is the case, return a lazy lambda
	// instead of evaluating.
	if itr.tailOptimize && ctx.IsTail {
		for _, id := range ctx.funcIDs {
			if id == data.ID {
				meta := ctx.Meta
				meta.UserData = LazyCall{
					FuncID: data.ID,
					Args:   args,
				}
				return Cons(nil, nil, meta), nil
			}
		}
	}

	// Create a closure for the function body.
	bodyClosure := NewClosure(ClosureOptParent(data.ParentClosure))

	// Set the closure values from the arguments.
	if err := itr.setArgs(bodyClosure, cadrSlice, args); err != nil {
		return nil, WrapError(err, ctx.Meta, ctx.Name)
	}

	// Create the context for executing the body.
	bodyCtx := &InterpreterContext{
		Ctx:     ctx.Ctx,
		Closure: bodyClosure,
		funcIDs: append(ctx.funcIDs, data.ID),
	}

	// Finally, evaluate the function body.
	val, err := itr.evalBody(bodyCtx, caddr, true)
	if itr.tailOptimize {
		// Handle lazy functions from tail call optimizations.
		for {
			if err != nil || val.IsNil() {
				return val, err
			}

			lazy, ok := val.Meta().UserData.(LazyCall)
			if !ok || lazy.FuncID != data.ID {
				return val, nil
			}

			bodyClosure = NewClosure(ClosureOptParent(data.ParentClosure))

			err = itr.setArgs(bodyClosure, cadrSlice, lazy.Args)
			if err != nil {
				return nil, err
			}

			bodyCtx = &InterpreterContext{
				Ctx:     ctx.Ctx,
				Closure: bodyClosure,
				funcIDs: append(ctx.funcIDs, data.ID),
			}

			val, err = itr.evalBody(bodyCtx, caddr, true)
		}
	}

	return val, err
}

// setArgs sets the values of a closure from funciton arguments.
func (itr *Interpreter) setArgs(
	closure *Closure,
	argSyms SExpSlice,
	args SExpSlice,
) error {
	if len(args) != len(argSyms) {
		return errors.New("unexpected number of arguments")
	}

	for i, symbol := range argSyms {
		closure.SetLocal(itr.varPrefix+symbol.MustSymbolVal(), args[i])
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
