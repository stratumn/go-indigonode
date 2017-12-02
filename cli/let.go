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

// Let is a command that binds a symbol to a value
var Let = BasicCmdWrapper{BasicCmd{
	Name:     "let",
	Use:      "let sym value",
	Short:    "Bind a symbol to a value",
	ExecSExp: letExec,
}}

func letExec(
	ctx context.Context,
	cli CLI,
	closure *script.Closure,
	call script.CallHandler,
	exp *script.SExp,
) (*script.SExp, error) {
	sym := exp.Cdr
	if sym == nil || sym.Type != script.TypeSym {
		return nil, NewUseError("expected a symbol")
	}

	val := sym.Cdr

	if val == nil {
		closure.SetLocal(sym.Str, nil)
		return nil, nil
	}

	v, err := val.ResolveEval(closure.Resolve, call)
	if err != nil {
		return nil, err
	}

	closure.SetLocal("$"+sym.Str, v)

	return v, nil
}
