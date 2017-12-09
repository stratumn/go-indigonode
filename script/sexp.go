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
	"bytes"
	"fmt"
	"strconv"
	"sync/atomic"
)

// ResolveHandler resolves symbols.
type ResolveHandler func(sym SExp) (SExp, error)

// ResolveName resolves symbols with their names.
func ResolveName(sym SExp) (SExp, error) {
	return String(sym.MustSymbolVal(), sym.Meta()), nil
}

var cellID = uint64(0)

// CellID generates a unique cell ID.
func CellID() uint64 {
	return atomic.AddUint64(&cellID, 1)
}

// Meta contains metadata for an S-Expression.
type Meta struct {
	Line   int
	Offset int

	// UserData is custom external data.
	UserData interface{}
}

// SExp is represents an S-Expression.
//
// See:
//
//	https://en.wikipedia.org/wiki/S-expression
type SExp interface {
	ID() uint64

	String() string

	Meta() Meta

	UnderlyingType() Type
	IsNil() bool
	IsAtom() bool

	StringVal() (string, bool)
	Int64Val() (int64, bool)
	BoolVal() (bool, bool)
	SymbolVal() (string, bool)

	MustStringVal() string
	MustInt64Val() int64
	MustSymbolVal() string
	MustBoolVal() bool

	Equals(SExp) bool

	Car() SExp
	Cdr() SExp
}

// SExpSlice is a slice of S-Expressions.
type SExpSlice []SExp

// Strings returns the strings value of every S-Expression in the slice.
//
// If quote is true, strings are quoted.
func (s SExpSlice) Strings(quote bool) []string {
	strings := make([]string, len(s))
	for i, exp := range s {
		if !quote {
			if str, ok := exp.StringVal(); ok {
				strings[i] = str
				continue
			}
		}
		strings[i] = fmt.Sprint(exp)
	}

	return strings
}

// nilVal represents the nil value.
var nilVal = &scell{sexp{0, Meta{}}, nil, nil}

func init() {
	nilVal.car = nilVal
	nilVal.cdr = nilVal
}

// Nil returns the nil value.
func Nil() SExp {
	return nilVal
}

// sexp is the base S-Expression type.
type sexp struct {
	id   uint64
	meta Meta
}

func (s sexp) ID() uint64 {
	return s.id
}

func (s sexp) Meta() Meta {
	return s.meta
}

func (s sexp) UnderlyingType() Type {
	panic("S-Expression is invalid")
}

func (s sexp) IsNil() bool {
	return false
}

func (s sexp) IsAtom() bool {
	return true
}

func (s sexp) StringVal() (string, bool) {
	return "", false
}

func (s sexp) Int64Val() (int64, bool) {
	return 0, false
}

func (s sexp) BoolVal() (bool, bool) {
	return false, false
}

func (s sexp) SymbolVal() (string, bool) {
	return "", false
}

func (s sexp) MustStringVal() string {
	panic("S-Expression is not a string")
}

func (s sexp) MustInt64Val() int64 {
	panic("S-Expression is not a 64-bit integer")
}

func (s sexp) MustBoolVal() bool {
	panic("S-Expression is not a boolean")
}

func (s sexp) MustSymbolVal() string {
	panic("S-Expression is not a symbol")
}

func (s sexp) Equals(SExp) bool {
	return false
}

func (s sexp) Car() SExp {
	return Nil()
}

func (s sexp) Cdr() SExp {
	return Nil()
}

func (s sexp) Add(SExp) SExp {
	return Nil()
}

// atomString is a string atom.
type atomString struct {
	sexp
	val string
}

// String creates a new string atom.
func String(val string, meta Meta) SExp {
	return atomString{sexp{CellID(), meta}, val}
}

func (a atomString) String() string {
	return strconv.Quote(a.val)
}

func (a atomString) UnderlyingType() Type {
	return TypeString
}

func (a atomString) StringVal() (string, bool) {
	return a.MustStringVal(), true
}

func (a atomString) MustStringVal() string {
	return a.val
}

func (a atomString) Equals(exp SExp) bool {
	v, ok := exp.StringVal()

	return ok && v == a.val
}

// atomInt64 is a 64-bit integer atom.
type atomInt64 struct {
	sexp
	val int64
}

// Int64 creates a new 64-bit integer atom.
func Int64(val int64, meta Meta) SExp {
	return atomInt64{sexp{CellID(), meta}, val}
}

func (a atomInt64) String() string {
	return fmt.Sprint(a.val)
}

func (a atomInt64) UnderlyingType() Type {
	return TypeInt64
}

func (a atomInt64) Int64Val() (int64, bool) {
	return a.MustInt64Val(), true
}

func (a atomInt64) MustInt64Val() int64 {
	return a.val
}

func (a atomInt64) Equals(exp SExp) bool {
	v, ok := exp.Int64Val()

	return ok && v == a.val
}

// atomBool is a boolean atom.
type atomBool struct {
	sexp
	val bool
}

// Bool creates a new boolean atom.
func Bool(val bool, meta Meta) SExp {
	return atomBool{sexp{CellID(), meta}, val}
}

func (a atomBool) String() string {
	return fmt.Sprint(a.val)
}

func (a atomBool) UnderlyingType() Type {
	return TypeBool
}

func (a atomBool) BoolVal() (bool, bool) {
	return a.MustBoolVal(), true
}

func (a atomBool) MustBoolVal() bool {
	return a.val
}

func (a atomBool) Equals(exp SExp) bool {
	v, ok := exp.BoolVal()

	return ok && v == a.val
}

// atomSymbol is a symbol atom.
type atomSymbol struct {
	sexp
	val string
}

// Symbol creates a new symbol atom.
func Symbol(val string, meta Meta) SExp {
	return atomSymbol{sexp{CellID(), meta}, val}
}

func (a atomSymbol) String() string {
	return (a.val)
}

func (a atomSymbol) UnderlyingType() Type {
	return TypeSymbol
}

func (a atomSymbol) SymbolVal() (string, bool) {
	return a.MustSymbolVal(), true
}

func (a atomSymbol) MustSymbolVal() string {
	return a.val
}

func (a atomSymbol) Equals(exp SExp) bool {
	v, ok := exp.SymbolVal()

	return ok && v == a.val
}

// scell is a cell S-Expression.
type scell struct {
	sexp
	car SExp
	cdr SExp
}

// Cons creates a new cell.
func Cons(car, cdr SExp, meta Meta) SExp {
	if car == nil {
		car = Nil()
	}
	if cdr == nil {
		cdr = Nil()
	}

	return &scell{sexp{CellID(), meta}, car, cdr}
}

func (c *scell) String() string {
	if c == nilVal {
		return "()"
	}

	if !IsList(c) {
		return fmt.Sprintf("(%v . %v)", c.car, c.cdr)
	}

	var buf bytes.Buffer

	buf.WriteRune('(')
	buf.WriteString(fmt.Sprint(c.car))

	for tail := c.Cdr(); tail != Nil(); tail = tail.Cdr() {
		buf.WriteRune(' ')
		buf.WriteString(fmt.Sprint(tail.Car()))
	}

	buf.WriteRune(')')

	return buf.String()
}

func (c *scell) UnderlyingType() Type {
	return TypeCell
}

func (c *scell) IsNil() bool {
	return c == nilVal
}

func (c *scell) IsAtom() bool {
	return c == nilVal
}

func (c *scell) Equals(exp SExp) bool {
	if c.IsNil() != exp.IsNil() {
		return false
	}

	v, ok := exp.(*scell)
	if !ok {
		return false
	}

	return c.car == v.car && c.cdr == v.cdr
}

func (c *scell) Car() SExp {
	return c.car
}

func (c *scell) Cdr() SExp {
	return c.cdr
}

// IsList returns whether an S-Expression is a list.
//
//	- nil is a list
//	- other atoms are not a list
//	- the cdr must be a list
func IsList(exp SExp) bool {
	if exp.IsNil() {
		return true
	}

	if exp.IsAtom() {
		return false
	}

	return IsList(exp.Cdr())
}

// ToSlice converts an S-Expression to a slice.
//
// The returned slice will contain the car values of all the expressions in the
// list. It returns nil if the expression is nil or an atom.
func ToSlice(exp SExp) SExpSlice {
	if exp.IsNil() || exp.IsAtom() {
		return nil
	}

	slice := SExpSlice{exp.Car()}

	for tail := exp.Cdr(); ; tail = tail.Cdr() {
		if tail.IsNil() {
			break
		}

		if tail.IsAtom() {
			return nil
		}

		slice = append(slice, tail.Car())
	}

	return slice
}
