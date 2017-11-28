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
// It returns a list of S-Expression which can be evaluated.
func (p *Parser) Parse(in string) (*SExp, error) {
	p.scanner.SetInput(in)
	p.scan()

	var head, tail *SExp

	for {
		instr, err := p.instr()

		if err != nil {
			return nil, err
		}

		if instr == nil {
			return head, nil
		}

		if tail == nil {
			head = instr
			tail = instr
			continue
		}

		tail.Cdr = instr
		tail = instr
	}
}

func (p *Parser) instr() (*SExp, error) {
	p.skipLines()

	if tok := p.consume(TokEOF); tok != nil {
		return nil, nil
	}

	tok := p.tok

	if tok.Type == TokLParen {
		list, err := p.list()
		if err != nil {
			return nil, err
		}

		if p.tok.Type != TokLine && p.tok.Type != TokEOF {
			return nil, errors.WithStack(ParserError{p.tok})
		}

		return list, nil
	}

	call, err := p.call(false)
	if err != nil {
		return nil, err
	}

	return &SExp{
		Type:   SExpList,
		List:   call,
		Line:   tok.Line,
		Offset: tok.Offset,
	}, nil
}

func (p *Parser) call(inList bool) (*SExp, error) {
	head, err := p.string(inList)
	if err != nil {
		return nil, err
	}

	tail := head

	for {
		if inList {
			p.skipLines()
		}

		if inList && p.tok.Type == TokRParen {
			return head, nil
		}

		if !inList && (p.tok.Type == TokLine || p.tok.Type == TokEOF) {
			return head, nil
		}

		operand, err := p.sexp(inList)
		if err != nil {
			return nil, err
		}

		tail.Cdr = operand
		tail = operand
	}
}

func (p *Parser) sexp(inList bool) (*SExp, error) {
	if inList {
		p.skipLines()
	}

	if p.tok.Type == TokLParen {
		return p.list()
	}

	return p.string(inList)
}

func (p *Parser) list() (*SExp, error) {
	p.skipLines()

	tok := p.consume(TokLParen)

	if tok == nil {
		return nil, errors.WithStack(ParserError{p.tok})
	}

	p.skipLines()

	call, err := p.call(true)
	if err != nil {
		return nil, err
	}

	p.skipLines()

	if p.consume(TokRParen) == nil {
		return nil, errors.WithStack(ParserError{p.tok})
	}

	return &SExp{
		Type:   SExpList,
		List:   call,
		Line:   tok.Line,
		Offset: tok.Offset,
	}, nil
}

func (p *Parser) string(inList bool) (*SExp, error) {
	if inList {
		p.skipLines()
	}

	tok := p.consume(TokString)

	if tok == nil {
		return nil, errors.WithStack(ParserError{p.tok})
	}

	return &SExp{
		Type:   SExpString,
		Str:    tok.Value,
		Line:   tok.Line,
		Offset: tok.Offset,
	}, nil
}
