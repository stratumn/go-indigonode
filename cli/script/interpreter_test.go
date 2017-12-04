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
	"context"
	"strings"
	"testing"
)

type evalTest struct {
	input string
	want  string
	err   string
}

var evalTests = []evalTest{{
	"",
	"",
	"",
}, {
	"echo",
	`""`,
	"",
}, {
	"echo ()",
	`""`,
	"",
}, {
	"echo 'hello'",
	`"hello"`,
	"",
}, {
	"echo (echo)",
	`""`,
	"",
}, {
	"echo 'hello' 'world'",
	`"hello world"`,
	"",
}, {
	"(echo 'hello')",
	`"hello"`,
	"",
}, {
	"echo 'hello  world'",
	`"hello  world"`,
	"",
}, {
	"echo 'hello'\n(echo 'world')",
	`"hello"
"world"`,
	"",
}, {
	"(echo (echo 'the world') (echo 'is beautiful') '!')",
	`"the world is beautiful !"`,
	"",
}, {
	"echo name",
	`"Alice"`,
	"",
}, {
	"+ 1 2",
	"",
	"1:1: +: unknown function",
}, {
	"echo (+ 1 2)",
	"",
	"1:1: echo: 1:7: +: unknown function",
}, {
	"echo ('echo' 1 2)",
	"",
	"1:1: echo: 1:7: \"echo\": function name is not a symbol",
}, {
	"echo 'echo",
	"",
	"1:6: unexpected token <invalid>",
}, {
	"echo hello",
	"",
	"1:1: echo: hello: could not resolve symbol",
}}

var testFuncs = map[string]InterpreterFuncHandler{
	"echo": func(ctx *InterpreterContext) (SExp, error) {
		args, err := ctx.EvalListToStrings(ctx.Ctx, ctx.Closure, ctx.Args)
		if err != nil {
			return nil, err
		}

		str := strings.Join(args, " ")
		return String(str, ctx.Meta), nil
	},
}

func TestInterpreter(t *testing.T) {
	closure := NewClosure()
	closure.Set("name", String("Alice", Meta{}))

	for _, tt := range evalTests {
		var got string

		itr := NewInterpreter(
			InterpreterOptClosure(closure),
			InterpreterOptFuncHandlers(testFuncs),
			InterpreterOptErrorHandler(func(error) {}),
			InterpreterOptValueHandler(func(exp SExp) {
				if got != "" {
					got += "\n"
				}
				got += exp.String()
			}),
		)

		err := itr.EvalInput(context.Background(), tt.input)
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

		if got != tt.want {
			t.Errorf("%q: output = %q want %q", tt.input, got, tt.want)
		}
	}
}
