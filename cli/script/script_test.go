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
	"fmt"
	"strings"

	"github.com/pkg/errors"
	"github.com/stratumn/alice/cli/script"
)

func Example() {
	src := `
		; This is a comment.

		echo (title hello world!)

		(echo
			(title goodbye)
			(title world!))
`

	var call script.CallHandler

	// call takes care of evaluating function calls.
	call = func(
		resolve script.ResolveHandler,
		name string,
		args script.SCell,
		meta script.Meta,
	) (script.SExp, error) {
		// In this case we just evaluable and join the arguments.
		argv, err := script.EvalListToStrings(resolve, call, args)
		if err != nil {
			return nil, err
		}

		str := strings.Join(argv, " ")

		switch name {
		case "echo":
			return script.String(str, meta), nil
		case "title":
			title := strings.Title(str)
			return script.String(title, meta), nil
		}

		return nil, errors.Errorf(
			"%d:%d: %s: unknown function",
			meta.Line,
			meta.Offset,
			name,
		)
	}

	printErr := func(err error) {
		panic(err)
	}

	scanner := script.NewScanner(script.OptErrorHandler(printErr))
	parser := script.NewParser(scanner)

	// Parser returns a list of instructions.
	list, err := parser.Parse(src)
	if err != nil {
		panic(err)
	}

	// Evaluate each expressions in the list.
	vals, err := script.EvalListToStrings(script.ResolveName, call, list)
	if err != nil {
		panic(err)
	}

	fmt.Println(strings.Join(vals, "\n"))

	// Output:
	// Hello World!
	// Goodbye World!
}
