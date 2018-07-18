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

var libCondTests = []libTest{{
	`if true "ok" "ko"`,
	`"ok"`,
	"",
}, {
	`if false "ok" "ko"`,
	`"ko"`,
	"",
}, {
	`if true "ok"`,
	`"ok"`,
	"",
}, {
	`if false "ok"`,
	"()",
	"",
}, {
	`if false "ko"`,
	"()",
	"",
}, {
	`if false () "ko"`,
	`"ko"`,
	"",
}, {
	`if true 1 else 2`,
	`1`,
	"",
}, {
	`if false 1 else 2`,
	`2`,
	"",
}, {
	"if",
	"",
	"1:1: if: missing condition expression",
}, {
	"if true",
	"",
	"1:1: if: missing then expression",
}, {
	"if true ok",
	"",
	"1:1: if: 1:9: ok: could not resolve symbol",
}, {
	`if false "ok" ko`,
	"",
	"1:1: if: 1:15: ko: could not resolve symbol",
}, {
	`if true "ok" "ko" "uncertain"`,
	"",
	"1:1: if: 1:19: unexpected expression",
}, {
	`if 1 "ok"`,
	"",
	"1:1: if: 1:4: 1: not a boolean",
}, {
	`if test "ok"`,
	"",
	"1:1: if: 1:4: test: could not resolve symbol",
}, {
	`if true 1 else`,
	"",
	"1:1: if: missing else expression",
}, {
	`if true 1 else 2 else 3`,
	"",
	"1:1: if: 1:18: unexpected expression",
}, {
	`unless true "ok" "ko"`,
	`"ko"`,
	"",
}, {
	`unless false "ok" "ko"`,
	`"ok"`,
	"",
}, {
	`unless true "ok"`,
	"()",
	"",
}, {
	`unless false "ok"`,
	`"ok"`,
	"",
}, {
	`unless true false "ko"`,
	`"ko"`,
	"",
}, {
	`unless false () "ko"`,
	"()",
	"",
}}

func TestLibCond(t *testing.T) {
	testLib(t, LibCond, libCondTests)
}
