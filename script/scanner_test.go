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
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

type scanTest struct {
	input  string
	tokens []Token
	errors []string
}

var scanTests = []scanTest{{
	"",
	[]Token{
		{TokEOF, "", 1, 1},
	},
	nil,
}, {
	"(hello \nworld)",
	[]Token{
		{TokLParen, "", 1, 1},
		{TokSymbol, "hello", 1, 2},
		{TokLine, "", 2, 0},
		{TokSymbol, "world", 2, 1},
		{TokRParen, "", 2, 6},
		{TokEOF, "", 2, 7},
	},
	nil,
}, {
	"one\r(two 3)",
	[]Token{
		{TokSymbol, "one", 1, 1},
		{TokLine, "", 2, 0},
		{TokLParen, "", 2, 1},
		{TokSymbol, "two", 2, 2},
		{TokInt, "3", 2, 6},
		{TokRParen, "", 2, 7},
		{TokEOF, "", 2, 8},
	},
	nil,
}, {
	"a 'b `c ,d",
	[]Token{
		{TokSymbol, "a", 1, 1},
		{TokQuote, "", 1, 3},
		{TokSymbol, "b", 1, 4},
		{TokQuasiquote, "", 1, 6},
		{TokSymbol, "c", 1, 7},
		{TokUnquote, "", 1, 9},
		{TokSymbol, "d", 1, 10},
		{TokEOF, "", 1, 11},
	},
	nil,
}, {
	"(hello 42 world)",
	[]Token{
		{TokLParen, "", 1, 1},
		{TokSymbol, "hello", 1, 2},
		{TokInt, "42", 1, 8},
		{TokSymbol, "world", 1, 11},
		{TokRParen, "", 1, 16},
		{TokEOF, "", 1, 17},
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
	"\"hello\\tworld\\n\\r\\\"\"",
	[]Token{
		{TokString, "hello\tworld\n\r\"", 1, 1},
		{TokEOF, "", 1, 21},
	},
	nil,
}, {
	"hello;world",
	[]Token{
		{TokSymbol, "hello", 1, 1},
		{TokEOF, "", 1, 12},
	},
	nil,
}, {
	"hello; world",
	[]Token{
		{TokSymbol, "hello", 1, 1},
		{TokEOF, "", 1, 13},
	},
	nil,
}, {
	"hello ; world \n world",
	[]Token{
		{TokSymbol, "hello", 1, 1},
		{TokLine, "", 2, 0},
		{TokSymbol, "world", 2, 2},
		{TokEOF, "", 2, 7},
	},
	nil,
}, {
	"♥ world",
	[]Token{
		{TokSymbol, "♥", 1, 1},
		{TokSymbol, "world", 1, 3},
		{TokEOF, "", 1, 8},
	},
	nil,
}, {
	"true",
	[]Token{
		{TokTrue, "", 1, 1},
		{TokEOF, "", 1, 5},
	},
	nil,
}, {
	"false",
	[]Token{
		{TokFalse, "", 1, 1},
		{TokEOF, "", 1, 6},
	},
	nil,
}, {
	"0(",
	[]Token{
		{TokInt, "0", 1, 1},
		{TokLParen, "", 1, 2},
		{TokEOF, "", 1, 3},
	},
	nil,
}, {
	"0",
	[]Token{
		{TokInt, "0", 1, 1},
		{TokEOF, "", 1, 2},
	},
	nil,
}, {
	"-0",
	[]Token{
		{TokInt, "-0", 1, 1},
		{TokEOF, "", 1, 3},
	},
	nil,
}, {
	"07",
	[]Token{
		{TokInt, "07", 1, 1},
		{TokEOF, "", 1, 3},
	},
	nil,
}, {
	"-07",
	[]Token{
		{TokInt, "-07", 1, 1},
		{TokEOF, "", 1, 4},
	},
	nil,
}, {
	"0x0faF",
	[]Token{
		{TokInt, "0x0faF", 1, 1},
		{TokEOF, "", 1, 7},
	},
	nil,
}, {
	"0xFF)",
	[]Token{
		{TokInt, "0xFF", 1, 1},
		{TokRParen, "", 1, 5},
		{TokEOF, "", 1, 6},
	},
	nil,
}, {
	"-",
	[]Token{
		{TokSymbol, "-", 1, 1},
		{TokEOF, "", 1, 2},
	},
	nil,
}, {
	"-a",
	[]Token{
		{TokSymbol, "-a", 1, 1},
		{TokEOF, "", 1, 3},
	},
	nil,
}, {
	"0xFQ",
	[]Token{
		{TokSymbol, "0xFQ", 1, 1},
		{TokEOF, "", 1, 5},
	},
	nil,
}, {
	"-0xFF",
	[]Token{
		{TokSymbol, "-0xFF", 1, 1},
		{TokEOF, "", 1, 6},
	},
	nil,
}, {
	"028",
	[]Token{
		{TokSymbol, "028", 1, 1},
		{TokEOF, "", 1, 4},
	},
	nil,
}, {
	"0xFQ",
	[]Token{
		{TokSymbol, "0xFQ", 1, 1},
		{TokEOF, "", 1, 5},
	},
	nil,
}, {
	`"hello world`,
	[]Token{
		{TokInvalid, "hello world", 1, 1},
	},
	[]string{`1:13: unexpected character '\x00'`},
}, {
	`hello"`,
	[]Token{
		{TokInvalid, "hello", 1, 1},
	},
	[]string{`1:6: unexpected character '"'`},
}, {
	`"hello \a`,
	[]Token{
		{TokInvalid, "hello ", 1, 1},
	},
	[]string{`1:9: unexpected character 'a'`},
}, {
	"[",
	[]Token{
		{TokInvalid, "", 1, 1},
	},
	[]string{"1:1: unexpected character '['"},
}, {
	"hello[",
	[]Token{
		{TokInvalid, "hello", 1, 1},
	},
	[]string{"1:6: unexpected character '['"},
}, {
	"42[",
	[]Token{
		{TokInvalid, "42", 1, 1},
	},
	[]string{"1:3: unexpected character '['"},
}, {
	"0x42[",
	[]Token{
		{TokInvalid, "0x42", 1, 1},
	},
	[]string{"1:5: unexpected character '['"},
}}

func TestScanner_Emit(t *testing.T) {
	for _, tt := range scanTests {
		t.Run(tt.input, func(t *testing.T) {
			var tokens []Token
			var errors []string

			errHandler := func(err error) {
				errors = append(errors, err.Error())
			}

			s := NewScanner(ScannerOptErrorHandler(errHandler))
			s.SetInput(tt.input)

			for {
				tok := s.Emit()

				tokens = append(tokens, tok)

				if tok.Type == TokEOF || tok.Type == TokInvalid {
					break
				}
			}

			assert.EqualValues(t, tt.tokens, tokens)
			assert.EqualValues(t, tt.errors, errors)
		})
	}
}

func TestScanner_invalidUTF8(t *testing.T) {
	var got error

	s := NewScanner(ScannerOptErrorHandler(func(err error) {
		got = errors.Cause(err)
	}))

	s.SetInput(string([]byte{0xff, 0xfe, 0xfd}))
	s.Emit()

	assert.Equal(t, ErrInvalidUTF8, got)
}
