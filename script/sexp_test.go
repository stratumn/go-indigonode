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
	"strings"
	"testing"
)

func TestTypeCell_String(t *testing.T) {
	if got, want := TypeCell.String(), "cell"; got != want {
		t.Errorf("String() = %s want %s", got, want)
	}
}

type sexpTest struct {
	sexp     SExp
	wantType Type
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
	wantType: TypeString,
	wantVal:  "test",
}, {
	sexp:     Int64(42, Meta{}),
	wantType: TypeInt64,
	wantVal:  int64(42),
}, {
	sexp:     Bool(true, Meta{}),
	wantType: TypeBool,
	wantVal:  true,
}, {
	sexp:     Symbol("test", Meta{}),
	wantType: TypeSymbol,
	wantVal:  "test",
}, {
	sexp:     cellTest1,
	wantType: TypeCell,
	wantVal:  cellTest1,
}, {
	sexp:     cellTest2,
	wantType: TypeCell,
	wantVal:  cellTest2,
}, {
	sexp:     cellTest3,
	wantType: TypeCell,
	wantVal:  cellTest3,
}, {
	sexp:     cellTest4,
	wantType: TypeCell,
	wantVal:  cellTest4,
}, {
	sexp:     cellTest5,
	wantType: TypeCell,
	wantVal:  cellTest5,
}}

func getVal(s SExp, typ Type) (interface{}, bool) {
	switch typ {
	case TypeString:
		return s.StringVal()
	case TypeInt64:
		return s.Int64Val()
	case TypeBool:
		return s.BoolVal()
	case TypeSymbol:
		return s.SymbolVal()
	case TypeCell:
		return s, !s.IsAtom()
	}

	return Nil(), false
}

func testSExp(t *testing.T, i int, test sexpTest) {
	var gotVal interface{}
	var ok bool

	s := test.sexp

	if got, want := s.UnderlyingType(), test.wantType; got != want {
		t.Errorf("UnderlyingType() = %s want %s", got, want)
	}

	for typ, name := range typeMap {
		if typ == TypeInvalid {
			continue
		}

		tname := strings.Title(name)
		gotVal, ok = getVal(s, typ)
		if test.wantType == typ {
			if wantVal := test.wantVal; gotVal != wantVal {
				t.Errorf(
					"%sVal(): val = %v want %v",
					tname,
					gotVal,
					wantVal,
				)
			}
			if !ok {
				t.Errorf("%sVal(): ok = %v want %v", tname, ok, true)
			}
		} else if ok {
			t.Errorf("%sVal(): ok = %v want %v", tname, ok, false)
		}

		if got, want := s.Equals(Nil()), s.IsNil(); got != want {
			fmt.Println(s)
			t.Errorf("s.Equals(()) = %v want %v", got, want)
		}

		for j, tt := range sexpTests {
			if got, want := s.Equals(tt.sexp), i == j; got != want {
				t.Errorf("s.Equals(%v) = %v want %v", tt.sexp, got, want)
			}
		}
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

	if got, want := s.Meta().Line, 10; got != want {
		t.Errorf("meta = %v want %v", got, want)
	}
}

func TestCell(t *testing.T) {
	// Pair.
	c := Cons(Symbol("sym", Meta{}), String("abc", Meta{}), Meta{})

	if got, want := c.Car().MustSymbolVal(), "sym"; got != want {
		t.Errorf("car = %v want %v", got, want)
	}
	if got, want := c.Cdr().MustStringVal(), "abc"; got != want {
		t.Errorf("car = %q want %q", got, want)
	}
	if got, want := IsList(c), false; got != want {
		t.Errorf("list = %v want %v", got, want)
	}
	if slice := ToSlice(c); slice != nil {
		t.Error("ToSlice() should return <nil>")
	}

	// List.
	c = Cons(
		Symbol("a", Meta{}),
		Cons(Symbol("b", Meta{}), nil, Meta{}),
		Meta{},
	)

	if got, want := c.Car().MustSymbolVal(), "a"; got != want {
		t.Errorf("car = %v want %v", got, want)
	}
	cdr := c.Cdr()
	if got, want := cdr.Car().MustSymbolVal(), "b"; got != want {
		t.Errorf("cdar = %v want %v", got, want)
	}
	if got := cdr.Cdr().IsNil(); !got {
		t.Errorf("cddr = %v want %v", got, true)
	}
	if got, want := IsList(c), true; got != want {
		t.Errorf("list = %v want %v", got, want)
	}
	slice := ToSlice(c)
	if slice == nil {
		t.Error("ToSlice() is nil")
	}
	if got, want := len(slice), 2; got != want {
		t.Errorf("length = %v want %v", got, want)
	}
	if got, want := slice[0].MustSymbolVal(), "a"; got != want {
		t.Errorf("slice[0] = %v want %v", got, want)
	}
	if got, want := slice[1].MustSymbolVal(), "b"; got != want {
		t.Errorf("slice[1] = %v want %v", got, want)
	}
}
