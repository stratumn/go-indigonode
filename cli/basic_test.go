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
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/spf13/pflag"
	"github.com/stratumn/alice/script"
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

	if got, want := cmd.Name(), "cmd"; got != want {
		t.Errorf("cmd.Name() = %v want %v", got, want)
	}

	if got, want := cmd.Short(), "A test command"; got != want {
		t.Errorf("cmd.Short() = %v want %v", got, want)
	}

	if got, want := cmd.Long(), "A test command\n\nUsage:\n  cmd\n\nFlags:\n  -h, --help   Invoke help on command"; got != want {
		t.Errorf("cmd.Long() = \n%v want\n%v", got, want)
	}
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

		if got != want {
			t.Errorf("%s: completions = %v want %v", tt.name, got, want)
		}
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

	if err != nil {
		t.Errorf(`cmd.Exec(): error: %s`, err)
	}

	select {
	case <-execCh:
	case <-ctx.Done():
		t.Errorf("command was not executed")
	}
}
