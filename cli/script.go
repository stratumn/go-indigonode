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
	"fmt"
)

// TokenType is a script token.
type TokenType uint8

// Available tokens.
const (
	TokInvalid TokenType = iota
	TokEOF
	TokLParen
	TokRParen
	TokString
)

var tokToStr = map[TokenType]string{
	TokInvalid: "<invalid>",
	TokEOF:     "<EOF>",
	TokLParen:  "(",
	TokRParen:  ")",
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
	return fmt.Sprintf("%d;%d: unexpected character %q", err.Line, err.Offset, err.Ch)
}

// Scanner produces tokens from a string.
type Scanner struct {
	runes  []rune
	len    int
	line   int
	pos    int
	offset int
	end    bool

	errHandler func(ScannerError)
}

// ScannerOpt is a scanner option.
type ScannerOpt func(*Scanner)

// ScannerOptErrorHandler sets the scanner's error handler.
func ScannerOptErrorHandler(h func(ScannerError)) ScannerOpt {
	return func(s *Scanner) {
		s.errHandler = h
	}
}

// NewScanner creates a new scanner.
func NewScanner(opts ...ScannerOpt) *Scanner {
	s := &Scanner{
		runes:      nil,
		len:        0,
		errHandler: func(ScannerError) {},
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
	s.end = false
}

// Emit emits the next token.
//
// It returns TokEOF if all tokens have been read, and TokInvalid if no valid
// tokens were found.
func (s *Scanner) Emit() Token {
	c := s.read()

	for isSpace(c) {
		c = s.read()
	}

	switch c {
	case 0:
		return Token{TokEOF, "", s.line, s.offset}
	case '(':
		return Token{TokLParen, "", s.line, s.offset}
	case ')':
		return Token{TokRParen, "", s.line, s.offset}
	}

	return s.string(c)
}

func (s *Scanner) read() rune {
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

func (s *Scanner) error(line, pos int, c rune) {
	s.errHandler(ScannerError{line, pos, c})
}

func (s *Scanner) string(c rune) Token {
	switch c {
	case '"':
		return s.innerString(c, '"')
	case '\'':
		return s.innerString(c, '\'')
	}

	return s.innerString(c, 0)
}

func (s *Scanner) innerString(c rune, quote rune) Token {
	buf := ""

	line, offset := s.line, s.offset

	if c == quote {
		c = s.read()
	}

	for {
		switch {
		case c == 0 && quote == 0:
			return Token{TokString, buf, line, offset}
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
				s.pos--
				s.offset--
				return Token{TokInvalid, buf, line, offset}
			}
			c = s.read()
		case isSpecial(c) && quote == 0:
			if !isLine(c) {
				s.pos--
				s.offset--
			}
			return Token{TokString, buf, line, offset}
		default:
			buf += string(c)
			c = s.read()
		}
	}
}

func isSpace(c rune) bool {
	return c == ' ' || c == '\t' || isLine(c)
}

func isLine(c rune) bool {
	return c == '\n' || c == '\r'
}

func isSpecial(c rune) bool {
	return c == '(' || c == ')' || isSpace(c) || isLine(c)
}

// ParserError represents an error from the parser.
type ParserError struct {
	Tok Token
}

// Error returns an error string.
func (err ParserError) Error() string {
	return fmt.Sprintf(
		"%d;%d: unexpected token %s",
		err.Tok.Line,
		err.Tok.Offset,
		err.Tok.Type,
	)
}

// Parser produces an abstract syntax tree from a scanner.
type Parser struct {
	scanner *Scanner
	tok     Token
}

// NewParser creates a new parser.
func NewParser(scanner *Scanner) *Parser {
	return &Parser{
		scanner: scanner,
	}
}

// Parse creates an S-Expression from the given input.
func (p *Parser) Parse(in string) (sexp *SExp, err error) {
	p.scanner.SetInput(in)

	p.tok = p.scanner.Emit()

	if tok := p.consume(TokEOF); tok != nil {
		return nil, nil
	}

	if p.tok.Type == TokLParen {
		sexp, err = p.sexp()
	} else {
		sexp, err = p.inner()
	}

	if p.consume(TokEOF) == nil {
		return nil, ParserError{p.tok}
	}

	return
}

func (p *Parser) scan() {
	p.tok = p.scanner.Emit()
}

func (p *Parser) consume(typ TokenType) *Token {
	tok := p.tok

	if tok.Type != typ {
		return nil
	}

	p.scan()

	return &tok
}

func (p *Parser) sexp() (*SExp, error) {
	if tok := p.consume(TokString); tok != nil {
		return &SExp{false, nil, nil, tok.Value, tok.Line, tok.Offset}, nil
	}

	if tok := p.consume(TokLParen); tok != nil {
		if p.consume(TokRParen) != nil {
			return nil, nil
		}

		sexp, err := p.inner()
		if err != nil {
			return nil, err
		}

		if p.consume(TokRParen) == nil {
			return nil, ParserError{p.tok}
		}

		return sexp, nil
	}

	return nil, ParserError{p.tok}
}

// one (two three) four
func (p *Parser) inner() (*SExp, error) {
	tok := p.tok
	atom := p.consume(TokString)

	if atom == nil {
		return nil, ParserError{p.tok}
	}

	parent := &SExp{
		List:   true,
		Line:   tok.Line,
		Offset: tok.Offset,
		Atom:   atom.Value,
	}

	curr := parent

	for {
		sexp, err := p.sexp()
		if err != nil {
			break
		}

		if sexp.List {
			sexp = &SExp{
				List:   true,
				Car:    sexp,
				Line:   curr.Line,
				Offset: curr.Offset,
			}
		}

		curr.Cdr = sexp
		curr = sexp
	}

	return parent, nil
}

// SExp is an S-Expression.
type SExp struct {
	List   bool
	Car    *SExp
	Cdr    *SExp
	Atom   string
	Line   int
	Offset int
}

// String returns a string representation of the S-Expression.
func (s *SExp) String() string {
	if s.Car != nil {
		return fmt.Sprintf("(%s . %s)", s.Car, s.Cdr)
	}

	return fmt.Sprintf("(%s . %s)", s.Atom, s.Cdr)
}

// SExpEvaluator evalutates an S-Expression.
type SExpEvaluator func(atom string, line int, offset int, args ...*SExp) (string, error)

// Eval evaluates the expression.
func (s *SExp) Eval(eval SExpEvaluator) (string, error) {
	if s == nil {
		return "", nil
	}

	var args []*SExp

	curr := s.Cdr

	for curr != nil {
		args = append(args, curr)
		curr = curr.Cdr
	}

	return eval(s.Atom, s.Line, s.Offset, args...)
}

// EvalMulti evaluates each given S-Expression.
func EvalMulti(eval SExpEvaluator, sexps ...*SExp) ([]string, error) {
	var vals []string

	for _, s := range sexps {
		if s.Car == nil {
			vals = append(vals, s.Atom)
			continue
		}

		v, err := s.Car.Eval(eval)
		if err != nil {
			return nil, err
		}

		vals = append(vals, v)
	}

	return vals, nil
}
