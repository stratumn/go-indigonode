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
)

// SExpResolver resolves symbols.
type SExpResolver func(sym *SExp) (*SExp, error)

// SExpNameResolver resolves symbols with their names.
func SExpNameResolver(sym *SExp) (*SExp, error) {
	atom := sym.Clone()
	atom.Type = SExpString

	return atom, nil
}

// SExpEvaluator evaluates S-Expression operations.
type SExpEvaluator func(SExpResolver, *SExp) (string, error)

// SExpType is a type of S-Expression.
type SExpType uint8

// Available S-Expression types.
const (
	SExpList SExpType = iota
	SExpSym
	SExpString
)

// SExp is an S-Expression.
type SExp struct {
	Type   SExpType // car type
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
		elems = append(elems, curr.CarString())
	}

	return "(" + strings.Join(elems, " ") + ")"
}

// CarString returns a string representation of the card of the S-Expression.
func (s *SExp) CarString() string {
	if s == nil {
		return ""
	}

	switch s.Type {
	case SExpList:
		if s.List == nil {
			return "()"
		}
		return s.List.String()
	case SExpSym:
		return s.Str
	case SExpString:
		return fmt.Sprintf("%q", s.Str)
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
func (s *SExp) ResolveEval(resolve SExpResolver, eval SExpEvaluator) (string, error) {
	switch s.Type {
	case SExpList:
		return eval(resolve, s.List)
	case SExpSym:
		v, err := resolve(s)
		if err != nil {
			return "", err
		}
		return eval(resolve, v)
	default:
		return s.Str, nil
	}
}

// ResolveEvalEach resolves symbols and evaluates each expression in a list.
func (s *SExp) ResolveEvalEach(resolve SExpResolver, eval SExpEvaluator) ([]string, error) {
	if s == nil {
		return nil, nil
	}

	var elems []string

	for curr := s; curr != nil; curr = curr.Cdr {
		v, err := curr.ResolveEval(resolve, eval)
		if err != nil {
			return nil, err
		}

		elems = append(elems, v)
	}

	return elems, nil
}
