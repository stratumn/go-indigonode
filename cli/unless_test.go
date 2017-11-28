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

func TestUnless(t *testing.T) {
	tt := []ExecTest{{
		"unless (title test) ko ok",
		"ok\n",
		nil,
		nil,
	}, {
		"unless (test) ko ok",
		"ko\n",
		nil,
		nil,
	}, {
		"unless (title test) (title ko) (title ok)",
		"Ok\n",
		nil,
		nil,
	}, {
		"unless (test) (title ko) (title ok)",
		"Ko\n",
		nil,
		nil,
	}, {
		"unless (title test) ko",
		"",
		nil,
		nil,
	}, {
		"unless (test) ko",
		"ko\n",
		nil,
		nil,
	}, {
		"unless",
		"",
		ErrUse,
		nil,
	}, {
		"unless test",
		"",
		ErrUse,
		nil,
	}, {
		"unless (title test)",
		"",
		ErrUse,
		nil,
	}}

	for i, test := range tt {
		t.Run(fmt.Sprintf("%d-%s", i, test.Command), func(t *testing.T) {
			test.TestInstr(t, cli.Unless.Cmd)
		})
	}
}
