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

// Package script defines types to implement a simple script interpreter with
// Lisp-like syntax.
//
// It producces S-Expressions that can hold either symbols, strings, 64-bit
// integers or cons cells.
//
// It was designed with shell-like scripting in mind. A script is a list of
// instructions. Instructions are function calls. For convenience, top-level
// function calls don't need to be wrapped in parenthesis if they are
// one-liners. The is useful, especially in the context of a command line
// interface, but it introduces more complexity in the grammar because new
// lines are handled differently depending on context.
//
// The EBNF syntax is:
//
//	Script          = { Instr }
//	Instr           = { NewLine } ( Call | InCall )
//	Call            = { NewLine } "(" InCallInParen { NewLine } ")"
//	InCallInParen   = { NewLine } Symbol SExpListInParen
//	InCall          = Symbol SExpList
//	SExpListInParen = { SExpInParen }
//	SExpList        = { SExp }
//	SExpInParen     = { NewLine } SExp
//	SExp            = List | Atom
//	List            = { NewLine } "(" SExpListInParen { NewLine } ")"
//	Atom            = symbol | string | int | "true" | "false"
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
// Eventually there will be a bytecode compiler and static type checking, but
// the current focus is on the design of the language.
//
// It comes with a few builtin functions and it's easy to add new functions.
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

// LambdaSymbol is the symbol used to declare lambda functions.
const LambdaSymbol = "lambda"
