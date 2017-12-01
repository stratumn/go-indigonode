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

	"github.com/pkg/errors"
	"github.com/stratumn/alice/cli/script"
)

// Cdr is a command that prints the tail of a list.
var Cdr = BasicCmdWrapper{BasicCmd{
	Name:     "cdr",
	Use:      "cdr (quote (a b c))",
	Short:    "Output the tail a list",
	ExecSExp: cdr,
}}

func cdr(
	ctx context.Context,
	cli CLI,
	closure *script.Closure,
	call script.CallHandler,
	exp *script.SExp,
) (*script.SExp, error) {
	if exp.Cdr == nil || exp.Cdr.Cdr != nil {
		return nil, NewUseError("expected a single expression")
	}

	v, err := exp.Cdr.ResolveEval(closure.Resolve, call)
	if err != nil {
		return nil, err
	}

	if v.Type != script.TypeList {
		return nil, NewUseError("expected a list")
	}

	if v.List == nil {
		return nil, errors.WithStack(ErrCdrNil)
	}

	if v.List.Cdr == nil {
		return nil, errors.WithStack(ErrCdrNil)
	}

	return &script.SExp{
		Type:   script.TypeList,
		List:   v.List.Cdr,
		Line:   exp.Line,
		Offset: exp.Offset,
	}, nil
}
