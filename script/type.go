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
)

// Type is an SExp type.
type Type uint8

// Expression types.
const (
	TypeInvalid Type = iota
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
	Type     Type
	IsQuoted bool
}

// TypeError is a type error.
type TypeError struct {
	Line   int
	Offset int
	Got    Type
	Want   Type
}

// Error retruns a string representation of the error.
func (e TypeError) Error() string {
	return fmt.Sprintf(
		"%d:%d: unexpected type %s, want %s",
		e.Line, e.Offset,
		e.Got, e.Want,
	)
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
	cache map[string]TypeInfo

	errHandler func(error)
}

// NewTypeChecker creates a new type checker.
func NewTypeChecker(opts ...TypeCheckerOpt) *TypeChecker {
	tc := &TypeChecker{}

	for _, o := range opts {
		o(tc)
	}

	if tc.errHandler == nil {
		tc.errHandler = func(error) {}
	}

	return tc
}

// Check type-checks the given script.
//
// It assumes the expression is a list of expressions.
func (tc *TypeChecker) Check(exp SExp) {
	for head := exp; head != Nil(); head = head.Cdr() {
		tc.Type(head.Car())
	}
}

func (tc *TypeChecker) error(exp SExp, err error) {
	tc.errHandler(WrapError(err, exp.Meta(), ""))
}

func (tc *TypeChecker) typeError(exp SExp, got, want Type) {
	err := TypeError{
		Line:   exp.Meta().Line,
		Offset: exp.Meta().Offset,
		Got:    got,
		Want:   want,
	}

	tc.errHandler(err)
}

// Type returns the type of an S-Expression.
func (tc *TypeChecker) Type(exp SExp) *TypeInfo {
	if exp.IsAtom() {
		return &TypeInfo{Type: exp.UnderlyingType()}
	}

	sym, ok := exp.Car().SymbolVal()
	if !ok {
		tc.error(exp, ErrInvalidCall)
		return &TypeInfo{Type: TypeInvalid}
	}

	switch sym {
	case "=":
		return &TypeInfo{Type: tc.checkSameTimes(exp, exp.Cdr(), 1, 0)}
	case "+", "-", "*", "/", "mod":
		tc.checkTimes(exp, exp.Cdr(), TypeInt64, 1, 0)
		return &TypeInfo{Type: TypeInt64}
	case "<", ">", "<=", ">=":
		tc.checkTimes(exp, exp.Cdr(), TypeInt64, 1, 0)
		return &TypeInfo{Type: TypeBool}
	case "not":
		tc.checkTimes(exp, exp.Cdr(), TypeBool, 1, 1)
		return &TypeInfo{Type: TypeBool}
	case "and", "or":
		tc.checkTimes(exp, exp.Cdr(), TypeBool, 1, 0)
		return &TypeInfo{Type: TypeBool}
	}

	tc.error(exp, ErrUnknownFunc)

	return &TypeInfo{Type: TypeInvalid}
}

func (tc *TypeChecker) checkTimes(parent SExp, exp SExp, typ Type, min, max int) {
	i, head := 0, exp

	for ; head != Nil(); head = head.Cdr() {
		if max != 0 && i >= max {
			break
		}

		t := tc.Type(head.Car())
		if t.Type == TypeInvalid {
			continue
		}

		if t.Type != typ {
			tc.typeError(head.Car(), t.Type, typ)
		}

		i++
	}

	if min != 0 && i < min {
		tc.error(parent, ErrTypeMissingArg)
	}

	if !head.Car().IsNil() {
		tc.error(head.Car(), ErrTypeExtraArg)
	}
}

func (tc *TypeChecker) checkSameTimes(parent SExp, exp SExp, min, max int) Type {
	typ := TypeInvalid
	i, head := 0, exp

	for ; head != Nil(); head = head.Cdr() {
		if max != 0 && i >= max {
			break
		}

		t := tc.Type(head.Car())
		if t.Type == TypeInvalid {
			continue
		}

		if typ == TypeInvalid {
			typ = t.Type
		}

		if t.Type != typ {
			tc.typeError(head.Car(), t.Type, typ)
		}

		i++
	}

	if min != 0 && i < min {
		tc.error(parent, ErrTypeMissingArg)
	}

	if !head.Car().IsNil() {
		tc.error(head.Car(), ErrTypeExtraArg)
	}

	return typ
}
