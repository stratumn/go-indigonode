// Copyright © 2017-2018 Stratumn SAS
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as
// published by the Free Software Foundation, either version 3 of the
// License, or (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <https://www.gnu.org/licenses/>.

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
	for p.consume(TokLine) != nil {
	}
}

// Parse parses instructions in the given input.
//
// It returns a list of S-Expressions which can be evaluated. It returns nil
// if there are not instructions.
func (p *Parser) Parse(in string) (SExp, error) {
	p.scanner.SetInput(in)
	p.scan()

	return p.script()
}

// List parses a single list.
func (p *Parser) List(in string) (SExp, error) {
	p.scanner.SetInput(in)
	p.scan()

	p.skipLines()

	list, err := p.list()
	if err != nil {
		return nil, err
	}

	p.skipLines()

	if p.tok.Type != TokEOF {
		return nil, errors.WithStack(ParseError{Tok: p.tok})
	}

	return list, nil
}

func (p *Parser) script() (SExp, error) {
	head, err := p.inBody()
	if err != nil {
		return nil, err
	}

	p.skipLines()

	if tok := p.consume(TokEOF); tok == nil {
		return nil, errors.WithStack(ParseError{Tok: p.tok})
	}

	return head, nil
}

func (p *Parser) inBody() (SExp, error) {
	meta := Meta{Line: p.tok.Line, Offset: p.tok.Offset}

	car, err := p.instr()
	if err != nil {
		return nil, err
	}

	if car == nil {
		car, err = p.inBodyTail()
		if err != nil {
			return nil, err
		}

		if car == nil {
			car = Nil()
		}

		return car, nil
	}

	cdr, err := p.inBodyTail()
	if err != nil {
		return nil, err
	}

	return Cons(car, cdr, meta), nil
}

func (p *Parser) inBodyTail() (SExp, error) {
	if tok := p.consume(TokLine); tok == nil {
		return nil, nil
	}

	p.skipLines()

	meta := Meta{Line: p.tok.Line, Offset: p.tok.Offset}

	car, err := p.instr()
	if car == nil || err != nil {
		return nil, err
	}

	cdr, err := p.inBodyTail()
	if err != nil {
		return nil, err
	}

	return Cons(car, cdr, meta), nil
}

func (p *Parser) instr() (SExp, error) {
	switch p.tok.Type {
	case TokLParen:
		return p.call()
	case TokSymbol:
		return p.inCall()
	}

	return nil, nil
}

func (p *Parser) call() (SExp, error) {
	if p.consume(TokLParen) == nil {
		// This actually never happens because the caller checks the
		// token.
		return nil, errors.WithStack(ParseError{Tok: p.tok})
	}

	call, err := p.inCallInParen()
	if err != nil {
		return nil, err
	}

	p.skipLines()

	if p.consume(TokRParen) == nil {
		// This actually never happens because sexpListInParen checks
		// the token before returning.
		return nil, errors.WithStack(ParseError{Tok: p.tok})
	}

	return call, nil
}

func (p *Parser) inCallInParen() (SExp, error) {
	p.skipLines()

	meta := Meta{Line: p.tok.Line, Offset: p.tok.Offset}

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

func (p *Parser) inCall() (SExp, error) {
	meta := Meta{Line: p.tok.Line, Offset: p.tok.Offset}

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

func (p *Parser) sexpListInParen() (SExp, error) {
	meta := Meta{Line: p.tok.Line, Offset: p.tok.Offset}

	car, err := p.sexpInParen()
	if car == nil || err != nil {
		return nil, err
	}

	cdr, err := p.sexpListInParen()
	if err != nil {
		return nil, err
	}

	return Cons(car, cdr, meta), nil
}

func (p *Parser) sexpList() (SExp, error) {
	meta := Meta{Line: p.tok.Line, Offset: p.tok.Offset}

	car, err := p.sexp()
	if car == nil || err != nil {
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
	switch p.tok.Type {
	case TokQuote:
		return p.quotedSExp()
	case TokQuasiquote:
		return p.quasiquotedSExp()
	case TokUnquote:
		return p.unquotedSExp()
	case TokLBrace:
		return p.body()
	case TokLParen:
		return p.list()
	case TokSymbol, TokString, TokInt, TokTrue, TokFalse:
		return p.atom()
	default:
		return nil, nil
	}
}

func (p *Parser) quotedSExp() (SExp, error) {
	tok := p.consume(TokQuote)
	if tok == nil {
		// This actually never happens because the caller checks the
		// token.
		return nil, errors.WithStack(ParseError{Tok: p.tok})
	}

	p.skipLines()

	exp, err := p.sexp()
	if err != nil {
		return nil, err
	}

	meta := Meta{Line: tok.Line, Offset: tok.Offset}

	return Cons(Symbol(QuoteSymbol, meta), Cons(exp, nil, meta), meta), nil
}

func (p *Parser) quasiquotedSExp() (SExp, error) {
	tok := p.consume(TokQuasiquote)
	if tok == nil {
		// This actually never happens because the caller checks the
		// token.
		return nil, errors.WithStack(ParseError{Tok: p.tok})
	}

	p.skipLines()

	exp, err := p.sexp()
	if err != nil {
		return nil, err
	}

	meta := Meta{Line: tok.Line, Offset: tok.Offset}

	return Cons(Symbol(QuasiquoteSymbol, meta), Cons(exp, nil, meta), meta), nil
}

func (p *Parser) unquotedSExp() (SExp, error) {
	tok := p.consume(TokUnquote)
	if tok == nil {
		// This actually never happens because the caller checks the
		// token.
		return nil, errors.WithStack(ParseError{Tok: p.tok})
	}

	p.skipLines()

	exp, err := p.sexp()
	if err != nil {
		return nil, err
	}

	meta := Meta{Line: tok.Line, Offset: tok.Offset}

	return Cons(Symbol(UnquoteSymbol, meta), Cons(exp, nil, meta), meta), nil
}

func (p *Parser) body() (SExp, error) {
	if p.consume(TokLBrace) == nil {
		return nil, errors.WithStack(ParseError{Tok: p.tok})
	}

	head, err := p.inBody()
	if err != nil {
		return nil, err
	}

	p.skipLines()

	if p.consume(TokRBrace) == nil {
		return nil, errors.WithStack(ParseError{Tok: p.tok})
	}

	return head, nil
}

func (p *Parser) list() (SExp, error) {
	if p.consume(TokLParen) == nil {
		return nil, errors.WithStack(ParseError{Tok: p.tok})
	}

	list, err := p.sexpListInParen()
	if err != nil {
		return nil, err
	}

	if list == nil {
		list = Nil()
	}

	p.skipLines()

	if p.consume(TokRParen) == nil {
		return nil, errors.WithStack(ParseError{Tok: p.tok})
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
	case TokTrue, TokFalse:
		return p.bool()
	}

	return nil, errors.WithStack(ParseError{Tok: p.tok})
}

func (p *Parser) symbol() (SExp, error) {
	tok := p.consume(TokSymbol)
	if tok == nil {
		// This actually never happens because the caller checks the
		// token.
		return nil, errors.WithStack(ParseError{Tok: p.tok})
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
		return nil, errors.WithStack(ParseError{Tok: p.tok})
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
		return nil, errors.WithStack(ParseError{Tok: p.tok})
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

func (p *Parser) bool() (SExp, error) {
	if tok := p.consume(TokTrue); tok != nil {
		return Bool(true, Meta{
			Line:   tok.Line,
			Offset: tok.Offset,
		}), nil
	}

	if tok := p.consume(TokFalse); tok != nil {
		return Bool(false, Meta{
			Line:   tok.Line,
			Offset: tok.Offset,
		}), nil
	}

	// This actually never happens because the caller checks the
	// token.
	return nil, errors.WithStack(ParseError{Tok: p.tok})
}
