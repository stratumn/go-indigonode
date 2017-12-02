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
	"context"

	"github.com/stratumn/alice/cli/script"
)

// Lambda is a command that creates an anomymous function.
var Lambda = BasicCmdWrapper{BasicCmd{
	Name:     "lambda",
	Use:      "lambda (args...) (body...)",
	Short:    "Create anonymous function",
	ExecSExp: lambdaExec,
}}

func lambdaExec(
	ctx context.Context,
	cli CLI,
	closure *script.Closure,
	call script.CallHandler,
	exp *script.SExp,
) (*script.SExp, error) {
	args := exp.Cdr
	if args == nil {
		return nil, NewUseError("missing arguments")
	}

	if args.Type != script.TypeList {
		return nil, NewUseError("expected arguments to be a list")
	}

	for head := args.List; head != nil; head = head.Cdr {
		if head.Type != script.TypeSym {
			return nil, NewUseError("expected argument to be a symbol")
		}
	}

	body := args.Cdr
	if body == nil {
		return nil, NewUseError("missing body")
	}

	if body.Type != script.TypeList {
		return nil, NewUseError("expected body to be a list")
	}

	lambda := &script.SExp{
		Type:     script.TypeList,
		List:     exp.Clone(),
		Line:     exp.Line,
		Offset:   exp.Offset,
		UserData: FuncData{closure},
	}

	return lambda, nil
}
