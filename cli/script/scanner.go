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

	"github.com/pkg/errors"
)

// TokenType is a script token.
type TokenType uint8

// Available tokens.
const (
	TokInvalid TokenType = iota
	TokEOF
	TokLine
	TokLParen
	TokRParen
	TokSym
	TokString
)

var tokToStr = map[TokenType]string{
	TokInvalid: "<invalid>",
	TokLine:    "<line>",
	TokEOF:     "<EOF>",
	TokLParen:  "(",
	TokRParen:  ")",
	TokSym:     "<sym>",
	TokString:  "<string>",
}

// String returns a string representation of a token type.
func (tok TokenType) String() string {
	return tokToStr[tok]
}

// Token represents a token.
type Token struct {
	Type   TokenType
	Value  string
	Line   int
	Offset int
}

// ScannerError represents an error from the scanner.
type ScannerError struct {
	Line   int
	Offset int
	Ch     rune
}

// Error returns an error string.
func (err ScannerError) Error() string {
	return fmt.Sprintf("%d:%d: unexpected character %q", err.Line, err.Offset, err.Ch)
}

// Scanner produces tokens from a string.
type Scanner struct {
	runes      []rune
	len        int
	line       int
	pos        int
	offset     int
	prevOffset int
	end        bool

	errHandler func(error)
}

// ScannerOpt is a scanner option.
type ScannerOpt func(*Scanner)

// ScannerOptErrorHandler sets the scanner's error handler.
func ScannerOptErrorHandler(h func(error)) ScannerOpt {
	return func(s *Scanner) {
		s.errHandler = h
	}
}

// NewScanner creates a new scanner.
func NewScanner(opts ...ScannerOpt) *Scanner {
	s := &Scanner{
		runes:      nil,
		len:        0,
		errHandler: func(error) {},
	}

	for _, o := range opts {
		o(s)
	}

	return s
}

// SetInput resets the scanner and sets its input string.
func (s *Scanner) SetInput(in string) {
	s.runes = []rune(in)
	s.len = len(s.runes)
	s.pos = 0
	s.line = 1
	s.offset = 0
	s.prevOffset = 0
	s.end = false
}

// Emit emits the next token.
//
// It returns TokEOF if all tokens have been read, and TokInvalid if no valid
// token was found.
func (s *Scanner) Emit() Token {
	c := s.read()

	for isSpace(c) || c == ';' {
		for isSpace(c) {
			c = s.read()
		}

		if c == ';' {
			c = s.stripComment()
		}
	}

	if isLine(c) {
		return Token{TokLine, "", s.line, s.offset}
	}

	switch c {
	case 0:
		return Token{TokEOF, "", s.line, s.offset}
	case '(':
		return Token{TokLParen, "", s.line, s.offset}
	case ')':
		return Token{TokRParen, "", s.line, s.offset}
	}

	return s.strOrSym(c)
}

func (s *Scanner) read() rune {
	s.prevOffset = s.offset

	if s.pos >= s.len {
		if !s.end {
			s.offset++
			s.end = true
		}

		return 0
	}

	c := s.runes[s.pos]

	s.pos++
	s.offset++

	if isLine(c) {
		s.line++
		s.offset = 0
	}

	return c
}

// back should be called at most once per Emit.
func (s *Scanner) back() {
	s.pos--
	s.offset--

	if isLine(s.runes[s.pos]) {
		s.line--
		s.offset = s.prevOffset
	}
}

func (s *Scanner) stripComment() rune {
	c := s.read()

	for !isLine(c) && c != 0 {
		c = s.read()
	}

	return c
}

func (s *Scanner) error(line, pos int, c rune) {
	s.errHandler(errors.WithStack(ScannerError{line, pos, c}))
}

func (s *Scanner) strOrSym(c rune) Token {
	switch c {
	case '"':
		return s.text(c, '"')
	case '\'':
		return s.text(c, '\'')
	}

	return s.text(c, 0)
}

func (s *Scanner) text(c rune, quote rune) Token {
	buf := ""

	line, offset := s.line, s.offset

	if c == quote {
		c = s.read()
	}

	for {
		switch {
		case c == 0 && quote == 0:
			return Token{TokSym, buf, line, offset}
		case c == 0:
			s.error(s.line, s.offset, 0)
			return Token{TokInvalid, buf, line, offset}
		case c == quote:
			return Token{TokString, buf, line, offset}
		case c == '\\' && quote != 0:
			c = s.read()
			switch c {
			case 'n':
				buf += "\n"
			case 'r':
				buf += "\r"
			case 't':
				buf += "\t"
			case quote:
				buf += string(c)
			default:
				s.error(s.line, s.offset, c)
				s.back()
				return Token{TokInvalid, buf, line, offset}
			}
			c = s.read()
		case isSpecial(c) && quote == 0:
			s.back()
			return Token{TokSym, buf, line, offset}
		default:
			buf += string(c)
			c = s.read()
		}
	}
}

func isSpace(c rune) bool {
	return c == ' ' || c == '\t'
}

func isLine(c rune) bool {
	return c == '\n' || c == '\r'
}

func isSpecial(c rune) bool {
	return c == '(' || c == ')' || c == ';' || isSpace(c) || isLine(c)
}
