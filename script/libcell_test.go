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

var libCellTests = []libTest{{
	"cons () ()",
	"(())",
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
	`cons ("hello") ()`,
	"",
	"1:1: cons: 1:7: function name is not a symbol",
}, {
	`cons () ("hello")`,
	"",
	"1:1: cons: 1:10: function name is not a symbol",
}, {
	"car ()",
	"()",
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
	"()",
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
