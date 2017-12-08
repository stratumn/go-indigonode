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

package script

import (
	"testing"
)

var libLambdaTests = []libTest{{
	"(lambda (a b c) (echo a b c))",
	"(lambda (a b c) (echo a b c))",
	"",
}, {
	"(lambda () ((echo a) (echo b) (echo c)))",
	"(lambda () ((echo a) (echo b) (echo c)))",
	"",
}, {
	`(lambda () "a")`,
	`(lambda () "a")`,
	"",
}, {
	"(lambda)",
	"",
	"1:2: lambda: missing function arguments",
}, {
	"(lambda (a b c))",
	"",
	"1:2: lambda: missing function body",
}, {
	`(lambda ("a") ())`,
	"",
	"1:2: lambda: 1:10: function argument is not a symbol",
}, {
	`(lambda "a" ())`,
	"",
	"1:2: lambda: 1:9: function arguments are not a list",
}}

func TestLibLambda(t *testing.T) {
	testLib(t, LibLambda, libLambdaTests)
}
