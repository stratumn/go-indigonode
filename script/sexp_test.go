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
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTypeCell_String(t *testing.T) {
	assert.Equal(t, "cell", SExpCell.String())
}

type sexpTest struct {
	sexp     SExp
	wantType SExpType
	wantVal  interface{}
}

var (
	cellTest1 = Cons(nil, nil, Meta{})
	cellTest2 = Cons(Int64(0, Meta{}), nil, Meta{})
	cellTest3 = Cons(Int64(1, Meta{}), nil, Meta{})
	cellTest4 = Cons(nil, Int64(0, Meta{}), Meta{})
	cellTest5 = Cons(nil, Int64(1, Meta{}), Meta{})
)

var sexpTests = [...]sexpTest{{
	sexp:     String("test", Meta{}),
	wantType: SExpString,
	wantVal:  "test",
}, {
	sexp:     Int64(42, Meta{}),
	wantType: SExpInt64,
	wantVal:  int64(42),
}, {
	sexp:     Bool(true, Meta{}),
	wantType: SExpBool,
	wantVal:  true,
}, {
	sexp:     Symbol("test", Meta{}),
	wantType: SExpSymbol,
	wantVal:  "test",
}, {
	sexp:     cellTest1,
	wantType: SExpCell,
	wantVal:  cellTest1,
}, {
	sexp:     cellTest2,
	wantType: SExpCell,
	wantVal:  cellTest2,
}, {
	sexp:     cellTest3,
	wantType: SExpCell,
	wantVal:  cellTest3,
}, {
	sexp:     cellTest4,
	wantType: SExpCell,
	wantVal:  cellTest4,
}, {
	sexp:     cellTest5,
	wantType: SExpCell,
	wantVal:  cellTest5,
}}

func getVal(s SExp, typ SExpType) (interface{}, bool) {
	switch typ {
	case SExpString:
		return s.StringVal()
	case SExpInt64:
		return s.Int64Val()
	case SExpBool:
		return s.BoolVal()
	case SExpSymbol:
		return s.SymbolVal()
	case SExpCell:
		return s, !s.IsAtom()
	}

	return Nil(), false
}

func testSExp(t *testing.T, i int, test sexpTest) {
	var gotVal interface{}
	var ok bool

	s := test.sexp

	assert.Equal(t, test.wantType, s.UnderlyingType())

	for typ, name := range sexpTypeMap {
		t.Run(strings.Title(name), func(t *testing.T) {
			if typ == SExpInvalid {
				return
			}

			gotVal, ok = getVal(s, typ)
			if test.wantType == typ {
				assert.True(t, ok)
				assert.Equal(t, test.wantVal, gotVal)
			} else {
				assert.False(t, ok)
			}

			assert.Equal(t, s.IsNil(), s.Equals(Nil()))

			for j, tt := range sexpTests {
				assert.Equalf(t, i == j, s.Equals(tt.sexp), "%v", tt.sexp)
			}
		})
	}
}

func TestSExp(t *testing.T) {
	for i, tt := range sexpTests {
		name := fmt.Sprint(tt.wantType)
		t.Run(name, func(t *testing.T) { testSExp(t, i, tt) })
	}
}

func TestSExp_Meta(t *testing.T) {
	s := String("", Meta{Line: 10})
	assert.Equal(t, 10, s.Meta().Line, "s.Meta().Line")
}

func TestCell(t *testing.T) {
	// Pair.
	c := Cons(Symbol("sym", Meta{}), String("abc", Meta{}), Meta{})

	assert.Equal(t, "sym", c.Car().MustSymbolVal(), "c.Car().MustSymbolVal()")
	assert.Equal(t, "abc", c.Cdr().MustStringVal(), "c.Cdr().MustStringVal()")
	assert.False(t, IsList(c), "IsList(c)")
	assert.Nil(t, ToSlice(c), "ToSlice(c)")

	// List.
	c = Cons(
		Symbol("a", Meta{}),
		Cons(Symbol("b", Meta{}), nil, Meta{}),
		Meta{},
	)

	assert.Equal(t, "a", c.Car().MustSymbolVal(), "c.Car().MustSymbolVal()")
	cdr := c.Cdr()
	assert.Equal(t, "b", cdr.Car().MustSymbolVal(), "cdr.Car().MustSymbolVal()")
	assert.True(t, cdr.Cdr().IsNil(), "cdr.Cdr().IsNil()")
	assert.True(t, IsList(c), "IsList(c)")

	slice := ToSlice(c)
	assert.NotNil(t, slice, "ToSlice(c)")
	assert.Len(t, slice, 2, "ToSlice(c)")
	assert.Equal(t, "a", slice[0].MustSymbolVal(), "slice[0].MustSymbolVal()")
	assert.Equal(t, "b", slice[1].MustSymbolVal(), "slice[1].MustSymbolVal()")
}
