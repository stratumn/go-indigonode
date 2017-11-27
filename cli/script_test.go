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

package cli

import (
	"reflect"
	"strings"
	"testing"

	"github.com/pkg/errors"
)

type scannerTest struct {
	input  string
	tokens []Token
	errors []string
}

var scannerTT = []scannerTest{{
	"",
	[]Token{
		{TokEOF, "", 1, 1},
	},
	nil,
}, {
	"(hello world)",
	[]Token{
		{TokLParen, "", 1, 1},
		{TokString, "hello", 1, 2},
		{TokString, "world", 1, 8},
		{TokRParen, "", 1, 13},
		{TokEOF, "", 1, 14},
	},
	nil,
}, {
	`"hello world"`,
	[]Token{
		{TokString, "hello world", 1, 1},
		{TokEOF, "", 1, 14},
	},
	nil,
}, {
	"'hello\\tworld\\n\\r\\''",
	[]Token{
		{TokString, "hello\tworld\n\r'", 1, 1},
		{TokEOF, "", 1, 21},
	},
	nil,
}, {
	"'hello world",
	[]Token{
		{TokInvalid, "hello world", 1, 1},
	},
	[]string{"1;13: unexpected character '\\x00'"},
}, {
	"'hello \\\"'",
	[]Token{
		{TokInvalid, "hello ", 1, 1},
	},
	[]string{"1;9: unexpected character '\"'"},
}}

func TestScanner_Emit(t *testing.T) {
	for _, test := range scannerTT {
		var tokens []Token
		var errors []string

		errHandler := func(err ScannerError) {
			errors = append(errors, err.Error())
		}

		s := NewScanner(ScannerOptErrorHandler(errHandler))
		s.SetInput(test.input)

		for {
			tok := s.Emit()

			tokens = append(tokens, tok)

			if tok.Type == TokEOF || tok.Type == TokInvalid {
				break
			}
		}

		if !reflect.DeepEqual(tokens, test.tokens) {
			t.Errorf("%q: tokens = %v want %v", test.input, tokens, test.tokens)
		}

		if !reflect.DeepEqual(errors, test.errors) {
			t.Errorf("%q: errors = %q want %q", test.input, errors, test.errors)
		}
	}
}

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
	"1;5: unexpected token )",
}, {
	"(one\ttwo",
	"",
	"1;9: unexpected token <EOF>",
}, {
	"(one\ntwo",
	"",
	"2;4: unexpected token <EOF>",
}, {
	"((one two) three)",
	"",
	"1;2: unexpected token (",
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

type evalTest struct {
	input  string
	output string
	err    string
}

var evalTT = []evalTest{{
	"echo hello",
	"hello",
	"",
}, {
	"echo echo hello",
	"echo hello",
	"",
}, {
	"(echo hello)",
	"hello",
	"",
}, {
	"(echo hello world)",
	"hello world",
	"",
}, {
	"(echo 'hello  world')",
	"hello  world",
	"",
}, {
	"(echo hello (echo world !))",
	"hello world !",
	"",
}, {
	"(echo (echo the world) (echo is beautiful) !)",
	"the world is beautiful !",
	"",
}, {
	"(+ 1 2)",
	"",
	"1;2: unknown function \"+\"",
}, {
	"(echo (+ 1 2))",
	"",
	"1;8: unknown function \"+\"",
}}

func testEvaluator(name string, line int, offset int, args ...*SExp) (string, error) {
	if name == "echo" {
		vals, err := EvalMulti(testEvaluator, args...)
		if err != nil {
			return "", err
		}

		return strings.Join(vals, " "), nil
	}

	return "", errors.Errorf("%d;%d: unknown function %q", line, offset, name)
}

func TestSExp_Eval(t *testing.T) {
	s := NewScanner()
	p := NewParser(s)

	for _, test := range evalTT {
		sexp, err := p.Parse(test.input)
		if err != nil {
			t.Errorf("%q: error: %s", test.input, err)
			continue
		}

		output, err := sexp.Eval(testEvaluator)
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

		if output != test.output {
			t.Errorf("%q: output = %q want %q", test.input, output, test.output)
		}
	}
}
