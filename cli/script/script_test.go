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

package script_test

import (
	"context"
	"strings"

	"github.com/stratumn/alice/cli/script"
)

func Example() {
	// Script that will be evaluated.
	src := `
		; This is a comment.

		echo (title "hello world!")

		(echo
			(title "goodbye")
			(title "world!"))
`

	// Define custom functions for the interpreter.
	funcs := map[string]script.InterpreterFuncHandler{
		"echo": func(ctx *script.InterpreterContext) (script.SExp, error) {
			// Evaluate the arguments to strings.
			args, err := ctx.EvalListToStrings(ctx.Ctx, ctx.Closure, ctx.Args)
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
			args, err := ctx.EvalListToStrings(ctx.Ctx, ctx.Closure, ctx.Args)
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
