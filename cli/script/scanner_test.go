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
	"reflect"
	"testing"
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
	"(hello \nworld)",
	[]Token{
		{TokLParen, "", 1, 1},
		{TokString, "hello", 1, 2},
		{TokLine, "", 2, 0},
		{TokString, "world", 2, 1},
		{TokRParen, "", 2, 6},
		{TokEOF, "", 2, 7},
	},
	nil,
}, {
	"one\r(two three)",
	[]Token{
		{TokString, "one", 1, 1},
		{TokLine, "", 2, 0},
		{TokLParen, "", 2, 1},
		{TokString, "two", 2, 2},
		{TokString, "three", 2, 6},
		{TokRParen, "", 2, 11},
		{TokEOF, "", 2, 12},
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
	"hello;world",
	[]Token{
		{TokString, "hello", 1, 1},
		{TokEOF, "", 1, 12},
	},
	nil,
}, {
	"hello; world",
	[]Token{
		{TokString, "hello", 1, 1},
		{TokEOF, "", 1, 13},
	},
	nil,
}, {
	"hello ; world \n world",
	[]Token{
		{TokString, "hello", 1, 1},
		{TokLine, "", 2, 0},
		{TokString, "world", 2, 2},
		{TokEOF, "", 2, 7},
	},
	nil,
}, {
	"'hello world",
	[]Token{
		{TokInvalid, "hello world", 1, 1},
	},
	[]string{"1:13: unexpected character '\\x00'"},
}, {
	"'hello \\\"'",
	[]Token{
		{TokInvalid, "hello ", 1, 1},
	},
	[]string{"1:9: unexpected character '\"'"},
}}

func TestScanner_Emit(t *testing.T) {
	for _, test := range scannerTT {
		var tokens []Token
		var errors []string

		errHandler := func(err error) {
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
