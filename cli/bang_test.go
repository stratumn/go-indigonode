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
	"github.com/stratumn/alice/cli/script"
)

func TestBang(t *testing.T) {
	tests := []ExecTest{{
		"!echo hello outside world!",
		"hello outside world!\n",
		nil,
		nil,
	}, {
		"! echo",
		"\n",
		nil,
		nil,
	}, {
		"! echo 'hello outside world!'",
		"hello outside world!\n",
		nil,
		nil,
	}, {
		"! cat '' (title 'hello outside world!')",
		"Hello Outside World!",
		nil,
		nil,
	}, {
		"! cat () 'hello outside world!'",
		"hello outside world!",
		nil,
		nil,
	}, {
		"! grep outside 'hello\noutside\nworld!'",
		"outside\n",
		nil,
		nil,
	}, {
		"! (a)",
		"",
		script.ErrInvalidOperand,
		nil,
	}, {
		"!",
		"",
		ErrUse,
		nil,
	}, {
		"! echo a b c",
		"",
		ErrUse,
		nil,
	}, {
		"!1234567890",
		"",
		ErrAny,
		nil,
	}}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("%d-%s", i, tt.Command), func(t *testing.T) {
			tt.Test(t, cli.Bang{})
		})
	}
}

func TestBang_strings(t *testing.T) {
	bang := cli.Bang{}

	tests := []struct {
		name  string
		value string
	}{{
		"Name()",
		bang.Name(),
	}, {
		"Short()",
		bang.Short(),
	}, {
		"Long()",
		bang.Long(),
	}, {
		"Use()",
		bang.Use(),
	}, {
		"LongUse()",
		bang.LongUse(),
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.value == "" {
				t.Errorf("%s is blank", tt.value)
			}
		})
	}
}
