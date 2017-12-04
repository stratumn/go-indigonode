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
	"if 'test' 'ok' 'ko'",
	`"ok"`,
	"",
}, {
	"if () 'ok' 'ko'",
	`"ko"`,
	"",
}, {
	"if 'test' 'ok'",
	`"ok"`,
	"",
}, {
	"if () 'ok'",
	"",
	"",
}, {
	"if 'test' () 'ko'",
	"",
	"",
}, {
	"if () () 'ko'",
	`"ko"`,
	"",
}, {
	"if",
	"",
	"1:1: if: missing condition expression",
}, {
	"if 'test'",
	"",
	"1:1: if: missing then expression",
}, {
	"if 'test' ok",
	"",
	"1:1: if: ok: could not resolve symbol",
}, {
	"if () 'ok' ko",
	"",
	"1:1: if: ko: could not resolve symbol",
}, {
	"if 'test' 'ok' 'ko' 'uncertain'",
	"",
	"1:1: if: unexpected expression",
}, {
	"if test 'ok'",
	"",
	"1:1: if: test: could not resolve symbol",
}, {
	"unless 'test' 'ok' 'ko'",
	`"ko"`,
	"",
}, {
	"unless () 'ok' 'ko'",
	`"ok"`,
	"",
}, {
	"unless 'test' 'ok'",
	"",
	"",
}, {
	"unless () 'ok'",
	`"ok"`,
	"",
}, {
	"unless 'test' () 'ko'",
	`"ko"`,
	"",
}, {
	"unless () () 'ko'",
	"",
	"",
}, {
	"unless",
	"",
	"1:1: unless: missing condition expression",
}, {
	"unless 'test'",
	"",
	"1:1: unless: missing then expression",
}, {
	"unless 'test' ok",
	"",
	"1:1: unless: ok: could not resolve symbol",
}, {
	"unless () ok 'ko'",
	"",
	"1:1: unless: ok: could not resolve symbol",
}, {
	"unless 'test' 'ok' 'ko' 'uncertain'",
	"",
	"1:1: unless: unexpected expression",
}, {
	"unless test 'ok'",
	"",
	"1:1: unless: test: could not resolve symbol",
}}

func TestLibCond(t *testing.T) {
	testLib(t, LibCond, libCondTests)
}
