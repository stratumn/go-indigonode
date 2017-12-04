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
	"io"
	"strings"
	"unicode/utf8"

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

// ScanError represents an error from the scanner.
type ScanError struct {
	Line   int
	Offset int
	Rune   rune
}

// Error returns an error string.
func (err ScanError) Error() string {
	return fmt.Sprintf("%d:%d: unexpected character %q", err.Line, err.Offset, err.Rune)
}

// Scanner produces tokens from a string.
type Scanner struct {
	reader     io.RuneScanner
	ch         rune
	prevCh     rune
	line       int
	pos        int
	offset     int
	prevOffset int
	end        bool

	errHandler func(error)
}

// ScannerOpt is a scanner option.
type ScannerOpt func(*Scanner)

// OptErrorHandler sets the scanner's error handler.
func OptErrorHandler(h func(error)) ScannerOpt {
	return func(s *Scanner) {
		s.errHandler = h
	}
}

// NewScanner creates a new scanner.
func NewScanner(opts ...ScannerOpt) *Scanner {
	s := &Scanner{errHandler: func(error) {}}

	for _, o := range opts {
		o(s)
	}

	return s
}

// SetReader resets the scanner ands sets its reader.
func (s *Scanner) SetReader(reader io.RuneScanner) {
	s.reader = reader
	s.ch = 0
	s.prevCh = 0
	s.pos = 0
	s.line = 1
	s.offset = 0
	s.prevOffset = 0
	s.end = false
}

// SetInput resets the scanner and sets its reader from an input string.
func (s *Scanner) SetInput(in string) {
	s.SetReader(strings.NewReader(in))
}

// Emit emits the next token.
//
// It returns TokEOF if all tokens have been read, and TokInvalid if no valid
// token was found.
func (s *Scanner) Emit() Token {
	s.read()

	for isSpace(s.ch) || s.ch == ';' {
		for isSpace(s.ch) {
			s.read()
		}

		if s.ch == ';' {
			s.stripComment()
		}
	}

	if isLine(s.ch) {
		return Token{TokLine, "", s.line, s.offset}
	}

	switch s.ch {
	case 0:
		return Token{TokEOF, "", s.line, s.offset}
	case '(':
		return Token{TokLParen, "", s.line, s.offset}
	case ')':
		return Token{TokRParen, "", s.line, s.offset}
	}

	return s.strOrSym()
}

func (s *Scanner) read() {
	s.prevOffset = s.offset
	s.prevCh = s.ch

	ch, _, err := s.reader.ReadRune()
	s.ch = ch

	if err == io.EOF {
		if !s.end {
			s.offset++
			s.end = true
		}
		return
	} else if err != nil {
		s.errHandler(errors.WithStack(err))
	}

	if ch == 0 || ch == utf8.RuneError {
		s.errHandler(ErrInvalidUTF8)
	}

	s.pos++
	s.offset++

	if isLine(ch) {
		s.line++
		s.offset = 0
	}
}

// back can only be called once after a read.
func (s *Scanner) back() {
	if err := s.reader.UnreadRune(); err != nil {
		s.errHandler(errors.WithStack(err))
	}

	s.pos--
	s.offset--

	if isLine(s.ch) {
		s.line--
		s.offset = s.prevOffset
	}

	s.ch = s.prevCh
	s.prevCh = 0
}

func (s *Scanner) stripComment() {
	s.read()

	for !isLine(s.ch) && s.ch != 0 {
		s.read()
	}
}

func (s *Scanner) error(line, pos int, ch rune) {
	s.errHandler(errors.WithStack(ScanError{line, pos, ch}))
}

func (s *Scanner) strOrSym() Token {
	switch s.ch {
	case '"':
		return s.text('"')
	case '\'':
		return s.text('\'')
	}

	return s.text(0)
}

func (s *Scanner) text(quote rune) Token {
	buf := ""

	line, offset := s.line, s.offset

	if s.ch == quote {
		s.read()
	}

	for {
		switch {
		case s.ch == 0 && quote == 0:
			return Token{TokSym, buf, line, offset}
		case s.ch == 0:
			s.error(s.line, s.offset, 0)
			return Token{TokInvalid, buf, line, offset}
		case s.ch == quote:
			return Token{TokString, buf, line, offset}
		case s.ch == '\\' && quote != 0:
			s.read()
			switch s.ch {
			case 'n':
				buf += "\n"
			case 'r':
				buf += "\r"
			case 't':
				buf += "\t"
			case quote:
				buf += string(s.ch)
			default:
				s.error(s.line, s.offset, s.ch)
				s.back()
				return Token{TokInvalid, buf, line, offset}
			}
			s.read()
		case isSpecial(s.ch) && quote == 0:
			s.back()
			return Token{TokSym, buf, line, offset}
		default:
			buf += string(s.ch)
			s.read()
		}
	}
}

func isSpace(ch rune) bool {
	return ch == ' ' || ch == '\t'
}

func isLine(ch rune) bool {
	return ch == '\n' || ch == '\r'
}

func isSpecial(ch rune) bool {
	return ch == '(' || ch == ')' || ch == ';' || isSpace(ch) || isLine(ch)
}
