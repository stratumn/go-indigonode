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
)

// ResolveHandler resolves symbols.
type ResolveHandler func(sym SExp) (SExp, error)

// ResolveName resolves symbols with their names.
func ResolveName(sym SExp) (SExp, error) {
	return String(sym.MustSymbolVal(), sym.Meta()), nil
}

// Type is an SExp type.
type Type uint8

// SExp types.
const (
	TypeInvalid Type = iota
	TypeString
	TypeInt64
	TypeBool
	TypeSymbol
	TypeCell
)

// Maps types to their names.
var typeMap = map[Type]string{
	TypeInvalid: "invalid",
	TypeString:  "string",
	TypeInt64:   "int64",
	TypeBool:    "bool",
	TypeSymbol:  "symbol",
	TypeCell:    "cell",
}

// String returns a string representation of the type.
func (t Type) String() string {
	return typeMap[t]
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
	String() string

	Meta() Meta

	UnderlyingType() Type

	StringVal() (string, bool)
	Int64Val() (int64, bool)
	BoolVal() (bool, bool)
	SymbolVal() (string, bool)
	CellVal() (SCell, bool)

	MustStringVal() string
	MustInt64Val() int64
	MustSymbolVal() string
	MustBoolVal() bool
	MustCellVal() SCell
}

// SExpSlice is a slice of S-Expressions.
type SExpSlice []SExp

// Strings returns the strings value of every S-Expression in the slice.
//
// If quote is true, strings are quoted.
func (s SExpSlice) Strings(quote bool) []string {
	strings := make([]string, len(s))
	for i, exp := range s {
		if !quote && exp != nil {
			if str, ok := exp.StringVal(); ok {
				strings[i] = str
				continue
			}
		}
		strings[i] = fmt.Sprint(exp)
	}

	return strings
}

// SCell represents a cell S-Expression.
type SCell interface {
	SExp
	Car() SExp
	Cdr() SExp
	IsList() bool
	ToSlice() (SExpSlice, bool)
	MustToSlice() SExpSlice
}

// sexp is the base S-Expression type.
type sexp struct {
	meta Meta
}

func (s sexp) Meta() Meta {
	return s.meta
}

func (s sexp) UnderlyingType() Type {
	panic("S-Expression is invalid")
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

func (s sexp) CellVal() (SCell, bool) {
	return nil, false
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

func (s sexp) MustCellVal() SCell {
	panic("S-Expression is not a cell")
}

// atomString is a string atom.
type atomString struct {
	sexp
	val string
}

// String creates a new string atom.
func String(val string, meta Meta) SExp {
	return atomString{sexp{meta}, val}
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

// atomInt64 is a 64-bit integer atom.
type atomInt64 struct {
	sexp
	val int64
}

// Int64 creates a new 64-bit integer atom.
func Int64(val int64, meta Meta) SExp {
	return atomInt64{sexp{meta}, val}
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

// atomBool is a boolean atom.
type atomBool struct {
	sexp
	val bool
}

// Bool creates a new boolean atom.
func Bool(val bool, meta Meta) SExp {
	return atomBool{sexp{meta}, val}
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

// atomSymbol is a symbol atom.
type atomSymbol struct {
	sexp
	val string
}

// Symbol creates a new symbol atom.
func Symbol(val string, meta Meta) SExp {
	return atomSymbol{sexp{meta}, val}
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

// scell is a cell S-Expression.
type scell struct {
	sexp
	car SExp
	cdr SExp
}

// Cons creates a new cell.
func Cons(car, cdr SExp, meta Meta) SCell {
	return &scell{sexp{meta}, car, cdr}
}

func (c scell) String() string {
	if c.IsList() {
		var buf bytes.Buffer

		buf.WriteRune('(')
		buf.WriteString(fmt.Sprint(c.car))
		cdr := c.cdr

		for cdr != nil {
			buf.WriteRune(' ')
			cell := cdr.MustCellVal()
			buf.WriteString(fmt.Sprint(cell.Car()))
			cdr = cell.Cdr()
		}

		buf.WriteRune(')')

		return buf.String()
	}

	return fmt.Sprintf("(%v . %v)", c.car, c.cdr)
}

func (c *scell) UnderlyingType() Type {
	return TypeCell
}

func (c *scell) CellVal() (SCell, bool) {
	return c.MustCellVal(), true
}

func (c *scell) MustCellVal() SCell {
	return c
}

func (c *scell) Car() SExp {
	return c.car
}

func (c *scell) Cdr() SExp {
	return c.cdr
}

func (c *scell) IsList() bool {
	cdr := c.cdr
	if cdr == nil {
		return true
	}

	stype := cdr.UnderlyingType()
	if stype != TypeCell {
		return false
	}

	return cdr.MustCellVal().IsList()
}

func (c *scell) ToSlice() (SExpSlice, bool) {
	if c.IsList() {
		return c.MustToSlice(), true
	}

	return nil, false
}

func (c *scell) MustToSlice() SExpSlice {
	slice := SExpSlice{c.car}
	cdr := c.cdr

	for cdr != nil {
		cell := cdr.MustCellVal()
		slice = append(slice, cell.Car())
		cdr = cell.Cdr()
	}

	return slice
}
