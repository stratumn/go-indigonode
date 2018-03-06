// Copyright © 2017-2018 Stratumn SAS
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
	`eval "hello world"`,
	`"hello world"`,
	"",
}, {
	`eval (quote (quote "hello world"))`,
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
	`eval (quote ("hello world"))`,
	``,
	`1:1: eval: 1:14: function name is not a symbol`,
}, {
	"quasiquote (hello world)",
	"(hello world)",
	"",
}, {
	"quasiquote (quote (quasiquote hello))",
	"(quote (quasiquote hello))",
	"",
}, {
	"quasiquote (unquote (+ 1 2))",
	"3",
	"",
}, {
	"quasiquote ((unquote (+ 1 2)))",
	"(3)",
	"",
}, {
	"quasiquote (a b (unquote (+ 1 2)) c)",
	"(a b 3 c)",
	"",
}, {
	"quasiquote (quasiquote (unquote (+ 1 2)))",
	"(quasiquote (unquote (+ 1 2)))",
	"",
}, {
	"quasiquote (quasiquote (unquote (unquote (+ 1 2))))",
	"(quasiquote (unquote 3))",
	"",
}, {
	"quasiquote",
	"",
	"1:1: quasiquote: expected a single expression",
}, {
	"quasiquote hello world",
	"",
	"1:1: quasiquote: expected a single expression",
}, {
	"unquote",
	"",
	"1:1: unquote: expected a single expression",
}, {
	"(quasiquote (unquote (unquote a)))",
	"",
	"1:2: quasiquote: 1:14: unquote: 1:23: unquote: outside of quasiquote",
}}

func TestLibMeta(t *testing.T) {
	testLib(t, LibMeta, libMetaTests)
}
