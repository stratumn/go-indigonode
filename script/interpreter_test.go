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
	`echo "hello"`,
	`"hello"`,
	"",
}, {
	"echo (echo)",
	`""`,
	"",
}, {
	`echo "hello" "world"`,
	`"hello world"`,
	"",
}, {
	`(echo "hello")`,
	`"hello"`,
	"",
}, {
	`echo "hello  world"`,
	`"hello  world"`,
	"",
}, {
	`echo "hello"
	(echo "world")`,
	`"hello"
"world"`,
	"",
}, {
	`(echo (echo "the world") (echo "is beautiful") "!")`,
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
	`echo ("echo" 1 2)`,
	"",
	`1:1: echo: 1:7: function name is not a symbol`,
}, {
	`echo "echo`,
	"",
	"1:6: unexpected token <invalid>",
}, {
	"echo hello",
	"",
	"1:1: echo: 1:6: hello: could not resolve symbol",
}, {
	`
	; Reverses a list recusively.
	let reverse (lambda (l) (
		; Define a nested recursive function with an accumulator.
		(let reverse-rec (lambda (l tail) (
			(if (nil? l)
				tail
				(reverse-rec (cdr l) (cons (car l) tail))))))
		; Start the recursion
		(reverse-rec l ())))
	
	reverse (quote (1 2 3 4 5 6 7 8 9 10))
	`,
	`(lambda (l) ((let reverse-rec (lambda (l tail) ((if (nil? l) tail (reverse-rec (cdr l) (cons (car l) tail)))))) (reverse-rec l <nil>)))
(10 9 8 7 6 5 4 3 2 1)`,
	"",
}, {
	`
	let reverse (lambda (l) (
		(let reverse-rec (lambda (l tail) (
			(let reverse-nested (lambda () (
				(if (nil? l)
					tail
					(reverse-rec (cdr l) (cons (car l) tail))))))
			(reverse-nested))))
		(reverse-rec l ())))
	
	reverse (quote (1 2 3 4 5 6 7 8 9 10))
	`,
	`(lambda (l) ((let reverse-rec (lambda (l tail) ((let reverse-nested (lambda <nil> ((if (nil? l) tail (reverse-rec (cdr l) (cons (car l) tail)))))) (reverse-nested)))) (reverse-rec l <nil>)))
(10 9 8 7 6 5 4 3 2 1)`,
	"",
}, {
	`
	let reverse (lambda (l) (
		(let start-rec (lambda () (
			(let reverse-rec (lambda (l tail) (
				(if (nil? l)
					tail
					(reverse-rec (cdr l) (cons (car l) tail))))))
			(reverse-rec l ()))))
		(start-rec)))
	
	reverse (quote (1 2 3 4 5 6 7 8 9 10))
	`,
	`(lambda (l) ((let start-rec (lambda <nil> ((let reverse-rec (lambda (l tail) ((if (nil? l) tail (reverse-rec (cdr l) (cons (car l) tail)))))) (reverse-rec l <nil>)))) (start-rec)))
(10 9 8 7 6 5 4 3 2 1)`,
	"",
}, {
	`
  	let reverse (lambda (l) (
  		(let reverse-rec-1 (lambda (l tail) (
  			(if (nil? l)
  				tail
  				(reverse-rec-2 (cdr l) (cons (car l) tail))))))
  		(let reverse-rec-2 (lambda (l tail) (
  			(if (nil? l)
  				tail
  				(reverse-rec-1 (cdr l) (cons (car l) tail))))))
  		(reverse-rec-1 l ())))

  	reverse (quote (1 2 3 4 5 6 7 8 9 10))
  	`,
	`(lambda (l) ((let reverse-rec-1 (lambda (l tail) ((if (nil? l) tail (reverse-rec-2 (cdr l) (cons (car l) tail)))))) (let reverse-rec-2 (lambda (l tail) ((if (nil? l) tail (reverse-rec-1 (cdr l) (cons (car l) tail)))))) (reverse-rec-1 l <nil>)))
(10 9 8 7 6 5 4 3 2 1)`,
	"",
}}

var testFuncs = map[string]InterpreterFuncHandler{
	"echo": func(ctx *InterpreterContext) (SExp, error) {
		args, err := ctx.EvalListToStrings(ctx, ctx.Args, false)
		if err != nil {
			return nil, err
		}

		str := strings.Join(args, " ")
		return String(str, ctx.Meta), nil
	},
}

func TestInterpreter(t *testing.T) {
	for _, tt := range evalTests {
		var got string

		closure := NewClosure()
		closure.Set("name", String("Alice", Meta{}))

		itr := NewInterpreter(
			InterpreterOptBuiltinLibs,
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
					t.Errorf("input\n%s:\nerror = %q want %q", tt.input, err, tt.err)
				}
			} else {
				t.Errorf("input\n%s:\nerror: %s", tt.input, err)
			}
			continue
		}

		if got != tt.want {
			t.Errorf("input\n%s\noutput\n%s\nwant\n%s", tt.input, got, tt.want)
		}
	}
}
