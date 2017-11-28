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
	"testing"
)

type parserTest struct {
	input string
	sexp  string
	err   string
}

var parserTT = []parserTest{{
	"",
	"()",
	"",
}, {
	"one",
	"((one))",
	"",
}, {
	"one two",
	"((one two))",
	"",
}, {
	"one two three",
	"((one two three))",
	"",
}, {
	"one\rtwo three",
	"((one) (two three))",
	"",
}, {
	"(one)",
	"((one))",
	"",
}, {
	"(one two)",
	"((one two))",
	"",
}, {
	"(one\n)",
	"((one))",
	"",
}, {
	"(one\rtwo three)",
	"((one two three))",
	"",
}, {
	"one\n two \rthree",
	"((one) (two) (three))",
	"",
}, {
	"one (two)",
	"((one (two)))",
	"",
}, {
	"one (two three) four",
	"((one (two three) four))",
	"",
}, {
	"one\n\ttwo (three)\n\tfour",
	"((one) (two (three)) (four))",
	"",
}, {
	"()",
	"",
	"1:2: unexpected token )",
}, {
	"(one) two",
	"",
	"1:7: unexpected token <string>",
}, {
	"(one\r) two",
	"",
	"2:3: unexpected token <string>",
}, {
	"one )",
	"",
	"1:5: unexpected token )",
}, {
	"one\t(two",
	"",
	"1:9: unexpected token <EOF>",
}, {
	"one ((two) three)",
	"",
	"1:6: unexpected token (",
}}

func TestParser_Parse(t *testing.T) {
	s := NewScanner()
	p := NewParser(s)

	for _, test := range parserTT {
		sexp, err := p.Parse(test.input)
		if err != nil {
			if test.err != "" {
				if err.Error() != test.err {
					t.Errorf("%q: error = %q want %q", test.input, err, test.err)
				}
			} else {
				t.Errorf("%q: error: %s", test.input, err)
			}
			continue
		}

		if got, want := sexp.String(), test.sexp; got != want {
			t.Errorf("%q: sexp = %q want %q", test.input, got, want)
		}
	}
}
