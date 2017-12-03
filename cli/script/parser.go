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

	var head, tail SCell

	for {
		instr, err := p.instr()
		if err != nil {
			return nil, err
		}

		if instr == nil {
			return head, nil
		}

		if head == nil {
			head = Cons(instr, nil, Meta{
				Line:   1,
				Offset: 1,
			})
			tail = head
			continue
		}

		cdr := Cons(instr, nil, instr.Meta())
		tail.SetCdr(cdr)
		tail = cdr
	}
}

// List parses a list.
func (p *Parser) List(in string) (SCell, error) {
	p.scanner.SetInput(in)
	p.scan()

	head, err := p.list(false)
	if err != nil {
		return nil, err
	}

	p.skipLines()

	if p.tok.Type != TokEOF {
		return nil, errors.WithStack(ParseError{p.tok})
	}

	return head, nil
}

func (p *Parser) instr() (SCell, error) {
	p.skipLines()

	if tok := p.consume(TokEOF); tok != nil {
		return nil, nil
	}

	tok := p.tok

	if tok.Type == TokLParen {
		head, err := p.list(true)
		if err != nil {
			return nil, err
		}

		if p.tok.Type != TokLine && p.tok.Type != TokEOF {
			return nil, errors.WithStack(ParseError{p.tok})
		}

		return head, nil
	}

	return p.cells(false, true)
}

func (p *Parser) sexp(inParen bool) (SExp, error) {
	if inParen {
		p.skipLines()
	}

	if p.tok.Type == TokLParen {
		return p.list(false)
	}

	return p.atom(inParen)
}

func (p *Parser) list(isCall bool) (SCell, error) {
	p.skipLines()

	tok := p.consume(TokLParen)

	if tok == nil {
		return nil, errors.WithStack(ParseError{p.tok})
	}

	p.skipLines()

	if !isCall && p.consume(TokRParen) != nil {
		return nil, nil
	}

	head, err := p.cells(true, isCall)
	if err != nil {
		return nil, err
	}

	p.skipLines()

	if p.consume(TokRParen) == nil {
		return nil, errors.WithStack(ParseError{p.tok})
	}

	return head, nil
}

func (p *Parser) cells(inParen, isCall bool) (SCell, error) {
	var head SCell
	var first SExp
	var err error

	if isCall {
		first, err = p.sym(inParen)

	} else {
		first, err = p.sexp(inParen)
	}

	if err != nil {
		return nil, err
	}

	head = Cons(first, nil, first.Meta())
	tail := head

	for {
		if inParen {
			p.skipLines()
		}

		if inParen && p.tok.Type == TokRParen {
			return head, nil
		}

		if !inParen && (p.tok.Type == TokLine || p.tok.Type == TokEOF) {
			return head, nil
		}

		exp, err := p.sexp(inParen)
		if err != nil {
			return nil, err
		}

		var cdr SCell

		if exp == nil {
			cdr = Cons(nil, nil, tail.Meta())
		} else {
			cdr = Cons(exp, nil, exp.Meta())
		}

		tail.SetCdr(cdr)
		tail = cdr
	}
}

func (p *Parser) atom(inParen bool) (SExp, error) {
	switch p.tok.Type {
	case TokSym:
		return p.sym(inParen)
	case TokString:
		return p.string(inParen)
	}

	return nil, errors.WithStack(ParseError{p.tok})
}

func (p *Parser) sym(inParen bool) (SExp, error) {
	if inParen {
		p.skipLines()
	}

	tok := p.consume(TokSym)

	if tok == nil {
		return nil, errors.WithStack(ParseError{p.tok})
	}

	return Symbol(tok.Value, Meta{
		Line:   tok.Line,
		Offset: tok.Offset,
	}), nil
}

func (p *Parser) string(inParen bool) (SExp, error) {
	if inParen {
		p.skipLines()
	}

	tok := p.consume(TokString)

	if tok == nil {
		return nil, errors.WithStack(ParseError{p.tok})
	}

	return String(tok.Value, Meta{
		Line:   tok.Line,
		Offset: tok.Offset,
	}), nil
}
