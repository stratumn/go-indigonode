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
	Name:     "if",
	Use:      "if (command) then else",
	ExecSExp: ifExec,
}}

func ifExec(
	ctx context.Context,
	cli CLI, w io.Writer,
	exec script.SExpExecutor,
	sexp *script.SExp,
) error {
	cond := sexp.Cdr
	if cond == nil {
		return NewUseError("missing condition expression")
	}

	if cond.Type != script.SExpList {
		return NewUseError("condition must be a list")
	}

	then := cond.Cdr
	if then == nil {
		return NewUseError("missing then expression")
	}

	otherwise := then.Cdr

	_, err := exec(cond.SExp)

	if err == nil {
		if then.Type == script.SExpList {
			s, err := exec(then.SExp)
			fmt.Fprint(w, s)
			return err
		}

		fmt.Fprintln(w, then.Str)

		return nil
	}

	if otherwise != nil {
		if otherwise.Type == script.SExpList {
			s, err := exec(otherwise.SExp)
			fmt.Fprint(w, s)
			return err
		}

		fmt.Fprintln(w, otherwise.Str)

		return nil
	}

	return nil
}
