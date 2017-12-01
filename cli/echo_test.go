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

func TestEcho(t *testing.T) {
	tests := []ExecTest{{
		"echo hello world",
		"hello world",
		nil,
		nil,
	}, {
		"echo --log debug hello world",
		"",
		nil,
		nil,
	}, {
		"echo --log normal hello world",
		"hello world",
		nil,
		nil,
	}, {
		"echo --log info hello world",
		"hello world",
		nil,
		nil,
	}, {
		"echo --log success hello world",
		"hello world",
		nil,
		nil,
	}, {
		"echo --log warning hello world",
		"hello world",
		nil,
		nil,
	}, {
		"echo --log error hello world",
		"hello world",
		nil,
		nil,
	}}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("%d-%s", i, tt.Command), func(t *testing.T) {
			tt.Test(t, cli.Echo)
		})
	}
}
