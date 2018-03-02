// Copyright © 2017-2018 Stratumn SAS
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

// Package script defines types to implement a simple script interpreter with
// Lisp-like syntax. It has a bit of syntactic sugar to make it easier to read.
//
// It producces S-Expressions that can hold either symbols, strings, 64-bit
// integers or cons cells.
//
// It comes with a few builtin functions and it's easy to add new functions.
//
// It was designed with shell-like scripting in mind. A script is a list of
// instructions. Instructions are function calls. For convenience, top-level
// function calls don't need to be wrapped in parenthesis if they are
// one-liners. The is useful, especially in the context of a command line
// interface, but it introduces more complexity in the grammar because new
// lines are handled differently depending on context.
//
// The lifecycle of a script is:
//
//	Source Code -> Scanner -> Tokens
//	Tokens -> Parser -> S-Expression
//	S-Expression -> Interpreter -> Output
//
// At this point it's not designed for speed, but it does tail call
// optimizations (trampoline) to avoid blowing the stack with recursive calls.
//
// The current focus is to "harden" the type system so that scripts can be
// statically type-checked. Once the language is usable, a faster bytecode
// compiler will be implemented.
//
// Grammar
//
// The EBNF syntax is:
//
//	Script          = [ InBody ] { NewLine } EOF
//	InBody          = [ Instr ] [ InBodyTail ]
//	InBodyTail      = { NewLine { NewLine } Instr }
//	Instr           = ( Call | InCall )
//	Call            = "(" InCallInParen { NewLine } ")"
//	InCallInParen   = { NewLine } Symbol SExpListInParen
//	InCall          = Symbol SExpList
//	SExpListInParen = { SExpInParen }
//	SExpList        = { SExp }
//	SExpInParen     = { NewLine } SExp
//	SExp            = QuotedSExp | Body | List | Atom
//	QuotedSExp      = "'" { NewLine } SExp
//	QuasiquotedSExp = "`" { NewLine } SExp
//	UnquotedSExp    = "," { NewLine } SExp
//	Body            = "{" [ InBody ] { NewLine } "}"
//	List            = "(" SExpListInParen { NewLine } ")"
//	Atom            = symbol | string | int | true | false
//
//	TODO: define tokens
//
// Syntactic sugar
//
// Say you have a list of function calls like this:
//
//	((funcA a (cons b c))
//	 (funcB d e f)
//	 (funcC h i))
//
// You can use braces to open a "body" instead:
//
//	{
//		funcA a (cons b c)
//		funcB d e f
//		funcC hi
//	}
//
// Within a body the top-level instructions don't need to be surrounded by
// parenthesis. It can make code significantly easier to read. For example
// this code:
//
//	; Reverses a list recusively.
//	let reverse (lambda (l) (
//		; Define a nested recursive function with an accumulator.
//		(let reverse-rec (lambda (l tail) (
//			(if (nil? l)
//				tail
//				(reverse-rec (cdr l) (cons (car l) tail))))))
//		; Start the recursion
//		(reverse-rec l ())))
//
// Can be rewritten as:
//
//	; Reverses a list recusively with syntactic sugar.
//	let reverse (lambda (l) {
//		; Define a nested recursive function with an accumulator.
//		let reverse-rec (lamda (l tail) {
//			if (nil? l) tail else {
//				reverse-rec (cdr l) (cons (car l) tail)
//			}
//		})
//		; Start the recursion
//		reverse-rec l ()
//	})
//
// This example also used the optional "else" symbol in the if statement for
// clarity.
//
// The reason you don't need to use parenthesis at the top-level is because it
// uses the same rules as inside braces.
//
// There a a few limitations to using braces -- currently all top-level
// instructions have to be function calls, otherwise the grammar would be
// ambiguous.
package script

import "github.com/pkg/errors"

var (
	// ErrInvalidUTF8 is returned when the input contains an invalid UTF8
	// character.
	ErrInvalidUTF8 = errors.New("invalid UTF8 string")

	// ErrFuncName is returned when a function name is not a symbol.
	ErrFuncName = errors.New("function name is not a symbol")

	// ErrUnknownFunc is returned when a function is not found.
	ErrUnknownFunc = errors.New("unknown function")

	// ErrSymNotFound is returned when a symbol could not be resolved.
	ErrSymNotFound = errors.New("could not resolve symbol")

	// ErrNotFunc is returned when an S-Expression is not a function.
	ErrNotFunc = errors.New("the expression is not a function")

	// ErrInvalidCall is returned when an S-Expression is not a function
	// call.
	ErrInvalidCall = errors.New("invalid function call")

	// ErrDivByZero is returned when dividing by zero.
	ErrDivByZero = errors.New("division by zero")
)

const (
	// LambdaSymbol is the symbol used to declare lambda functions.
	LambdaSymbol = "lambda"

	// QuoteSymbol is the symbol used to quote expressions.
	QuoteSymbol = "quote"

	// QuasiquoteSymbol is the symbol used to quasiquote expressions.
	QuasiquoteSymbol = "quasiquote"

	// UnquoteSymbol is the symbol used to unquote expressions.
	UnquoteSymbol = "unquote"

	// ElseSymbol is the name of the optional "else" symbol in "if" and
	// "unless" statements.
	ElseSymbol = "else"
)
