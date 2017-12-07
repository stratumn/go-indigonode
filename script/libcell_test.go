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

var libCellTests = []libTest{{
	"cons () ()",
	"(<nil>)",
	"",
}, {
	`cons "a" "b"`,
	`("a" . "b")`,
	"",
}, {
	`cons "a" ()`,
	`("a")`,
	"",
}, {
	`cons "a" (cons "b" "c")`,
	`("a" . ("b" . "c"))`,
	"",
}, {
	"cons",
	"",
	"1:1: cons: missing car",
}, {
	"cons ()",
	"",
	"1:1: cons: missing cdr",
}, {
	`cons ("hello") ()`,
	"",
	"1:1: cons: 1:7: function name is not a symbol",
}, {
	`cons () ("hello")`,
	"",
	"1:1: cons: 1:10: function name is not a symbol",
}, {
	"car ()",
	"",
	"",
}, {
	`car (cons "a" "b")`,
	`"a"`,
	"",
}, {
	`car (cons (cons "a" "b") "c")`,
	`("a" . "b")`,
	"",
}, {
	"car",
	"",
	"1:1: car: expected a single element",
}, {
	`car "a"`,
	"",
	`1:1: car: 1:5: "a": not a cell`,
}, {
	"car (a)",
	"",
	"1:1: car: 1:6: a: unknown function",
}, {
	"cdr ()",
	"",
	"",
}, {
	`cdr (cons "a" "b")`,
	`"b"`,
	"",
}, {
	`cdr (cons "a" (cons "b" "c"))`,
	`("b" . "c")`,
	"",
}, {
	"cdr",
	"",
	"1:1: cdr: expected a single element",
}, {
	`cdr "a"`,
	"",
	`1:1: cdr: 1:5: "a": not a cell`,
}, {
	"cdr (a)",
	"",
	"1:1: cdr: 1:6: a: unknown function",
}}

func TestLibCell(t *testing.T) {
	testLib(t, LibCell, libCellTests)
}
