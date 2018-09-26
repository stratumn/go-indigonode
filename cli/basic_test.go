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

package cli

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/spf13/pflag"
	"github.com/stratumn/go-node/script"
	"github.com/stretchr/testify/assert"
)

type basicContentMock string

func (s basicContentMock) TextBeforeCursor() string {
	return string(s)
}

func (s basicContentMock) GetWordBeforeCursor() string {
	parts := strings.Split(string(s), " ")
	if len(parts) < 1 {
		return ""
	}
	return parts[len(parts)-1]
}

func TestBasicCmdWrapper_strings(t *testing.T) {
	cmd := BasicCmdWrapper{BasicCmd{
		Name:  "cmd",
		Short: "A test command",
	}}

	assert := assert.New(t)
	assert.Equal("cmd", cmd.Name(), "invalid command name")
	assert.Equal("A test command", cmd.Short(), "invalid command short description")
	assert.Equal(
		"A test command\n\nUsage:\n  cmd\n\nFlags:\n  -h, --help   Invoke help on command",
		cmd.Long(),
		"invalid command long description")
}

func TestBasicCmdWrapper_strings_noFlags(t *testing.T) {
	cmd := BasicCmdWrapper{BasicCmd{
		Name:    "cmd",
		Short:   "A test command",
		NoFlags: true,
	}}

	assert.Equal(t, "A test command\n\nUsage:\n  cmd", cmd.Long(), "invalid command long description")
}

func TestBasicCmdWrapper_Suggest(t *testing.T) {
	cmd := BasicCmdWrapper{BasicCmd{
		Name:  "cmd",
		Short: "A test command",
		Flags: func() *pflag.FlagSet {
			flags := pflag.NewFlagSet("cmd", pflag.ContinueOnError)
			flags.String("flag", "", "")
			return flags
		},
		Exec: func(*BasicContext) error {
			return nil
		},
	}}

	tests := []struct {
		name   string
		text   string
		expect []string
	}{{
		"empty",
		"",
		[]string{"cmd"},
	}, {
		"partial",
		"cm",
		[]string{"cmd"},
	}, {
		"mismatch",
		"cdm",
		[]string{},
	}, {
		"empty flag",
		"cmd -",
		[]string{"--flag", "--help"},
	}, {
		"partial flag",
		"cmd --f",
		[]string{"--flag"},
	}, {
		"flag mismatch",
		"cmd --fgas",
		[]string{},
	}}

	for _, tt := range tests {
		content := basicContentMock(tt.text)
		suggs := cmd.Suggest(content)
		completions := make([]string, len(suggs))
		for i, s := range suggs {
			completions[i] = s.Text
		}

		got := fmt.Sprintf("%s", completions)
		want := fmt.Sprintf("%s", tt.expect)

		assert.Equalf(t, want, got, "%s: invalid completions", tt.name)
	}
}

func TestBasicCmdWrapper_Exec(t *testing.T) {
	execCh := make(chan struct{})

	cmd := BasicCmdWrapper{BasicCmd{
		Name:  "cmd",
		Short: "A test command",
		Exec: func(*BasicContext) error {
			close(execCh)
			return nil
		},
	}}

	closure := script.NewClosure(script.ClosureOptResolver(script.ResolveName))

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	_, err := cmd.Exec(&script.InterpreterContext{
		Ctx:     ctx,
		Closure: closure,
		EvalListToStrings: func(
			*script.InterpreterContext,
			script.SExp,
			bool,
		) ([]string, error) {
			return nil, nil
		},
	}, nil)

	assert.NoError(t, err)

	select {
	case <-execCh:
	case <-ctx.Done():
		assert.Fail(t, "command was not executed")
	}
}
