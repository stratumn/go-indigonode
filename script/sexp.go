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
	"bytes"
	"fmt"
	"strconv"
	"sync/atomic"
)

// SExpType is an S-Expression type.
type SExpType uint8

// S-Expression types.
const (
	SExpInvalid SExpType = iota
	SExpString
	SExpInt64
	SExpBool
	SExpSymbol
	SExpCell
)

// Maps types to their names.
var sexpTypeMap = map[SExpType]string{
	SExpInvalid: "invalid",
	SExpString:  "string",
	SExpInt64:   "int64",
	SExpBool:    "bool",
	SExpSymbol:  "symbol",
	SExpCell:    "cell",
}

// String returns a string representation of the type.
func (t SExpType) String() string {
	return sexpTypeMap[t]
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

// SExp is a dynamic implementation of an S-Expression.
//
// See:
//
//	https://en.wikipedia.org/wiki/S-expression
type SExp interface {
	ID() uint64

	String() string

	Meta() Meta

	UnderlyingType() SExpType
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
var nilVal = &scell{sexp: sexp{id: 0, meta: Meta{}}, car: nil, cdr: nil}

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

func (s sexp) UnderlyingType() SExpType {
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
	return atomString{sexp: sexp{id: CellID(), meta: meta}, val: val}
}

func (a atomString) String() string {
	return strconv.Quote(a.val)
}

func (a atomString) UnderlyingType() SExpType {
	return SExpString
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
	return atomInt64{sexp: sexp{id: CellID(), meta: meta}, val: val}
}

func (a atomInt64) String() string {
	return fmt.Sprint(a.val)
}

func (a atomInt64) UnderlyingType() SExpType {
	return SExpInt64
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
	return atomBool{sexp: sexp{id: CellID(), meta: meta}, val: val}
}

func (a atomBool) String() string {
	return fmt.Sprint(a.val)
}

func (a atomBool) UnderlyingType() SExpType {
	return SExpBool
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
	return atomSymbol{sexp: sexp{id: CellID(), meta: meta}, val: val}
}

func (a atomSymbol) String() string {
	return (a.val)
}

func (a atomSymbol) UnderlyingType() SExpType {
	return SExpSymbol
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

	return &scell{sexp: sexp{id: CellID(), meta: meta}, car: car, cdr: cdr}
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

func (c *scell) UnderlyingType() SExpType {
	return SExpCell
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
