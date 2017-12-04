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
	"github.com/stratumn/alice/cli/script"
)

// Cdr is a command that prints the tail of a list.
var Cdr = BasicCmdWrapper{BasicCmd{
	Name:     "cdr",
	Use:      "cdr <Cell>",
	Short:    "Output the cdr of a cons cell",
	ExecSExp: cdr,
}}

func cdr(ctx *ExecContext) (script.SExp, error) {
	// Return the cdr of the evaluated car cell.

	if ctx.Args == nil || ctx.Args.Cdr() != nil {
		return nil, NewUseError("expected a single element")
	}

	v, err := script.Eval(ctx.Closure.Resolve, ctx.Call, ctx.Args.Car())
	if err != nil {
		return nil, err
	}

	if v == nil {
		return nil, nil
	}

	cell, ok := v.CellVal()
	if !ok {
		return nil, NewUseError("expected a cell")
	}

	return cell.Cdr(), nil
}
