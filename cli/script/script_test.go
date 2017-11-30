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

	var eval script.Evaluator

	eval = func(resolve script.Resolver, exp *script.SExp) (string, error) {
		if exp.Type == script.TypeStr {
			return exp.Str, nil
		}

		args, err := exp.Cdr.ResolveEvalEach(resolve, eval)
		if err != nil {
			return "", err
		}

		str := strings.Join(args, " ")

		switch exp.Str {
		case "echo":
			fmt.Println(str)
			return "", nil
		case "title":
			return strings.Title(str), nil
		}

		return "", fmt.Errorf("invalid operand: %q", exp.Str)
	}

	printErr := func(err error) {
		panic(err)
	}

	scanner := script.NewScanner(script.OptErrorHandler(printErr))
	parser := script.NewParser(scanner)

	head, err := parser.Parse(src)
	if err != nil {
		panic(err)
	}

	for ; head != nil; head = head.Cdr {
		_, err = eval(script.ResolveName, head.List)
		if err != nil {
			panic(err)
		}
	}

	// Output:
	// Hello World!
	// Goodbye World!
}
