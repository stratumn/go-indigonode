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
	"strings"

	"github.com/pkg/errors"
)

// ResolveHandler resolves symbols.
type ResolveHandler func(sym *SExp) (*SExp, error)

// ResolveName resolves symbols with their names.
func ResolveName(sym *SExp) (*SExp, error) {
	exp := sym.Clone()
	exp.Type = TypeStr

	return exp, nil
}

// CallHandler handles function calls.
type CallHandler func(ResolveHandler, *SExp) (*SExp, error)

// Type is a type of S-Expression.
type Type uint8

// Available S-Expression types.
const (
	TypeList Type = iota
	TypeSym
	TypeStr
)

// SExp is an S-Expression.
type SExp struct {
	Type   Type // car type
	List   *SExp
	Str    string
	Cdr    *SExp
	Line   int
	Offset int
}

// String returns a string representation of the S-Expression.
func (s *SExp) String() string {
	var elems []string

	for curr := s; curr != nil; curr = curr.Cdr {
		elems = append(elems, curr.CarString(true))
	}

	return "(" + strings.Join(elems, " ") + ")"
}

// CarString returns a string representation of the card of the S-Expression.
//
// If quite is true then strings are quoted.
func (s *SExp) CarString(quote bool) string {
	if s == nil {
		return ""
	}

	switch s.Type {
	case TypeList:
		if s.List == nil {
			return "()"
		}
		return s.List.String()
	case TypeSym:
		return s.Str
	case TypeStr:
		if quote {
			return fmt.Sprintf("%q", s.Str)
		}
		return s.Str
	}

	return "<error>"
}

// Clone creates a copy of the S-Expression.
func (s *SExp) Clone() *SExp {
	if s == nil {
		return nil
	}

	return &SExp{
		Type:   s.Type,
		List:   s.List.Clone(),
		Str:    s.Str,
		Cdr:    s.Cdr.Clone(),
		Line:   s.Line,
		Offset: s.Offset,
	}
}

// ResolveEval resolves symbols and evaluates the S-Expression.
func (s *SExp) ResolveEval(resolve ResolveHandler, call CallHandler) (*SExp, error) {
	switch s.Type {
	case TypeList:
		if s.List == nil {
			return nil, nil
		}
		if s.List.Type != TypeSym {
			return nil, errors.Wrapf(
				ErrInvalidOperand,
				"%d:%d",
				s.List.Line,
				s.List.Offset,
			)
		}
		return call(resolve, s.List)
	case TypeSym:
		v, err := resolve(s)
		if err != nil {
			return nil, err
		}
		return v, nil
	default:
		return s, nil
	}
}

// ResolveEvalEach resolves symbols and evaluates each expression in a list.
func (s *SExp) ResolveEvalEach(resolve ResolveHandler, call CallHandler) (SExpSlice, error) {
	if s == nil {
		return nil, nil
	}

	var elems SExpSlice

	for curr := s; curr != nil; curr = curr.Cdr {
		v, err := curr.ResolveEval(resolve, call)
		if err != nil {
			return nil, err
		}

		elems = append(elems, v)
	}

	return elems, nil
}

// SExpSlice is a slice of S-Expressions.
type SExpSlice []*SExp

// Strings returns the string values of the car of every S-Expression in the
// slice.
//
// If quote is true strings are quoted.
func (s SExpSlice) Strings(quote bool) []string {
	str := make([]string, len(s))
	for i, exp := range s {
		str[i] = exp.CarString(quote)
	}

	return str
}

// JoinCars returns the string values of the car of every S-Expression in the
// slice separated by the given string.
//
// If quote is true strings are quoted.
func (s SExpSlice) JoinCars(sep string, quote bool) string {
	return strings.Join(s.Strings(quote), sep)
}
