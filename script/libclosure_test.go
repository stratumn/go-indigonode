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

var libClosureTests = []libTest{{
	"let a 'hello'",
	`"hello"`,
	"",
}, {
	"let a",
	"",
	"",
}, {
	"let",
	"",
	"1:1: let: missing symbol",
}, {
	"let 'a'",
	"",
	"1:1: let: 1:5: not a symbol",
}, {
	"let a 'hello'\nlet a 'world'",
	`"hello"`,
	"2:1: let: a value is already bound to the symbol",
}}

func TestLibClosure(t *testing.T) {
	testLib(t, LibClosure, libClosureTests)
}
