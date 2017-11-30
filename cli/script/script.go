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

// Package script defines types to implement a very rudimentary script
// interpreter.
//
// It is designed with shell-like scripting in mind. It uses simplified
// S-Expressions that can hold either symbols, strings or lists. Everything
// evaluates to a string instead of S-Expressions.
//
// There are no builtin operators.
package script

import "github.com/pkg/errors"

var (
	// ErrInvalidOperand is returned when an operand is not a symbol.
	ErrInvalidOperand = errors.New("operand must be a symbol")

	// ErrSymNotFound is returned when a symbol could not be resolved.
	ErrSymNotFound = errors.New("could not resolve symbol")
)
