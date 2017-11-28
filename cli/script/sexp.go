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

// SExpType is a type of S-Expression.
type SExpType uint8

// Available S-Expression types.
const (
	SExpList SExpType = iota
	SExpString
)

// SExp is an S-Expression.
type SExp struct {
	Type   SExpType // car type
	SExp   *SExp
	Str    string
	Cdr    *SExp
	Line   int
	Offset int
}

// String returns a string representation of the S-Expression.
func (s *SExp) String() string {
	if s.Type == SExpList {
		return fmt.Sprintf("(%s . %s)", s.SExp, s.Cdr)
	}

	return fmt.Sprintf("(%s . %s)", s.Str, s.Cdr)
}

// SExpExecutor executes S-Expression operations.
type SExpExecutor func(list *SExp) (string, error)

// Clone creates a copy of the S-Expression.
func (s *SExp) Clone() *SExp {
	if s == nil {
		return nil
	}

	return &SExp{
		Type:   s.Type,
		SExp:   s.SExp.Clone(),
		Str:    s.Str,
		Cdr:    s.Cdr.Clone(),
		Line:   s.Line,
		Offset: s.Offset,
	}
}

// EvalEach evaluates each element in a list.
func (s *SExp) EvalEach(exec SExpExecutor) ([]string, error) {
	if s == nil {
		return nil, nil
	}

	var elems []string

	for curr := s; curr != nil; curr = curr.Cdr {
		if curr.Type == SExpList {
			v, err := exec(curr.SExp)
			if err != nil {
				return nil, err
			}
			elems = append(elems, v)
		} else {
			elems = append(elems, curr.Str)
		}
	}

	return elems, nil
}
