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
