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

func TestIf(t *testing.T) {
	tests := []ExecTest{{
		"if (title test) ok ko",
		"ok",
		nil,
		nil,
	}, {
		"if (test) ok ko",
		"ko",
		nil,
		nil,
	}, {
		"if (title test) (title ok) (title ko)",
		"Ok",
		nil,
		nil,
	}, {
		"if (test) (title ok) (title ko)",
		"Ko",
		nil,
		nil,
	}, {
		"if (title test) ok",
		"ok",
		nil,
		nil,
	}, {
		"if (test) ok",
		"",
		nil,
		nil,
	}, {
		"if test ok",
		"ok",
		nil,
		nil,
	}, {
		"if $test ok ko",
		"ko",
		nil,
		nil,
	}, {
		"if test ((title o) (title k)) ((title k) (title o))",
		"K",
		nil,
		nil,
	}, {
		"if $test ((title o) (title k)) ((title k) (title o))",
		"O",
		nil,
		nil,
	}, {
		"if test ((title o) k)",
		"k",
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
		"if test $ok",
		"",
		script.ErrSymNotFound,
		nil,
	}, {
		"if (test) ok $ok",
		"",
		script.ErrSymNotFound,
		nil,
	}, {
		"if (title test)",
		"",
		ErrUse,
		nil,
	}}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("%d-%s", i, tt.Command), func(t *testing.T) {
			tt.Test(t, cli.If)
		})
	}
}
