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

	call = func(resolve script.ResolveHandler, exp *script.SExp) (*script.SExp, error) {
		args, err := exp.Cdr.ResolveEvalEach(resolve, call)
		if err != nil {
			return nil, err
		}

		str := args.JoinCars(" ", false)

		switch exp.Str {
		case "echo":
			fmt.Println(str)
			return nil, nil
		case "title":
			return &script.SExp{
				Type:   script.TypeStr,
				Str:    strings.Title(str),
				Line:   exp.Line,
				Offset: exp.Offset,
			}, nil
		}

		return nil, errors.Errorf("%d:%d: unknown function %q", exp.Line, exp.Offset, exp.Str)
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
		_, err = call(script.ResolveName, head.List)
		if err != nil {
			panic(err)
		}
	}

	// Output:
	// Hello World!
	// Goodbye World!
}
