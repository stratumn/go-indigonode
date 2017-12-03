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
	"context"
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/stratumn/alice/cli"
	"github.com/stratumn/alice/cli/script"
)

func TestLambda(t *testing.T) {
	tests := []ExecTest{{
		"(lambda (a b c) (echo $a $b $c))",
		"(lambda (a b c) (echo $a $b $c))",
		nil,
		nil,
	}, {
		"(lambda () ((echo $a) (echo $b) (echo $c)))",
		"(lambda <nil> ((echo $a) (echo $b) (echo $c)))",
		nil,
		nil,
	}, {
		"(lambda () 'a')",
		"(lambda <nil> \"a\")",
		nil,
		nil,
	}, {
		"(lambda)",
		"",
		ErrUse,
		nil,
	}, {
		"(lambda (a b c))",
		"",
		ErrUse,
		nil,
	}, {
		"(lambda ('a') ())",
		"",
		ErrUse,
		nil,
	}, {
		"(lambda 'a' ())",
		"",
		ErrUse,
		nil,
	}}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("%d-%s", i, tt.Command), func(t *testing.T) {
			tt.Test(t, cli.Lambda)
		})
	}
}

func TestLambda_exec(t *testing.T) {
	tests := []struct {
		command string
		args    string
		want    string
	}{{
		"(lambda () hello)",
		"()",
		"hello",
	}, {
		"(lambda () (echo hello world!))",
		"()",
		"hello world!",
	}, {
		"(lambda (a b c) (echo $a $b $c))",
		"(one two three)",
		"one two three",
	}, {
		"(lambda (a b c) ((echo $a) (echo $b) (echo $c)))",
		"((echo one) (echo two) (echo three))",
		"three",
	}}

	parser := script.NewParser(script.NewScanner())
	closure := script.NewClosure(script.OptResolver(cli.Resolver))

	createCall := func(context.Context, *script.Closure) script.CallHandler {
		return execCall
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("%d-%s", i, tt.command), func(t *testing.T) {
			et := ExecTest{Command: tt.command}

			fn, err := et.Exec(t, ioutil.Discard, cli.Lambda)
			if err != nil {
				t.Fatalf("tt.Exec(t, cli.Lambda, buf): error: %s", err)
			}

			args, err := parser.List(tt.args)
			if err != nil {
				t.Fatalf("%s: parser error: %s", tt.args, err)
			}

			exp, err := cli.ExecFunc(
				context.Background(),
				closure,
				createCall,
				"test",
				fn,
				args,
			)
			if err != nil {
				t.Fatalf("%s %s: exec error: %s", tt.command, tt.args, err)
			}

			if got := exp.MustStringVal(); got != tt.want {
				t.Errorf("%s %s: val = %v want %v", tt.command, tt.args, got, tt.want)
			}
		})
	}
}
