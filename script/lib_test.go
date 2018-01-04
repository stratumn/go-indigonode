// Copyright © 2017-2018 Stratumn SAS
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
		var got string

		itr := NewInterpreter(
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
			continue
		} else if tt.err != "" {
			assert.Fail(t, tt.err)
		}

		assert.Equal(t, tt.want, got)
	}
}
