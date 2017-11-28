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

package cli_test

import (
	"fmt"
	"testing"

	"github.com/stratumn/alice/cli"
)

func TestIf(t *testing.T) {
	tt := []ExecTest{{
		"if (title test) ok ko",
		"ok\n",
		nil,
		nil,
	}, {
		"if (test) ok ko",
		"ko\n",
		nil,
		nil,
	}, {
		"if (title test) (title ok) (title ko)",
		"Ok\n",
		nil,
		nil,
	}, {
		"if (test) (title ok) (title ko)",
		"Ko\n",
		nil,
		nil,
	}, {
		"if (title test) ok",
		"ok\n",
		nil,
		nil,
	}, {
		"if (test) ok",
		"",
		nil,
		nil,
	}, {
		"if () ok ko",
		"ko\n",
		nil,
		nil,
	}, {
		"if",
		"",
		ErrUse,
		nil,
	}, {
		"if test",
		"",
		ErrUse,
		nil,
	}, {
		"if (title test)",
		"",
		ErrUse,
		nil,
	}}

	for i, test := range tt {
		t.Run(fmt.Sprintf("%d-%s", i, test.Command), func(t *testing.T) {
			test.TestSExp(t, cli.If.Cmd)
		})
	}
}
