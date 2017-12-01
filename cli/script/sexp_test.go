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
	"testing"

	"github.com/pkg/errors"
)

type evalTest struct {
	input  string
	output string
	err    string
}

var evalTests = []evalTest{{
	"",
	"",
	"",
}, {
	"echo",
	"",
	"",
}, {
	"echo hello",
	"hello",
	"",
}, {
	"echo (echo)",
	"",
	"",
}, {
	"echo echo hello",
	"echo hello",
	"",
}, {
	"(echo hello)",
	"hello",
	"",
}, {
	"echo 'hello  world'",
	"hello  world",
	"",
}, {
	"echo hello\n(echo world)",
	"hello\nworld",
	"",
}, {
	"(echo hello (echo world !))",
	"hello world !",
	"",
}, {
	"(echo (echo the world) (echo is beautiful) !)",
	"the world is beautiful !",
	"",
}, {
	"+ 1 2",
	"",
	"1:1: unknown function \"+\"",
}, {
	"echo (+ 1 2)",
	"",
	"1:7: unknown function \"+\"",
}, {
	"echo ('echo' 1 2)",
	"",
	"1:7: operand must be a symbol",
}}

func testCall(resolve ResolveHandler, exp *SExp) (*SExp, error) {
	if exp.Str == "echo" {
		args, err := exp.Cdr.ResolveEvalEach(resolve, testCall)
		if err != nil {
			return nil, err
		}

		return &SExp{
			Type:   TypeStr,
			Str:    args.JoinCars(" ", false),
			Line:   exp.Line,
			Offset: exp.Offset,
		}, nil
	}

	return nil, errors.Errorf("%d:%d: unknown function %q", exp.Line, exp.Offset, exp.Str)
}

func TestSExp_eval(t *testing.T) {
	s := NewScanner()
	p := NewParser(s)

	for _, tt := range evalTests {
		head, err := p.Parse(tt.input)
		if err != nil {
			t.Errorf("%q: error: %s", tt.input, err)
			continue
		}

		var vals SExpSlice
		var v *SExp

		for ; head != nil; head = head.Cdr {
			v, err = testCall(ResolveName, head.List)
			if err != nil {
				break
			}

			vals = append(vals, v)
		}

		if err != nil {
			if tt.err != "" {
				if err.Error() != tt.err {
					t.Errorf("%q: error = %q want %q", tt.input, err, tt.err)
				}
			} else {
				t.Errorf("%q: error: %s", tt.input, err)
			}
			continue
		}

		output := vals.JoinCars("\n", false)

		if output != tt.output {
			t.Errorf("%q: output = %q want %q", tt.input, output, tt.output)
		}
	}
}
