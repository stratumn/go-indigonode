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

var libClosureTests = []libTest{{
	`let a "hello"`,
	`"hello"`,
	"",
}, {
	"let a",
	"()",
	"",
}, {
	"let",
	"",
	"1:1: let: missing symbol",
}, {
	`let "a"`,
	"",
	"1:1: let: 1:5: not a symbol",
}, {
	`let a "hello"
	let a "world"`,
	`"hello"`,
	"2:2: let: 2:6: a: a value is already bound to the symbol",
}}

func TestLibClosure(t *testing.T) {
	testLib(t, LibClosure, libClosureTests)
}
