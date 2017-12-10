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

import "testing"

type typeTest struct {
	name string
	src  string
	typ  string
	errs []string
}

var typeTests = []typeTest{{
	"invalid call",
	"+ (true)",
	"int64",
	[]string{
		"1:4: invalid function call",
	},
}, {
	"unknown function",
	"echo true",
	"invalid",
	[]string{
		"1:1: echo: unknown function",
	},
}, {
	"equals 1",
	"= 1 2",
	"bool",
	nil,
}, {
	"equals 2",
	"= 1 true",
	"bool",
	[]string{
		"1:5: unexpected type bool, want int64",
	},
}, {
	"equals 3",
	`= 1 true "a"`,
	"bool",
	[]string{
		"1:5: unexpected type bool, want int64",
		"1:10: unexpected type string, want int64",
	},
}, {
	"equals 4",
	"=",
	"bool",
	[]string{
		"1:1: missing function argument",
	},
}, {
	"addition 1",
	"+ 1 2",
	"int64",
	nil,
}, {
	"addition 2",
	"+ 1 true",
	"int64",
	[]string{
		"1:5: unexpected type bool, want int64",
	},
}, {
	"addition 3",
	"+",
	"int64",
	[]string{
		"1:1: missing function argument",
	},
}, {
	"greater",
	"> 1 2",
	"bool",
	nil,
}, {
	"not 1",
	"not true",
	"bool",
	nil,
}, {
	"not 2",
	"not",
	"bool",
	[]string{
		"1:1: missing function argument",
	},
}, {
	"not 3",
	"not true false",
	"bool",
	[]string{
		"1:10: extra function argument",
	},
}, {
	"and 1",
	"and true",
	"bool",
	nil,
}, {
	"and 2",
	"and 1 true",
	"bool",
	[]string{
		"1:5: unexpected type int64, want bool",
	},
}, {
	"and 3",
	"and",
	"bool",
	[]string{
		"1:1: missing function argument",
	},
}, {
	"cons 1",
	"cons 1 ()",
	"list<int64>",
	nil,
}, {
	"cons 2",
	"cons 1 (cons 2 ())",
	"list<int64>",
	nil,
}, {
	"cons 3",
	"cons 1 (cons true ())",
	"list<any>",
	nil,
}, {
	"cons 4",
	"cons 1 true",
	"pair<int64,bool>",
	nil,
}, {
	"cons 5",
	"cons () 1",
	"invalid",
	[]string{
		"1:6: car cannot be nil",
	},
}, {
	"cons 7",
	"cons",
	"invalid",
	[]string{
		"1:1: missing function argument",
	},
}, {
	"cons 8",
	"cons 1",
	"invalid",
	[]string{
		"1:1: missing function argument",
	},
}, {
	"cons 9",
	"cons 1 2 3",
	"invalid",
	[]string{
		"1:10: extra function argument",
	},
}, {
	"cons 10",
	"cons 1 (+ 1 true)",
	"pair<int64,int64>",
	[]string{
		"1:13: unexpected type bool, want int64",
	},
}, {
	"cons 11",
	"cons 1 (cons)",
	"invalid",
	[]string{
		"1:9: missing function argument",
	},
}, {
	"car 1",
	"car ()",
	"nil",
	nil,
}, {
	"car 2",
	"car (cons 1 2)",
	"int64",
	nil,
}, {
	"car 3",
	"car (cons 1 (cons 2 ()))",
	"int64",
	nil,
}, {
	"car 4",
	"car",
	"invalid",
	[]string{
		"1:1: missing function argument",
	},
}, {
	"car 5",
	"car (cons 1 2) 3",
	"invalid",
	[]string{
		"1:16: extra function argument",
	},
}, {
	"car 6",
	"car (car)",
	"invalid",
	[]string{
		"1:6: missing function argument",
	},
}, {
	"car 7",
	"car 1",
	"invalid",
	[]string{
		"1:5: not a cell",
	},
}, {
	"cdr 1",
	"cdr ()",
	"nil",
	nil,
}, {
	"cdr 2",
	"cdr (cons 1 2)",
	"int64",
	nil,
}, {
	"cdr 3",
	"cdr (cons 1 ())",
	"int64",
	nil,
}, {
	"cdr 4",
	"cdr (cons 1 2) 3",
	"invalid",
	[]string{
		"1:16: extra function argument",
	},
}, {
	"cdr 5",
	"cdr (cdr)",
	"invalid",
	[]string{
		"1:6: missing function argument",
	},
}, {
	"cdr 6",
	"cdr 1",
	"invalid",
	[]string{
		"1:5: not a cell",
	},
}, {
	"let 1",
	"let a true",
	"bool",
	nil,
}, {
	"let 2",
	"let a (+ 1 2)",
	"int64",
	nil,
}, {
	"let 3",
	"let",
	"invalid",
	[]string{
		"1:1: missing function argument",
	},
}, {
	"let 4",
	"let a",
	"invalid",
	[]string{
		"1:1: missing function argument",
	},
}, {
	"let 5",
	"let true 1",
	"invalid",
	[]string{
		"1:5: not a symbol",
	},
}, {
	"let 6",
	"let a ()",
	"invalid",
	[]string{
		"1:7: value cannot be nil",
	},
}, {
	"let 7",
	"let a 3 4",
	"invalid",
	[]string{
		"1:9: extra function argument",
	},
}, {
	"let 8",
	"let a 3\nlet a true",
	"bool",
	[]string{
		"2:5: a: symbol is already bound locally",
	},
}, {
	"let 9",
	"let a (cons)",
	"invalid",
	[]string{
		"1:8: missing function argument",
	},
}, {
	"context 1",
	"let a 1\n+ 1 a",
	"int64",
	nil,
}, {
	"context 2",
	"let a 3\n+ 1 b",
	"int64",
	[]string{
		"2:5: b: symbol is not bound to a value",
	},
}}

func TestTypeChecker(t *testing.T) {
	for _, tt := range typeTests {
		t.Run(tt.name, func(t *testing.T) {
			testType(t, tt)
		})
	}
}

func testType(t *testing.T, tt typeTest) {
	var errs []string

	errHandler := func(err error) {
		errs = append(errs, err.Error())
	}

	parser := NewParser(NewScanner())

	exp, err := parser.Parse(tt.src)
	if err != nil {
		t.Fatalf("parser.Parse(tt.src): error: %s", err)
	}

	tc := NewTypeChecker(TypeCheckerOptErrorHandler(errHandler))

	ti := tc.BodyType(exp)
	if got, want := ti.String(), tt.typ; got != want {
		t.Errorf("type = %q want %q", got, want)
	}

	i := 0

	for ; i < len(errs) && i < len(tt.errs); i++ {
		if got, want := errs[i], tt.errs[i]; got != want {
			t.Errorf("error = %q want %q", got, want)
		}
	}

	for ; i < len(errs); i++ {
		t.Errorf("unexpected error: %q", errs[i])
	}

	for ; i < len(tt.errs); i++ {
		t.Errorf("missing error: %q", tt.errs[i])
	}
}
