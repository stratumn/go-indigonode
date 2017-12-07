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

var cellTest = Cons(nil, nil, Meta{})

var sexpTests = [...]sexpTest{{
	sexp:     String("test", Meta{}),
	wantType: TypeString,
	wantVal:  "test",
}, {
	sexp:     Int64(42, Meta{}),
	wantType: TypeInt64,
	wantVal:  int64(42),
}, {
	sexp:     Symbol("test", Meta{}),
	wantType: TypeSymbol,
	wantVal:  "test",
}, {
	sexp:     cellTest,
	wantType: TypeCell,
	wantVal:  cellTest.(SCell),
}}

func getVal(s SExp, typ Type) (interface{}, bool) {
	switch typ {
	case TypeString:
		return s.StringVal()
	case TypeInt64:
		return s.Int64Val()
	case TypeSymbol:
		return s.SymbolVal()
	case TypeCell:
		return s.CellVal()
	}

	return nil, false
}

func testSExp(t *testing.T, test sexpTest) {
	var gotVal interface{}
	var ok bool

	s := test.sexp

	if got, want := s.UnderlyingType(), test.wantType; got != want {
		t.Errorf("GetUnderlyingType() = %s want %s", got, want)
	}

	for typ, name := range typeMap {
		tname := strings.Title(name)
		gotVal, ok = getVal(s, typ)
		if test.wantType == typ {
			if wantVal := test.wantVal; gotVal != wantVal {
				t.Errorf(
					"Get%s(): val = %v want %v",
					tname,
					gotVal,
					wantVal,
				)
			}
			if !ok {
				t.Errorf("Get%s(): ok = %v want %v", tname, ok, true)
			}
		} else if ok {
			t.Errorf("Get%s(): ok = %v want %v", tname, ok, false)
		}
	}
}

func TestSExp(t *testing.T) {
	for _, tt := range sexpTests {
		name := fmt.Sprint(tt.wantType)
		t.Run(name, func(t *testing.T) { testSExp(t, tt) })
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
	c := Cons(
		Symbol("sym", Meta{}),
		String("abc", Meta{}),
		Meta{},
	).MustCellVal()

	if got, want := c.Car().MustSymbolVal(), "sym"; got != want {
		t.Errorf("car = %v want %v", got, want)
	}
	if got, want := c.Cdr().MustStringVal(), "abc"; got != want {
		t.Errorf("car = %q want %q", got, want)
	}
	if got, want := c.IsList(), false; got != want {
		t.Errorf("list = %v want %v", got, want)
	}
	if _, ok := c.ToSlice(); ok {
		t.Error("ToSlice() should return false")
	}

	// List.
	c = Cons(
		Symbol("a", Meta{}),
		Cons(
			Symbol("b", Meta{}),
			nil,
			Meta{},
		),
		Meta{},
	).MustCellVal()

	if got, want := c.Car().MustSymbolVal(), "a"; got != want {
		t.Errorf("car = %v want %v", got, want)
	}
	cdr := c.Cdr().MustCellVal()
	if got, want := cdr.Car().MustSymbolVal(), "b"; got != want {
		t.Errorf("cdar = %v want %v", got, want)
	}
	if got := cdr.Cdr(); got != nil {
		t.Errorf("cddr = %v want %v", got, nil)
	}
	if got, want := c.IsList(), true; got != want {
		t.Errorf("list = %v want %v", got, want)
	}
	list, ok := c.ToSlice()
	if !ok {
		t.Errorf("ToSlice(): ok = %v want %v", ok, true)
	}
	if got, want := len(list), 2; got != want {
		t.Errorf("length = %v want %v", got, want)
	}
	if got, want := list[0].MustSymbolVal(), "a"; got != want {
		t.Errorf("list[0] = %v want %v", got, want)
	}
	if got, want := list[1].MustSymbolVal(), "b"; got != want {
		t.Errorf("list[1] = %v want %v", got, want)
	}
}
