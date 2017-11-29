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
	"fmt"
	"io"

	"github.com/stratumn/alice/cli/script"
)

// If is a command that evaluates conditional expressions.
var If = BasicCmdWrapper{BasicCmd{
	Name:      "if",
	Use:       "if cond then [else]",
	Short:     "Evaluate conditional expressions",
	ExecInstr: ifExec,
}}

func ifExec(
	ctx context.Context,
	cli CLI, w io.Writer,
	closure *script.Closure,
	eval script.SExpEvaluator,
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

	otherwise := then.Cdr

	_, err := cond.ResolveEval(closure.Resolve, eval)

	if err == nil {
		s, err := then.ResolveEval(closure.Resolve, eval)
		if err == nil {
			fmt.Fprintln(w, s)
		}

		return err
	}

	if otherwise != nil {
		s, err := otherwise.ResolveEval(closure.Resolve, eval)
		if err == nil {
			fmt.Fprintln(w, s)
		}

		return err
	}

	return nil
}
