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

package cli

import (
	"github.com/pkg/errors"
	"github.com/stratumn/alice/cli/script"
)

// stackTracer is used to get the stack trace from errors created by the
// github.com/pkg/errors package.
type stackTracer interface {
	StackTrace() errors.StackTrace
}

// StackTrace returns the stack trace from an error created using the
// github.com/pkg/errors package.
func StackTrace(err error) errors.StackTrace {
	if err, ok := err.(stackTracer); ok {
		return err.StackTrace()
	}

	return nil
}

// evalSExpBody evaluates an S-Expression 'body'.
//
// If the expression is a list beginning with a list, it evaluates each
// expression in the list and returns the last value.
//
// Otherwise the expression is evaluated directly.
func evalSExpBody(
	resolve script.ResolveHandler,
	call script.CallHandler,
	body script.SExp,
) (script.SExp, error) {
	evalDirectly := func() (script.SExp, error) {
		return script.Eval(resolve, call, body)
	}

	// Eval directly unless body is a list or car is a list.

	if body == nil {
		return evalDirectly()
	}

	cell, ok := body.CellVal()
	if !ok || !cell.IsList() {
		return evalDirectly()
	}

	car := cell.Car()
	if car == nil {
		return evalDirectly()
	}

	carCell, ok := car.CellVal()
	if !ok || !carCell.IsList() {
		return evalDirectly()
	}

	// Eval each expression in the body list.
	vals, err := script.EvalListToSlice(resolve, call, cell)
	if err != nil {
		return nil, err
	}

	// Return the last evaluated expression if there is one.
	l := len(vals)
	if l > 0 {
		return vals[l-1], nil
	}

	return nil, nil
}
