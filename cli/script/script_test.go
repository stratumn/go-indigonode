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

	var exec script.SExpExecutor

	exec = func(instr *script.SExp) (string, error) {
		args, err := instr.Cdr.EvalEach(exec)
		if err != nil {
			return "", err
		}

		str := strings.Join(args, " ")

		switch instr.Str {
		case "echo":
			fmt.Println(str)
			return "", nil
		case "title":
			return strings.Title(str), nil
		}

		return "", fmt.Errorf("invalid operand: %q", instr.Str)
	}

	printErr := func(err error) {
		panic(err)
	}

	scanner := script.NewScanner(script.ScannerOptErrorHandler(printErr))
	parser := script.NewParser(scanner)

	head, err := parser.Parse(src)
	if err != nil {
		panic(err)
	}

	for ; head != nil; head = head.Cdr {
		_, err = exec(head.List)
		if err != nil {
			panic(err)
		}
	}

	// Output:
	// Hello World!
	// Goodbye World!
}
