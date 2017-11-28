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
)

// ParserError represents an error from the parser.
type ParserError struct {
	Tok Token
}

// Error returns an error string.
func (err ParserError) Error() string {
	return fmt.Sprintf(
		"%d:%d: unexpected token %s",
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
func (p *Parser) Parse(in string) (*SExp, error) {
	p.scanner.SetInput(in)

	p.tok = p.scanner.Emit()

	if tok := p.consume(TokEOF); tok != nil {
		return nil, nil
	}

	var sexp *SExp
	var err error

	if p.tok.Type == TokLParen {
		sexp, err = p.list()
		if err != nil {
			return nil, err
		}
		if sexp != nil {
			sexp = sexp.SExp
		}
	} else {
		sexp, err = p.inner()
		if err != nil {
			return nil, err
		}
	}

	if p.consume(TokEOF) == nil {
		return nil, ParserError{p.tok}
	}

	return sexp, nil
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
	if p.tok.Type == TokLParen {
		return p.list()
	}

	return p.string()
}

func (p *Parser) string() (*SExp, error) {
	tok := p.consume(TokString)

	if tok == nil {
		return nil, ParserError{p.tok}
	}

	return &SExp{
		Type:   SExpString,
		Str:    tok.Value,
		Line:   tok.Line,
		Offset: tok.Offset,
	}, nil
}

func (p *Parser) list() (*SExp, error) {
	tok := p.consume(TokLParen)

	if tok == nil {
		return nil, ParserError{p.tok}
	}

	if p.consume(TokRParen) != nil {
		return &SExp{
			Type:   SExpList,
			SExp:   nil,
			Line:   tok.Line,
			Offset: tok.Offset,
		}, nil
	}

	sexp, err := p.inner()
	if err != nil {
		return nil, err
	}

	if p.consume(TokRParen) == nil {
		return nil, ParserError{p.tok}
	}

	return &SExp{
		Type:   SExpList,
		SExp:   sexp,
		Line:   tok.Line,
		Offset: tok.Offset,
	}, nil
}

func (p *Parser) inner() (*SExp, error) {
	tok := p.tok
	atom := p.consume(TokString)

	if atom == nil {
		return nil, ParserError{p.tok}
	}

	head := &SExp{
		Type:   SExpString,
		Str:    atom.Value,
		Line:   tok.Line,
		Offset: tok.Offset,
	}

	curr := head

	for {
		cdr, err := p.sexp()
		if err != nil {
			break
		}

		curr.Cdr = cdr
		curr = cdr
	}

	return head, nil
}
