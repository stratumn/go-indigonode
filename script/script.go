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
// It producces S-Expressions that can hold either symbols, strings, or cons
// cells.
//
// It is designed with shell-like scripting in mind. A script is a list of
// instructions. Instructions are function calls. For convenience, top-level
// function calls don't need to be wrapped in parenthesis if they are
// one-liners. The is useful, especially in the context of a command line
// interface, but it introduces more complexity in the grammar because new
// lines are meaningful.
//
// The syntax is:
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
//	Atom            = symbol | string
package script

import "github.com/pkg/errors"

var (
	// ErrInvalidUTF8 is returned when a strings contains an invalid UTF8
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
)

// LambdaSymbol is the symbol used to declare lambda functions.
const LambdaSymbol = "lambda"
