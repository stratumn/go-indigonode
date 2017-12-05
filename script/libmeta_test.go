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

var libMetaTests = []libTest{{
	"quote hello",
	"hello",
	"",
}, {
	"quote (hello)",
	"(hello)",
	"",
}, {
	"quote (hello world)",
	"(hello world)",
	"",
}, {
	"quote (quote (quote hello))",
	"(quote (quote hello))",
	"",
}, {
	"quote",
	"",
	"1:1: quote: expected a single expression",
}, {
	"quote hello world",
	"",
	"1:1: quote: expected a single expression",
}, {
	"eval 'hello world'",
	`"hello world"`,
	"",
}, {
	"eval (quote (quote 'hello world'))",
	`"hello world"`,
	"",
}, {
	"eval",
	"",
	"1:1: eval: expected a single expression",
}, {
	"eval a",
	"",
	"1:1: eval: 1:6: a: could not resolve symbol",
}, {
	"eval a b",
	"",
	"1:1: eval: expected a single expression",
}, {
	"eval (quote ('hello world'))",
	``,
	`1:1: eval: 1:14: function name is not a symbol`,
}}

func TestLibMeta(t *testing.T) {
	testLib(t, LibMeta, libMetaTests)
}
