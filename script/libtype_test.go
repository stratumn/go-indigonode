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
