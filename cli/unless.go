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
	"io"

	"github.com/stratumn/alice/cli/script"
)

// Unless is a command that does the opposite of the If command.
var Unless = BasicCmdWrapper{BasicCmd{
	Name:     "unless",
	Use:      "unless cond then [else]",
	Short:    "Opposite of the if command",
	ExecSExp: unlessExec,
}}

func unlessExec(
	ctx context.Context,
	cli CLI, w io.Writer,
	closure *script.Closure,
	eval script.Evaluator,
	exp *script.SExp,
) error {
	cond := exp.Cdr
	if cond == nil {
		return NewUseError("missing condition expression")
	}

	then := cond.Cdr
	if then == nil {
		return NewUseError("missing then expression")
	}

	_, err := cond.ResolveEval(closure.Resolve, eval)

	if err != nil {
		return evalSExpBody(w, closure.Resolve, eval, then)
	}

	if then.Cdr != nil {
		return evalSExpBody(w, closure.Resolve, eval, then.Cdr)
	}

	return nil
}
