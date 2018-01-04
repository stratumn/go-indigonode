// Copyright © 2017-2018 Stratumn SAS
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
	"testing"

	"github.com/stretchr/testify/assert"
)

type parserTest struct {
	input string
	sexp  string
	err   string
}

var parseTests = []parserTest{{
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
	`(one
	two "three")`,
	"((one two \"three\"))",
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
	"one ()",
	"((one ()))",
	"",
}, {
	"echo (() one)",
	"((echo (() one)))",
	"",
}, {
	"one () two",
	"((one () two))",
	"",
}, {
	"one ((two))",
	"((one ((two))))",
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
	"(quote (true false))",
	"((quote (true false)))",
	"",
}, {
	"echo '(true false)",
	"((echo (quote (true false))))",
	"",
}, {
	"lambda (x) { echo one\necho two }",
	"((lambda (x) ((echo one) (echo two))))",
	"",
}, {
	`
	; Reverses a list recusively.
	let reverse (lambda (l) (
		; Define a nested recursive function with an accumulator.
		(let reverse-rec (lambda (l tail) (
			(if (nil? l)
				tail
				(reverse-rec (cdr l) (cons (car l) tail))))))
		; Start the recursion
		(reverse-rec l ())))
	
	reverse '(1 2 3 4 5 6 7 8 9 10)
	`,
	"((let reverse (lambda (l) ((let reverse-rec (lambda (l tail) ((if (nil? l) tail (reverse-rec (cdr l) (cons (car l) tail)))))) (reverse-rec l ())))) (reverse (quote (1 2 3 4 5 6 7 8 9 10))))",
	"",
}, {
	"()",
	"",
	"1:2: unexpected token )",
}, {
	`("one")`,
	"",
	"1:2: unexpected token <string>",
}, {
	"(true)",
	"",
	"1:2: unexpected token true",
}, {
	"(one) two",
	"",
	"1:7: unexpected token <sym>",
}, {
	"((one) two)",
	"",
	"1:2: unexpected token (",
}, {
	"(one\r) two",
	"",
	"2:3: unexpected token <sym>",
}, {
	`"one"`,
	"",
	"1:1: unexpected token <string>",
}, {
	"one )",
	"",
	"1:5: unexpected token )",
}, {
	`one "`,
	"",
	"1:5: unexpected token <invalid>",
}, {
	"one two )",
	"",
	"1:9: unexpected token )",
}, {
	"one\t(two",
	"",
	"1:9: unexpected token <EOF>",
}, {
	"one\n(two",
	"",
	"2:5: unexpected token <EOF>",
}, {
	"echo 999999999999999999999999999999999999999",
	"",
	`1:6: strconv.ParseInt: parsing "999999999999999999999999999999999999999": value out of range`,
}, {
	"(echo {)",
	"",
	"1:8: unexpected token )",
}}

var listTests = []parserTest{{
	"()",
	"()",
	"",
}, {
	"(a b c)",
	"(a b c)",
	"",
}, {
	"",
	"",
	"1:1: unexpected token <EOF>",
}, {
	"(",
	"",
	"1:2: unexpected token <EOF>",
}, {
	"(a b) (c d)",
	"",
	"1:7: unexpected token (",
}, {
	"(one",
	"",
	"1:5: unexpected token <EOF>",
}, {
	"(one two",
	"",
	"1:9: unexpected token <EOF>",
}}

func TestParser_Parse(t *testing.T) {
	s := NewScanner()
	p := NewParser(s)

	for _, tt := range parseTests {
		exp, err := p.Parse(tt.input)
		if err != nil {
			if tt.err != "" {
				assert.Equal(t, tt.err, err.Error())
			} else {
				assert.NoError(t, err)
			}
			continue
		}

		assert.Equal(t, tt.sexp, fmt.Sprint(exp))
	}
}

func TestParser_List(t *testing.T) {
	s := NewScanner()
	p := NewParser(s)

	for _, tt := range listTests {
		exp, err := p.List(tt.input)
		if err != nil {
			if tt.err != "" {
				assert.Equal(t, tt.err, err.Error())
			} else {
				assert.NoError(t, err)
			}
			continue
		}

		assert.Equal(t, tt.sexp, fmt.Sprint(exp))
	}
}
