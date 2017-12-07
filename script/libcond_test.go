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
	"",
	"",
}, {
	`if false "ko"`,
	"",
	"",
}, {
	`if false () "ko"`,
	`"ko"`,
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
	`unless true "ok" "ko"`,
	`"ko"`,
	"",
}, {
	`unless false "ok" "ko"`,
	`"ok"`,
	"",
}, {
	`unless true "ok"`,
	"",
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
	"",
	"",
}}

func TestLibCond(t *testing.T) {
	testLib(t, LibCond, libCondTests)
}
