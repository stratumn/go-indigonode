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

package script_test

import (
	"context"
	"strings"

	"github.com/stratumn/go-node/script"
)

func Example_simple() {
	// Script that will be evaluated.
	src := `
		; You can choose not to evaluate a list by quoting it:
		let my-list '(A B C)

		; You can do conditional expressions:
		if (list? my-list) "it's a list" "it's not a list"

		; You can manipulate lists:
		car my-list
		cdr my-list
		car (cdr my-list)

		; You can create functions:
		let cadr (lambda (l) (car (cdr l)))
		cadr my-list

		; By the way can use parenthesis at the top level if you want:
		(cadr my-list)
`

	// Initialize an interpreter with the builtin libraries.
	itr := script.NewInterpreter(script.InterpreterOptBuiltinLibs)

	// Evaluate the script.
	err := itr.EvalInput(context.Background(), src)
	if err != nil {
		panic(err)
	}

	// Output:
	// (A B C)
	// "it's a list"
	// A
	// (B C)
	// B
	// (lambda (l) (car (cdr l)))
	// B
	// B
}

func Example_recursion() {
	// Script that will be evaluated.
	src := `
		; Reverses a list recusively.
		let reverse (lambda (l) (
			; Define a nested recursive function with an accumulator.
			(let reverse-rec (lambda (l tail) (
				(if (nil? l) 
					tail
					(reverse-rec (cdr l) (cons (car l) tail))))))
			; Start the recursion
			(reverse-rec l ())))

		; Test it out.
		reverse '(A L I C E)
`

	// Initialize an interpreter with the builtin libraries.
	itr := script.NewInterpreter(script.InterpreterOptBuiltinLibs)

	// Evaluate the script.
	err := itr.EvalInput(context.Background(), src)
	if err != nil {
		panic(err)
	}

	// Output:
	// (lambda (l) ((let reverse-rec (lambda (l tail) ((if (nil? l) tail (reverse-rec (cdr l) (cons (car l) tail)))))) (reverse-rec l ())))
	// (E C I L A)
}

func Example_recursion_braces() {
	// Script that will be evaluated.
	src := `
		; Reverses a list recusively with syntactic sugar.
		let reverse (λ (l) {
			; Define a nested recursive function with an accumulator.
			let reverse-rec (λ (l tail) {
				if (nil? l) tail else {
					reverse-rec (cdr l) (cons (car l) tail)
				}
			})
			; Start the recursion
			reverse-rec l ()
		})

		; Test it out.
		reverse '(A L I C E)
`

	// Initialize an interpreter with the builtin libraries.
	itr := script.NewInterpreter(script.InterpreterOptBuiltinLibs)

	// Evaluate the script.
	err := itr.EvalInput(context.Background(), src)
	if err != nil {
		panic(err)
	}

	// Output:
	// (lambda (l) ((let reverse-rec (λ (l tail) ((if (nil? l) tail else ((reverse-rec (cdr l) (cons (car l) tail))))))) (reverse-rec l ())))
	// (E C I L A)
}

func Example_fib() {
	// Script that will be evaluated.
	src := `
		; Computes the nth element of the Fibonacci sequence.
		let fib (lambda (n) (
			(let fib-rec (lambda (n f1 f2) (
				(if (< n 1)
					f1
					(fib-rec (- n 1) f2 (+ f1 f2))))))
			(fib-rec n 0 1)))

		; Test it out.
		fib 10
`

	// Initialize an interpreter with the builtin libraries.
	itr := script.NewInterpreter(script.InterpreterOptBuiltinLibs)

	// Evaluate the script.
	err := itr.EvalInput(context.Background(), src)
	if err != nil {
		panic(err)
	}

	// Output:
	// (lambda (n) ((let fib-rec (lambda (n f1 f2) ((if (< n 1) f1 (fib-rec (- n 1) f2 (+ f1 f2)))))) (fib-rec n 0 1)))
	// 55
}

func Example_search_tree() {
	// Script that will be evaluated.
	src := `
		let search-tree (lambda (tree fn) (
			(unless (nil? tree)
				(if (atom? tree) 
					(if (fn tree) tree)
					((let match (search-tree (car tree) fn))
					 (unless (nil? match)
						match
						(search-tree (cdr tree) fn)))))))
						
		(search-tree
			'(1 2 (3 (5 6) (7 8)) (9 10)) 
			(lambda (leaf) (= (mod leaf 4) 0)))
`

	// Initialize an interpreter with the builtin libraries.
	itr := script.NewInterpreter(script.InterpreterOptBuiltinLibs)

	// Evaluate the script.
	err := itr.EvalInput(context.Background(), src)
	if err != nil {
		panic(err)
	}

	// Output:
	// (lambda (tree fn) ((unless (nil? tree) (if (atom? tree) (if (fn tree) tree) ((let match (search-tree (car tree) fn)) (unless (nil? match) match (search-tree (cdr tree) fn)))))))
	// 8
}

func Example_customFunctions() {
	// Script that will be evaluated.
	src := `
		echo (title "hello world!")

		(echo
			(title "goodbye")
			(title "world!"))
`

	// Define custom functions for the interpreter.
	funcs := map[string]script.InterpreterFuncHandler{
		"echo": func(ctx *script.InterpreterContext) (script.SExp, error) {
			// Evaluate the arguments to strings.
			args, err := ctx.EvalListToStrings(ctx, ctx.Args, false)
			if err != nil {
				return nil, err
			}

			// Join the argument strings.
			str := strings.Join(args, " ")

			// Return a string value.
			return script.String(str, ctx.Meta), nil
		},
		"title": func(ctx *script.InterpreterContext) (script.SExp, error) {
			// Evaluate the arguments to strings.
			args, err := ctx.EvalListToStrings(ctx, ctx.Args, false)
			if err != nil {
				return nil, err
			}

			// Join the argument strings and convert it to a title.
			title := strings.Title(strings.Join(args, " "))

			// Return a string value.
			return script.String(title, ctx.Meta), nil
		},
	}

	// Initialize an interpreter with the custom functions.
	itr := script.NewInterpreter(script.InterpreterOptFuncHandlers(funcs))

	// Evaluate the script.
	err := itr.EvalInput(context.Background(), src)
	if err != nil {
		panic(err)
	}

	// Output:
	// "Hello World!"
	// "Goodbye World!"
}
