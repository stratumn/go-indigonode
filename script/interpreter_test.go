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
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
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
	"x 1 2",
	"",
	"1:1: x: unknown function",
}, {
	"echo (x 1 2)",
	"",
	"1:1: echo: 1:7: x: unknown function",
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
	
	reverse '(1 2 3 4 5 6 7 8 9 10)
	`,
	`(lambda (l) ((let reverse-rec (lambda (l tail) ((if (nil? l) tail (reverse-rec (cdr l) (cons (car l) tail)))))) (reverse-rec l ())))
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
	
	reverse '(1 2 3 4 5 6 7 8 9 10)
	`,
	`(lambda (l) ((let reverse-rec (lambda (l tail) ((let reverse-nested (lambda () ((if (nil? l) tail (reverse-rec (cdr l) (cons (car l) tail)))))) (reverse-nested)))) (reverse-rec l ())))
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
	
	reverse '(1 2 3 4 5 6 7 8 9 10)
	`,
	`(lambda (l) ((let start-rec (lambda () ((let reverse-rec (lambda (l tail) ((if (nil? l) tail (reverse-rec (cdr l) (cons (car l) tail)))))) (reverse-rec l ())))) (start-rec)))
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

  	reverse '(1 2 3 4 5 6 7 8 9 10)
  	`,
	`(lambda (l) ((let reverse-rec-1 (lambda (l tail) ((if (nil? l) tail (reverse-rec-2 (cdr l) (cons (car l) tail)))))) (let reverse-rec-2 (lambda (l tail) ((if (nil? l) tail (reverse-rec-1 (cdr l) (cons (car l) tail)))))) (reverse-rec-1 l ())))
(10 9 8 7 6 5 4 3 2 1)`,
	"",
}, {
	"echo `(2 ,(+ 1 2) 4)",
	`"(2 3 4)"`,
	"",
}, {
	"let letter `a\n" +
		"echo `(letter ,letter ',letter)",
	`a
"(letter a (quote a))"`,
	"",
}}

func testEcho(ctx *InterpreterContext) (SExp, error) {
	args, err := ctx.EvalListToStrings(ctx, ctx.Args, false)
	if err != nil {
		return nil, err
	}

	str := strings.Join(args, " ")
	return String(str, ctx.Meta), nil
}

func TestInterpreter(t *testing.T) {
	testInterpreter(t)
}

func TestInterpreter_no_tail_opts(t *testing.T) {
	testInterpreter(t, InterpreterOptTailOptimizations(false))
}

func testInterpreter(t *testing.T, opts ...InterpreterOpt) {
	for _, tt := range evalTests {
		var got string

		closure := NewClosure()
		closure.Set("name", String("Alice", Meta{}))

		opts = append(
			opts,
			InterpreterOptClosure(closure),
			InterpreterOptBuiltinLibs,
			InterpreterOptErrorHandler(func(error) {}),
			InterpreterOptValueHandler(func(exp SExp) {
				if got != "" {
					got += "\n"
				}
				got += exp.String()
			}),
		)

		itr := NewInterpreter(opts...)
		itr.AddFuncHandler("echo", testEcho)

		err := itr.EvalInput(context.Background(), tt.input)
		if err != nil {
			if tt.err != "" {
				assert.Equal(t, tt.err, err.Error())
			} else {
				assert.NoError(t, err)
			}
			continue
		}

		assert.Equal(t, tt.want, got)
	}
}
