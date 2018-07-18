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

var libTypeTests = []libTest{{
	"nil? ()",
	"true",
	"",
}, {
	"nil? 10",
	"false",
	"",
}, {
	"nil?",
	"",
	"1:1: nil?: expected a single expression",
}, {
	"nil? hello world",
	"",
	"1:1: nil?: expected a single expression",
}, {
	"nil? a",
	"",
	"1:1: nil?: 1:6: a: could not resolve symbol",
}, {
	"atom? ()",
	"true",
	"",
}, {
	"atom? 10",
	"true",
	"",
}, {
	"atom? (quote (a b))",
	"false",
	"",
}, {
	"atom?",
	"",
	"1:1: atom?: expected a single expression",
}, {
	"atom? hello world",
	"",
	"1:1: atom?: expected a single expression",
}, {
	"atom? a",
	"",
	"1:1: atom?: 1:7: a: could not resolve symbol",
}, {
	"list? ()",
	"true",
	"",
}, {
	"list? 10",
	"false",
	"",
}, {
	"list? (quote (a b))",
	"true",
	"",
}, {
	"list?",
	"",
	"1:1: list?: expected a single expression",
}, {
	"list? hello world",
	"",
	"1:1: list?: expected a single expression",
}, {
	"list? a",
	"",
	"1:1: list?: 1:7: a: could not resolve symbol",
}, {
	"sym? (quote a)",
	"true",
	"",
}, {
	"sym? ()",
	"false",
	"",
}, {
	"sym?",
	"",
	"1:1: sym?: expected a single expression",
}, {
	"sym? hello world",
	"",
	"1:1: sym?: expected a single expression",
}, {
	"sym? a",
	"",
	"1:1: sym?: 1:6: a: could not resolve symbol",
}, {
	`string? "hello"`,
	"true",
	"",
}, {
	"string? ()",
	"false",
	"",
}, {
	"int64? 0xFF",
	"true",
	"",
}, {
	"int64? ()",
	"false",
	"",
}, {
	"bool? true",
	"true",
	"",
}, {
	"bool? ()",
	"false",
	"",
}}

func TestLibType(t *testing.T) {
	testLib(t, LibType, libTypeTests)
}
