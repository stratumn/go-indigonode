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
	errs []string
}

var typeTests = []typeTest{{
	"unknown function",
	"echo true",
	[]string{
		"1:1: unknown function",
	},
}, {
	"equals 1",
	"= 1 2",
	nil,
}, {
	"equals 2",
	"= 1 true",
	[]string{
		"1:5: unexpected type bool, want int64",
	},
}, {
	"equals 3",
	`= 1 true "a"`,
	[]string{
		"1:5: unexpected type bool, want int64",
		"1:10: unexpected type string, want int64",
	},
}, {
	"equals 4",
	"=",
	[]string{
		"1:1: missing function argument",
	},
}, {
	"addition 1",
	"+ 1 2",
	nil,
}, {
	"addition 2",
	"+ 1 true",
	[]string{
		"1:5: unexpected type bool, want int64",
	},
}, {
	"addition 3",
	"+",
	[]string{
		"1:1: missing function argument",
	},
}, {
	"not 1",
	"not true",
	nil,
}, {
	"not 2",
	"not",
	[]string{
		"1:1: missing function argument",
	},
}, {
	"not 3",
	"not true false",
	[]string{
		"1:10: extra function argument",
	},
}, {
	"and 1",
	"and true",
	nil,
}, {
	"and 2",
	"and 1 true",
	[]string{
		"1:5: unexpected type int64, want bool",
	},
}, {
	"and 3",
	"and",
	[]string{
		"1:1: missing function argument",
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
	tc.Check(exp)

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
