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

var libIntTests = []libTest{{
	"+ 1",
	"1",
	"",
}, {
	"+ 1 2 3",
	"6",
	"",
}, {
	"+",
	"",
	"1:1: +: missing argument",
}, {
	"+ ()",
	"",
	"1:1: +: 1:3: not an integer",
}, {
	"+ 1 ()",
	"",
	"1:1: +: 1:5: not an integer",
}, {
	"+ true",
	"",
	"1:1: +: 1:3: not an integer",
}, {
	"+ 1 true",
	"",
	"1:1: +: 1:5: not an integer",
}, {
	"+ test",
	"",
	"1:1: +: 1:3: test: could not resolve symbol",
}, {
	"+ 1 test",
	"",
	"1:1: +: 1:5: test: could not resolve symbol",
}, {
	"- 1 2 3",
	"-4",
	"",
}, {
	"* 2 3 4",
	"24",
	"",
}, {
	"/ 20 2 5",
	"2",
	"",
}, {
	"/ 1 0",
	"",
	"1:1: /: 1:5: division by zero",
}, {
	"mod 113 20 10",
	"3",
	"",
}, {
	"mod",
	"",
	"1:1: mod: missing argument",
}, {
	"mod 1 0",
	"",
	"1:1: mod: 1:7: division by zero",
}, {
	"= 1",
	"true",
	"",
}, {
	"= 1 1 1",
	"true",
	"",
}, {
	"= 1 1 2",
	"false",
	"",
}, {
	"=",
	"",
	"1:1: =: missing argument",
}, {
	"= ()",
	"",
	"1:1: =: 1:3: not an integer",
}, {
	"= 1 ()",
	"",
	"1:1: =: 1:5: not an integer",
}, {
	"= true",
	"",
	"1:1: =: 1:3: not an integer",
}, {
	"= 1 true",
	"",
	"1:1: =: 1:5: not an integer",
}, {
	"= test",
	"",
	"1:1: =: 1:3: test: could not resolve symbol",
}, {
	"= 1 test",
	"",
	"1:1: =: 1:5: test: could not resolve symbol",
}, {
	"< 1 2 3",
	"true",
	"",
}, {
	"< 1 3 3",
	"false",
	"",
}, {
	"> 3 2 1",
	"true",
	"",
}, {
	"> 3 3 2",
	"false",
	"",
}, {
	"<= 1 2 2",
	"true",
	"",
}, {
	"<= 1 3 2",
	"false",
	"",
}, {
	">= 3 2 1",
	"true",
	"",
}, {
	">= 3 2 3",
	"false",
	"",
}}

func TestLibInt(t *testing.T) {
	testLib(t, LibInt, libIntTests)
}
