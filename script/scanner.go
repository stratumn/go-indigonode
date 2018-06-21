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
	TokLBrace
	TokRBrace
	TokQuote
	TokQuasiquote
	TokUnquote
	TokSymbol
	TokString
	TokInt
	TokTrue
	TokFalse
)

var tokToStr = map[TokenType]string{
	TokInvalid:    "<invalid>",
	TokLine:       "<line>",
	TokEOF:        "<EOF>",
	TokLParen:     "(",
	TokRParen:     ")",
	TokLBrace:     "{",
	TokRBrace:     "}",
	TokQuote:      "'",
	TokQuasiquote: "`",
	TokUnquote:    ",",
	TokSymbol:     "<sym>",
	TokString:     "<string>",
	TokInt:        "<int>",
	TokTrue:       "true",
	TokFalse:      "false",
}

// String returns a string representation of a token type.
func (tok TokenType) String() string {
	return tokToStr[tok]
}

// Keyword token map.
var keywords = map[string]TokenType{
	"true":  TokTrue,
	"false": TokFalse,
}

// Characters reserved for future syntax.
var reserved = map[rune]struct{}{
	'@':  struct{}{},
	'[':  struct{}{},
	']':  struct{}{},
	'\\': struct{}{},
	'|':  struct{}{},
	':':  struct{}{},
}

// Token represents a token.
type Token struct {
	Type   TokenType
	Value  string
	Line   int
	Offset int
}

// String returns a string representation of the token.
func (t Token) String() string {
	return fmt.Sprintf("%d:%d: %s %s", t.Line, t.Offset, t.Type, t.Value)
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

// ScannerOpt is a scanner option.
type ScannerOpt func(*Scanner)

// ScannerOptErrorHandler sets the scanner's error handler.
func ScannerOptErrorHandler(h func(error)) ScannerOpt {
	return func(s *Scanner) {
		s.errHandler = h
	}
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
	endOffset  int

	errHandler func(error)
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
	s.endOffset = 0
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
		return Token{Type: TokLine, Value: "", Line: s.line, Offset: s.offset}
	}

	if isReserved(s.ch) {
		s.error(s.line, s.offset, s.ch)
		return Token{Type: TokInvalid, Value: "", Line: s.line, Offset: s.offset}
	}

	switch s.ch {
	case 0:
		return Token{Type: TokEOF, Value: "", Line: s.line, Offset: s.offset}
	case '(':
		return Token{Type: TokLParen, Value: "", Line: s.line, Offset: s.offset}
	case ')':
		return Token{Type: TokRParen, Value: "", Line: s.line, Offset: s.offset}
	case '{':
		return Token{Type: TokLBrace, Value: "", Line: s.line, Offset: s.offset}
	case '}':
		return Token{Type: TokRBrace, Value: "", Line: s.line, Offset: s.offset}
	case '\'':
		return Token{Type: TokQuote, Value: "", Line: s.line, Offset: s.offset}
	case '`':
		return Token{Type: TokQuasiquote, Value: "", Line: s.line, Offset: s.offset}
	case ',':
		return Token{Type: TokUnquote, Value: "", Line: s.line, Offset: s.offset}
	}

	return s.atom()
}

func (s *Scanner) read() {
	s.prevOffset = s.offset
	s.prevCh = s.ch

	ch, _, err := s.reader.ReadRune()
	s.ch = ch

	if err == io.EOF {
		if s.end {
			s.offset = s.endOffset
		} else {
			s.offset++
			s.endOffset = s.offset
			s.end = true
		}
		return
	} else if err != nil {
		// Should never happen.
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
	if !s.end {
		if err := s.reader.UnreadRune(); err != nil {
			// Should never happen.
			s.errHandler(errors.WithStack(err))
		}
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
	s.errHandler(errors.WithStack(ScanError{Line: line, Offset: pos, Rune: ch}))
}

func (s *Scanner) atom() Token {
	switch {
	case s.ch == '"':
		return s.string()
	case isDigit(s.ch) || s.ch == '-':
		return s.intOrSymbol()
	}

	return s.symbolOrKeyword()
}

func (s *Scanner) string() Token {
	line, offset := s.line, s.offset
	buf := ""
	s.read()

	for {
		switch {
		case s.ch == 0:
			s.error(s.line, s.offset, 0)
			return Token{Type: TokInvalid, Value: buf, Line: line, Offset: offset}
		case s.ch == '"':
			return Token{Type: TokString, Value: buf, Line: line, Offset: offset}
		case s.ch == '\\':
			s.read()
			switch s.ch {
			case 'n':
				buf += "\n"
			case 'r':
				buf += "\r"
			case 't':
				buf += "\t"
			case '"':
				buf += string(s.ch)
			default:
				s.error(s.line, s.offset, s.ch)
				return Token{Type: TokInvalid, Value: buf, Line: line, Offset: offset}
			}
			s.read()
		default:
			buf += string(s.ch)
			s.read()
		}
	}
}

func (s *Scanner) symbolOrKeyword() Token {
	return s.appendSymbolOrKeyword("", s.line, s.offset)
}

func (s *Scanner) appendSymbolOrKeyword(buf string, line, offset int) Token {
	for {
		switch {
		case s.ch == 0:
			return makeSymbolOrKeyword(buf, line, offset)
		case isSpecial(s.ch):
			s.back()
			return makeSymbolOrKeyword(buf, line, offset)
		case isReserved(s.ch):
			s.error(s.line, s.offset, s.ch)
			return Token{Type: TokInvalid, Value: buf, Line: s.line, Offset: offset}
		case s.ch == '"':
			s.error(s.line, s.offset, s.ch)
			return Token{Type: TokInvalid, Value: buf, Line: line, Offset: offset}
		default:
			buf += string(s.ch)
			s.read()
		}
	}
}

// Any invalid int is a valid symbol. This is a design decision so that flags
// can be scanned. Perhaps this could be disabled using an option.
func (s *Scanner) intOrSymbol() Token {
	line, offset := s.line, s.offset
	buf := ""
	octal := false

	if s.ch == '0' {
		s.read()

		// Handle hexadecimal number.
		if s.ch == 'x' || s.ch == 'X' {
			buf += "0" + string(s.ch)
			s.read()

			for {
				switch {
				case isReserved(s.ch):
					s.error(s.line, s.offset, s.ch)
					return Token{Type: TokInvalid, Value: buf, Line: line, Offset: offset}
				case s.ch == 0:
					return Token{Type: TokInt, Value: buf, Line: line, Offset: offset}
				case isSpecial(s.ch):
					s.back()
					return Token{Type: TokInt, Value: buf, Line: line, Offset: offset}
				case isHex(s.ch):
					buf += string(s.ch)
					s.read()
				default:
					return s.appendSymbolOrKeyword(buf, line, offset)
				}
			}
		}

		s.back()
		octal = true
	}

	if s.ch == '-' {
		buf += string(s.ch)
		s.read()

		if !isDigit(s.ch) {
			return s.appendSymbolOrKeyword("-", line, offset)
		}

		octal = s.ch == '0'
	}

	for {
		switch {
		case isReserved(s.ch):
			s.error(s.line, s.offset, s.ch)
			return Token{Type: TokInvalid, Value: buf, Line: line, Offset: offset}
		case s.ch == 0:
			return Token{Type: TokInt, Value: buf, Line: line, Offset: offset}
		case isSpecial(s.ch):
			s.back()
			return Token{Type: TokInt, Value: buf, Line: line, Offset: offset}
		case octal && isOctal(s.ch):
			buf += string(s.ch)
			s.read()
		case !octal && isDigit(s.ch):
			buf += string(s.ch)
			s.read()
		default:
			return s.appendSymbolOrKeyword(buf, line, offset)
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
	return ch == '(' || ch == ')' || ch == '{' || ch == '}' ||
		ch == '\'' || ch == ';' || isSpace(ch) || isLine(ch)
}

func isDigit(ch rune) bool {
	return ch >= '0' && ch <= '9'
}

func isHex(ch rune) bool {
	return isDigit(ch) || ch >= 'a' && ch <= 'f' || ch >= 'A' && ch <= 'F'
}

func isOctal(ch rune) bool {
	return ch >= '0' && ch <= '7'
}

func isReserved(ch rune) bool {
	_, ok := reserved[ch]

	return ok
}

func makeSymbolOrKeyword(s string, line, offset int) Token {
	for k, v := range keywords {
		if k == s {
			return Token{Type: v, Value: "", Line: line, Offset: offset}
		}
	}

	return Token{Type: TokSymbol, Value: s, Line: line, Offset: offset}
}
