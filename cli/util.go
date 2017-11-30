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
	"fmt"
	"io"

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

// evalSExpBody evaluates an S-Expression "body".
//
// If the car of the expression is a list beginning with a list, it evaluates
// each expression in that list and outputs the last value to the writer.
//
// If the car isn't a list, it is directly evaluated.
func evalSExpBody(
	w io.Writer,
	resolve script.Resolver,
	eval script.Evaluator,
	body *script.SExp,
) error {
	exp := body

	isList := func(x *script.SExp) bool {
		return x.Type == script.TypeList
	}

	if isList(exp) && exp.List != nil && isList(exp.List) {
		// Ignore output of every expression in the list except the
		// last one
		for exp = exp.List; exp.Cdr != nil; exp = exp.Cdr {
			_, err := exp.ResolveEval(resolve, eval)
			if err != nil {
				return err
			}
		}
	}

	// At this point exp is either the given expression or the last
	// expression in the list. In any case we want to write its output.
	s, err := exp.ResolveEval(resolve, eval)
	if err == nil {
		fmt.Fprintln(w, s)
	}

	return err
}
