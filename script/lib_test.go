// Copyright © 2017-2018 Stratumn SAS
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as
// published by the Free Software Foundation, either version 3 of the
// License, or (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <https://www.gnu.org/licenses/>.

package script

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

type libTest struct {
	input string
	want  string
	err   string
}

func testLib(t *testing.T, lib map[string]InterpreterFuncHandler, tests []libTest) {
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			var got string

			itr := NewInterpreter(
				InterpreterOptFuncHandlers(LibOp),
				InterpreterOptFuncHandlers(LibMeta),
				InterpreterOptFuncHandlers(lib),
				InterpreterOptErrorHandler(func(error) {}),
				InterpreterOptValueHandler(func(exp SExp) {
					if exp == nil {
						return
					}
					if got != "" {
						got += "\n"
					}
					got += exp.String()
				}),
			)

			err := itr.EvalInput(context.Background(), tt.input)
			if err != nil {
				if tt.err != "" {
					assert.Equal(t, tt.err, err.Error())
				} else {
					assert.NoError(t, err)
				}
				return
			} else if tt.err != "" {
				assert.Fail(t, tt.err)
			}

			assert.Equal(t, tt.want, got)
		})
	}
}
