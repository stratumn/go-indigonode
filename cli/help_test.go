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
	"github.com/stratumn/alice/cli/mockcli"
)

func TestHelp(t *testing.T) {
	cmds := []cli.Cmd{
		cli.BasicCmdWrapper{
			Cmd: cli.BasicCmd{
				Name:  "cmd1",
				Short: "Command 1",
			},
		},
		cli.BasicCmdWrapper{
			Cmd: cli.BasicCmd{
				Name:  "cmd2",
				Short: "Command 2",
			},
		},
	}

	tt := []ExecTest{{
		"help",
		"cmd1  Command 1\ncmd2  Command 2\n",
		nil,
		func(c *mockcli.MockCLI) {
			c.EXPECT().Commands().Return(cmds).AnyTimes()
		},
	}, {
		"help cmd2",
		`Command 2

Usage:
  cmd2

Flags:
  -h, --help   Invoke help on command
`,
		nil,
		func(c *mockcli.MockCLI) {
			c.EXPECT().Commands().Return(cmds).AnyTimes()
		},
	}, {
		"help earth moon",
		"",
		ErrUse,
		nil,
	}, {
		"help cmd3",
		"",
		ErrAny,
		func(c *mockcli.MockCLI) {
			c.EXPECT().Commands().Return(cmds).AnyTimes()
		},
	}}

	for i, test := range tt {
		t.Run(fmt.Sprintf("%d-%s", i, test.Command), func(t *testing.T) {
			test.TestStrings(t, cli.Help.Cmd)
		})
	}
}
