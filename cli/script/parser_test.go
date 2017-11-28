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
	"one two",
	"(one . (two . <nil>))",
	"",
}, {
	"one two three",
	"(one . (two . (three . <nil>)))",
	"",
}, {
	"one\r(two three)",
	"(one . ((two . (three . <nil>)) . <nil>))",
	"",
}, {
	"one (two)",
	"(one . ((two . <nil>) . <nil>))",
	"",
}, {
	"one (two three) four",
	"(one . ((two . (three . <nil>)) . (four . <nil>)))",
	"",
}, {
	"(one two)",
	"(one . (two . <nil>))",
	"",
}, {
	"one )",
	"",
	"1:5: unexpected token )",
}, {
	"(one\ttwo",
	"",
	"1:9: unexpected token <EOF>",
}, {
	"(one\ntwo",
	"",
	"2:4: unexpected token <EOF>",
}, {
	"((one two) three)",
	"",
	"1:2: unexpected token (",
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
