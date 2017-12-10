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
	"fmt"

	"github.com/pkg/errors"
)

var (
	// ErrTypeMissingArg is emitted when a function argument is missing.
	ErrTypeMissingArg = errors.New("missing function argument")

	// ErrTypeExtraArg is emitted when a function call has too many
	// arguments.
	ErrTypeExtraArg = errors.New("extra function argument")

	// ErrCarNil is emitted when the car of a cell is nil.
	ErrCarNil = errors.New("car cannot be nil")

	// ErrNotSymbol is emitted when a symbol is expected.
	ErrNotSymbol = errors.New("not a symbol")

	// ErrNotCell is emitted when a cell is expected.
	ErrNotCell = errors.New("not a cell")

	// ErrNil is emitted when a value cannot be nil.
	ErrNil = errors.New("value cannot be nil")

	// ErrBound is emitted when a symbol is already locally bound.
	ErrBound = errors.New("symbol is already bound locally")

	// ErrNotBound is emitted when a symbol is not bound to a value.
	ErrNotBound = errors.New("symbol is not bound to a value")
)

// Type is an SExp type.
type Type uint8

// Expression types.
const (
	TypeInvalid Type = iota
	TypeNil
	TypeAny
	TypeString
	TypeInt64
	TypeBool
	TypeSymbol
	TypeCell
	TypeFunction
)

// Maps types to their names.
var typeMap = map[Type]string{
	TypeInvalid:  "invalid",
	TypeNil:      "nil",
	TypeAny:      "any",
	TypeString:   "string",
	TypeInt64:    "int64",
	TypeBool:     "bool",
	TypeSymbol:   "symbol",
	TypeCell:     "cell",
	TypeFunction: "function",
}

// String returns a string representation of the type.
func (t Type) String() string {
	return typeMap[t]
}

// TypeInfo contains information about a type.
type TypeInfo struct {
	Type    Type
	IsList  bool
	CarType *TypeInfo
	CdrType *TypeInfo // Nil when list
}

// String returns a string representation of the type info.
func (ti *TypeInfo) String() string {
	s := ""

	switch {
	case ti.Type == TypeCell && ti.IsList:
		s = "list<" + ti.CarType.String() + ">"
	case ti.Type == TypeCell:
		s = "pair<"
		if ti.CarType == nil {
			s += "nil"
		} else {
			s += ti.CarType.String()
		}
		s += ","
		if ti.CdrType == nil {
			s += "nil"
		} else {
			s += ti.CdrType.String()
		}
		s += ">"
	default:
		s = ti.Type.String()
	}

	return s
}

// Equals returns whether the type is the same as another type.
func (ti *TypeInfo) Equals(t *TypeInfo) bool {
	if ti.Type != t.Type {
		return false
	}

	if ti.IsList != t.IsList {
		return false
	}

	if ti.IsList && !ti.CarType.Equals(t.CarType) {
		return false
	}

	if ti.Type == TypeCell {
		if !ti.CarType.Equals(t.CarType) {
			return false
		}

		if !ti.CdrType.Equals(t.CdrType) {
			return false
		}
	}

	return true
}

// Compatible returns whether the given type is compatible with the type info.
//
// The order matters. For instance a receiver with type Any is compatible with
// any type, but not the other way around. A type is only compatible with a
// more specific type.
func (ti *TypeInfo) Compatible(t *TypeInfo) bool {
	if ti.Type == TypeAny {
		return true
	}

	if ti.Type != t.Type {
		return false
	}

	if ti.IsList != t.IsList {
		return false
	}

	if ti.IsList && !ti.CarType.Compatible(t.CarType) {
		return false
	}

	if ti.Type == TypeCell {
		if !ti.CarType.Compatible(t.CarType) {
			return false
		}

		if !ti.CdrType.Compatible(t.CdrType) {
			return false
		}
	}

	return true
}

// TypeError is a type error.
type TypeError struct {
	Line   int
	Offset int
	Got    *TypeInfo
	Want   *TypeInfo
}

// Error retruns a string representation of the error.
func (e TypeError) Error() string {
	return fmt.Sprintf(
		"%d:%d: unexpected type %s, want %s",
		e.Line, e.Offset,
		e.Got, e.Want,
	)
}

// TypeCheckerContext holds information about the context in wich a type is
// being checked.
type TypeCheckerContext struct {
	locals    map[string]*TypeInfo
	inherited map[string]*TypeInfo

	// If an expression is known to be nil, isNil will be true. If it is
	// known to not be nil, isNil will be false. If it is unknown, it won't
	// have a value.
	isNil map[uint64]bool
}

// NewTypeCheckerContext creates a new type checker context.
//
// Parent can be nil.
func NewTypeCheckerContext(parent *TypeCheckerContext) *TypeCheckerContext {
	ctx := &TypeCheckerContext{
		locals:    map[string]*TypeInfo{},
		inherited: map[string]*TypeInfo{},
		isNil:     map[uint64]bool{},
	}

	if parent == nil {
		return ctx
	}

	for k, v := range parent.inherited {
		ctx.inherited[k] = v
	}

	for k, v := range parent.locals {
		ctx.inherited[k] = v
	}

	for k, v := range parent.isNil {
		ctx.isNil[k] = v
	}

	return ctx
}

// Set sets the type of a local symbol.
func (ctx *TypeCheckerContext) Set(symbol string, ti *TypeInfo) {
	ctx.locals[symbol] = ti
}

// Get returns the type of a local or inherited symbol.
func (ctx *TypeCheckerContext) Get(symbol string) (*TypeInfo, bool) {
	if ti, ok := ctx.locals[symbol]; ok {
		return ti, true
	}

	ti, ok := ctx.inherited[symbol]

	return ti, ok
}

// IsLocal retruns whether a value is locally bound to the symbol.
func (ctx *TypeCheckerContext) IsLocal(symbol string) bool {
	_, ok := ctx.locals[symbol]

	return ok
}

// TypeCheckerOpt is a type checker option.
type TypeCheckerOpt func(*TypeChecker)

// TypeCheckerOptErrorHandler sets the type checker's error handler.
func TypeCheckerOptErrorHandler(h func(error)) TypeCheckerOpt {
	return func(p *TypeChecker) {
		p.errHandler = h
	}
}

// TypeChecker type-checks S-Expressions.
type TypeChecker struct {
	cache map[uint64]*TypeInfo

	errHandler func(error)
}

// NewTypeChecker creates a new type checker.
func NewTypeChecker(opts ...TypeCheckerOpt) *TypeChecker {
	tc := &TypeChecker{
		cache: map[uint64]*TypeInfo{},
	}

	for _, o := range opts {
		o(tc)
	}

	if tc.errHandler == nil {
		tc.errHandler = func(error) {}
	}

	return tc
}

func (tc *TypeChecker) error(exp SExp, err error) {
	tc.errHandler(WrapError(err, exp.Meta(), ""))
}

func (tc *TypeChecker) typeError(exp SExp, got, want *TypeInfo) {
	err := TypeError{
		Line:   exp.Meta().Line,
		Offset: exp.Meta().Offset,
		Got:    got,
		Want:   want,
	}

	tc.errHandler(errors.WithStack(err))
}

func (tc *TypeChecker) save(exp SExp, t *TypeInfo) *TypeInfo {
	tc.cache[exp.ID()] = t
	return t
}

// BodyType returns the type info of a body of expressions, emitting type
// errors to the error handler along the way.
func (tc *TypeChecker) BodyType(exp SExp) *TypeInfo {
	var ti *TypeInfo

	ctx := NewTypeCheckerContext(nil)

	if IsList(exp) {
		if !IsList(exp.Car()) {
			return tc.check(ctx, exp)
		}
	} else {
		return tc.check(ctx, exp)
	}

	for head := exp; head != Nil(); head = head.Cdr() {
		ti = tc.check(ctx, head.Car())
	}

	return ti
}

func (tc *TypeChecker) check(ctx *TypeCheckerContext, exp SExp) *TypeInfo {
	if exp.IsNil() {
		return &TypeInfo{Type: TypeNil}
	}

	if symbol, ok := exp.SymbolVal(); ok {
		v, ok := ctx.Get(symbol)
		if !ok {
			tc.error(exp, errors.Wrap(ErrNotBound, symbol))
			return &TypeInfo{Type: TypeInvalid}
		}

		return v
	}

	if exp.IsAtom() {
		return &TypeInfo{Type: exp.UnderlyingType()}
	}

	symbol, ok := exp.Car().SymbolVal()
	if !ok {
		tc.error(exp, ErrInvalidCall)
		return &TypeInfo{Type: TypeInvalid}
	}

	if ti, ok := tc.cache[exp.ID()]; ok {
		return ti
	}

	switch symbol {
	case "=":
		tc.checkSameTimes(ctx, exp, exp.Cdr(), 1, 0)
		return tc.save(exp, &TypeInfo{Type: TypeBool})

	case "+", "-", "*", "/", "mod":
		tc.checkTimes(ctx, exp, exp.Cdr(), &TypeInfo{Type: TypeInt64}, 1, 0)
		return tc.save(exp, &TypeInfo{Type: TypeInt64})

	case "<", ">", "<=", ">=":
		tc.checkTimes(ctx, exp, exp.Cdr(), &TypeInfo{Type: TypeInt64}, 1, 0)
		return tc.save(exp, &TypeInfo{Type: TypeBool})

	case "not":
		tc.checkTimes(ctx, exp, exp.Cdr(), &TypeInfo{Type: TypeBool}, 1, 1)
		return tc.save(exp, &TypeInfo{Type: TypeBool})

	case "and", "or":
		tc.checkTimes(ctx, exp, exp.Cdr(), &TypeInfo{Type: TypeBool}, 1, 0)
		return tc.save(exp, &TypeInfo{Type: TypeBool})

	case "cons":
		return tc.save(exp, tc.checkCons(ctx, exp))

	case "car":
		return tc.save(exp, tc.checkCar(ctx, exp))

	case "cdr":
		return tc.save(exp, tc.checkCdr(ctx, exp))

	case "let":
		return tc.save(exp, tc.checkLet(ctx, exp))

	}

	tc.error(exp, errors.Wrap(ErrUnknownFunc, symbol))

	return tc.save(exp, &TypeInfo{Type: TypeInvalid})
}

// checkTimes checks that arguments are compatible with the given type.
//
// If min is zero, there is no restriction on the minimum number or arguments.
// If max is zero, there is no restriction on the maximum number or arguments.
func (tc *TypeChecker) checkTimes(
	ctx *TypeCheckerContext,
	parent SExp, exp SExp,
	ti *TypeInfo,
	min, max int,
) {
	head := exp
	i := 0

	for ; head != Nil(); head = head.Cdr() {
		if max != 0 && i >= max {
			break
		}

		t := tc.check(ctx, head.Car())
		if t.Type == TypeInvalid {
			i++
			continue
		}

		if !ti.Compatible(t) {
			tc.typeError(head.Car(), t, ti)
		}

		i++
	}

	if min != 0 && i < min {
		tc.error(parent, errors.WithStack(ErrTypeMissingArg))
	}

	if !head.Car().IsNil() {
		tc.error(head.Car(), errors.WithStack(ErrTypeExtraArg))
	}
}

// checkTimes checks that arguments have the same type.
//
// If min is zero, there is no restriction on the minimum number or arguments.
// If max is zero, there is no restriction on the maximum number or arguments.
// It returns the most specific type found.
func (tc *TypeChecker) checkSameTimes(
	ctx *TypeCheckerContext,
	parent SExp, exp SExp,
	min, max int,
) *TypeInfo {
	ti := &TypeInfo{Type: TypeInvalid}
	head := exp
	i := 0

	for ; head != Nil(); head = head.Cdr() {
		if max != 0 && i >= max {
			break
		}

		t := tc.check(ctx, head.Car())
		if t.Type == TypeInvalid {
			i++
			continue
		}

		if ti.Type == TypeInvalid {
			ti = t
		}

		if !ti.Compatible(t) {
			if t.Compatible(ti) {
				ti = t
			} else {
				tc.typeError(head.Car(), t, ti)
			}
		}

		i++
	}

	if min != 0 && i < min {
		tc.error(parent, errors.WithStack(ErrTypeMissingArg))
	}

	if !head.Car().IsNil() {
		tc.error(head.Car(), errors.WithStack(ErrTypeExtraArg))
	}

	return ti
}

func (tc *TypeChecker) checkCons(ctx *TypeCheckerContext, exp SExp) *TypeInfo {
	cdr := exp.Cdr()
	if cdr.IsNil() {
		tc.error(exp, errors.WithStack(ErrTypeMissingArg))
		return &TypeInfo{Type: TypeInvalid}
	}

	cddr := cdr.Cdr()
	if cddr.IsNil() {
		tc.error(exp, errors.WithStack(ErrTypeMissingArg))
		return &TypeInfo{Type: TypeInvalid}
	}

	if !cddr.Cdr().IsNil() {
		tc.error(cddr.Cdr(), errors.WithStack(ErrTypeExtraArg))
		return &TypeInfo{Type: TypeInvalid}
	}

	cadr, caddr := cdr.Car(), cddr.Car()
	lti, rti := tc.check(ctx, cadr), tc.check(ctx, caddr)

	if lti.Type == TypeInvalid || rti.Type == TypeInvalid {
		return &TypeInfo{Type: TypeInvalid}
	}

	ti := &TypeInfo{Type: TypeCell}

	if lti.Type == TypeNil {
		tc.error(cdr, errors.WithStack(ErrCarNil))
		return &TypeInfo{Type: TypeInvalid}
	}

	if rti.Type != TypeNil && !rti.IsList {
		ti.CarType = lti
		ti.CdrType = rti

		return ti
	}

	ti.IsList = true

	if rti.Type != TypeNil && !lti.Equals(rti.CarType) {
		ti.CarType = &TypeInfo{Type: TypeAny}
	} else {
		ti.CarType = lti
	}

	return ti
}

func (tc *TypeChecker) checkCar(ctx *TypeCheckerContext, exp SExp) *TypeInfo {
	cdr := exp.Cdr()
	if cdr.IsNil() {
		tc.error(exp, errors.WithStack(ErrTypeMissingArg))
		return &TypeInfo{Type: TypeInvalid}
	}

	if !cdr.Cdr().IsNil() {
		tc.error(cdr.Cdr(), errors.WithStack(ErrTypeExtraArg))
		return &TypeInfo{Type: TypeInvalid}
	}

	ti := tc.check(ctx, cdr.Car())

	if ti.Type == TypeInvalid {
		return &TypeInfo{Type: TypeInvalid}
	}

	if ti.Type == TypeNil {
		return &TypeInfo{Type: TypeNil}
	}

	if ti.Type != TypeCell {
		tc.error(cdr, errors.WithStack(ErrNotCell))
		return &TypeInfo{Type: TypeInvalid}
	}

	return ti.CarType
}

func (tc *TypeChecker) checkCdr(ctx *TypeCheckerContext, exp SExp) *TypeInfo {
	cdr := exp.Cdr()
	if cdr.IsNil() {
		tc.error(exp, errors.WithStack(ErrTypeMissingArg))
		return &TypeInfo{Type: TypeInvalid}
	}

	if !cdr.Cdr().IsNil() {
		tc.error(cdr.Cdr(), errors.WithStack(ErrTypeExtraArg))
		return &TypeInfo{Type: TypeInvalid}
	}

	ti := tc.check(ctx, cdr.Car())

	if ti.Type == TypeInvalid {
		return &TypeInfo{Type: TypeInvalid}
	}

	if ti.Type == TypeNil {
		return &TypeInfo{Type: TypeNil}
	}

	if ti.Type != TypeCell {
		tc.error(cdr, errors.WithStack(ErrNotCell))
		return &TypeInfo{Type: TypeInvalid}
	}

	if ti.IsList {
		return ti.CarType
	}

	return ti.CdrType
}

func (tc *TypeChecker) checkLet(ctx *TypeCheckerContext, exp SExp) *TypeInfo {
	cdr := exp.Cdr()
	if cdr.IsNil() {
		tc.error(exp, errors.WithStack(ErrTypeMissingArg))
		return &TypeInfo{Type: TypeInvalid}
	}

	cddr := cdr.Cdr()
	if cddr.IsNil() {
		tc.error(exp, errors.WithStack(ErrTypeMissingArg))
		return &TypeInfo{Type: TypeInvalid}
	}

	if !cddr.Cdr().IsNil() {
		tc.error(cddr.Cdr(), errors.WithStack(ErrTypeExtraArg))
		return &TypeInfo{Type: TypeInvalid}
	}

	cadr := cdr.Car()

	if cadr.UnderlyingType() != TypeSymbol {
		tc.error(cdr, errors.WithStack(ErrNotSymbol))
		return &TypeInfo{Type: TypeInvalid}
	}

	caddr := cddr.Car()
	ti := tc.check(ctx, caddr)

	if ti.Type == TypeInvalid {
		return &TypeInfo{Type: TypeInvalid}
	}

	if ti.Type == TypeNil {
		tc.error(cddr, errors.WithStack(ErrNil))
		return &TypeInfo{Type: TypeInvalid}
	}

	symbol := cadr.MustSymbolVal()

	if ctx.IsLocal(symbol) {
		tc.error(cadr, errors.Wrap(ErrBound, symbol))
	}

	ctx.Set(symbol, ti)

	return ti
}
