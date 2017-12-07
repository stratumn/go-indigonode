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
	"strconv"

	"github.com/pkg/errors"
)

// ParseError represents an error from the parser.
type ParseError struct {
	Tok Token
}

// Error returns an error string.
func (err ParseError) Error() string {
	return fmt.Sprintf(
		"%d:%d: unexpected token %s",
		err.Tok.Line,
		err.Tok.Offset,
		err.Tok.Type,
	)
}

// Parser produces an S-Expression list from scanner tokens.
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

func (p *Parser) skipLines() {
	for {
		if tok := p.consume(TokLine); tok == nil {
			return
		}
	}
}

// Parse parses instructions in the given input.
//
// It returns a list of S-Expressions which can be evaluated. It returns nil
// if there are not instructions.
func (p *Parser) Parse(in string) (SCell, error) {
	p.scanner.SetInput(in)
	p.scan()

	return p.script()
}

// List parses a single list.
func (p *Parser) List(in string) (SCell, error) {
	p.scanner.SetInput(in)
	p.scan()

	p.skipLines()

	list, err := p.list()
	if err != nil {
		return nil, err
	}

	p.skipLines()

	if p.tok.Type != TokEOF {
		return nil, errors.WithStack(ParseError{p.tok})
	}

	return list, nil
}

func (p *Parser) script() (SCell, error) {
	meta := Meta{
		Line:   p.tok.Line,
		Offset: p.tok.Offset,
	}

	car, err := p.instr()
	if err != nil {
		return nil, err
	}

	if car == nil {
		return nil, nil
	}

	cdr, err := p.script()
	if err != nil {
		return nil, err
	}

	return Cons(car, cdr, meta), nil
}

func (p *Parser) instr() (SCell, error) {
	p.skipLines()

	if tok := p.consume(TokEOF); tok != nil {
		return nil, nil
	}

	tok := p.tok

	if tok.Type == TokLParen {
		call, err := p.call()
		if err != nil {
			return nil, err
		}

		if p.tok.Type != TokLine && p.tok.Type != TokEOF {
			return nil, errors.WithStack(ParseError{p.tok})
		}

		return call, nil
	}

	return p.inCall()
}

func (p *Parser) call() (SCell, error) {
	p.skipLines()

	tok := p.consume(TokLParen)
	if tok == nil {
		// This actually never happens because the caller checks the
		// token.
		return nil, errors.WithStack(ParseError{p.tok})
	}

	call, err := p.inCallInParen()
	if err != nil {
		return nil, err
	}

	p.skipLines()

	if p.consume(TokRParen) == nil {
		// This actually never happens because sexpListInParen checks
		// the token before returning.
		return nil, errors.WithStack(ParseError{p.tok})
	}

	return call, nil
}

func (p *Parser) inCallInParen() (SCell, error) {
	p.skipLines()

	meta := Meta{
		Line:   p.tok.Line,
		Offset: p.tok.Offset,
	}

	car, err := p.symbol()
	if err != nil {
		return nil, err
	}

	cdr, err := p.sexpListInParen()
	if err != nil {
		return nil, err
	}

	return Cons(car, cdr, meta), nil
}

func (p *Parser) inCall() (SCell, error) {
	meta := Meta{
		Line:   p.tok.Line,
		Offset: p.tok.Offset,
	}

	car, err := p.symbol()
	if err != nil {
		return nil, err
	}

	cdr, err := p.sexpList()
	if err != nil {
		return nil, err
	}

	return Cons(car, cdr, meta), nil
}

func (p *Parser) sexpListInParen() (SCell, error) {
	p.skipLines()

	meta := Meta{
		Line:   p.tok.Line,
		Offset: p.tok.Offset,
	}

	if p.tok.Type == TokRParen {
		return nil, nil
	}

	car, err := p.sexpInParen()
	if err != nil {
		return nil, err
	}

	cdr, err := p.sexpListInParen()
	if err != nil {
		return nil, err
	}

	return Cons(car, cdr, meta), nil
}

func (p *Parser) sexpList() (SCell, error) {
	meta := Meta{
		Line:   p.tok.Line,
		Offset: p.tok.Offset,
	}

	if p.tok.Type == TokLine || p.tok.Type == TokEOF {
		return nil, nil
	}

	car, err := p.sexp()
	if err != nil {
		return nil, err
	}

	cdr, err := p.sexpList()
	if err != nil {
		return nil, err
	}

	return Cons(car, cdr, meta), nil
}

func (p *Parser) sexpInParen() (SExp, error) {
	p.skipLines()

	return p.sexp()
}

func (p *Parser) sexp() (SExp, error) {
	if p.tok.Type == TokLParen {
		return p.list()
	}

	return p.atom()
}

func (p *Parser) list() (SCell, error) {
	tok := p.consume(TokLParen)
	if tok == nil {
		return nil, errors.WithStack(ParseError{p.tok})
	}

	if p.consume(TokRParen) != nil {
		return nil, nil
	}

	list, err := p.sexpListInParen()
	if err != nil {
		return nil, err
	}

	p.skipLines()

	if p.consume(TokRParen) == nil {
		// This actually never happens because sexpListInParen checks
		// the token before returning.
		return nil, errors.WithStack(ParseError{p.tok})
	}

	return list, nil
}

func (p *Parser) atom() (SExp, error) {
	switch p.tok.Type {
	case TokSymbol:
		return p.symbol()
	case TokString:
		return p.string()
	case TokInt:
		return p.int()
	}

	return nil, errors.WithStack(ParseError{p.tok})
}

func (p *Parser) symbol() (SExp, error) {
	tok := p.consume(TokSymbol)

	if tok == nil {
		return nil, errors.WithStack(ParseError{p.tok})
	}

	return Symbol(tok.Value, Meta{
		Line:   tok.Line,
		Offset: tok.Offset,
	}), nil
}

func (p *Parser) string() (SExp, error) {
	tok := p.consume(TokString)

	if tok == nil {
		// This actually never happens because the caller checks the
		// token.
		return nil, errors.WithStack(ParseError{p.tok})
	}

	return String(tok.Value, Meta{
		Line:   tok.Line,
		Offset: tok.Offset,
	}), nil
}

func (p *Parser) int() (SExp, error) {
	tok := p.consume(TokInt)

	if tok == nil {
		// This actually never happens because the caller checks the
		// token.
		return nil, errors.WithStack(ParseError{p.tok})
	}

	i, err := strconv.ParseInt(tok.Value, 0, 64)
	if err != nil {
		return nil, errors.Wrapf(err, "%d:%d", tok.Line, tok.Offset)
	}

	return Int64(i, Meta{
		Line:   tok.Line,
		Offset: tok.Offset,
	}), nil
}
