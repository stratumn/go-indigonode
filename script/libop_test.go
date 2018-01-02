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

var libOpTests = []libTest{{
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
	"1:1: +: 1:3: (): not an integer",
}, {
	"+ 1 ()",
	"",
	"1:1: +: 1:5: (): not an integer",
}, {
	"+ true",
	"",
	"1:1: +: 1:3: true: not an integer",
}, {
	"+ 1 true",
	"",
	"1:1: +: 1:5: true: not an integer",
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
	"1:1: /: 1:5: 0: division by zero",
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
	"1:1: mod: 1:7: 0: division by zero",
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
	"= () ()",
	"true",
	"",
}, {
	"= (quote (a b c)) (quote (a b c))",
	"false",
	"",
}, {
	"= 1 ()",
	"false",
	"",
}, {
	"= () 1",
	"false",
	"",
}, {
	`= 1 "one"`,
	"false",
	"",
}, {
	"= 1 true",
	"false",
	"",
}, {
	"= 1 (quote one)",
	"false",
	"",
}, {
	"= 1 (quote (one))",
	"false",
	"",
}, {
	"=",
	"",
	"1:1: =: missing argument",
}, {
	"= test",
	"",
	"1:1: =: 1:3: test: could not resolve symbol",
}, {
	"= 1 test",
	"",
	"1:1: =: 1:5: test: could not resolve symbol",
}, {
	"< 1",
	"true",
	"",
}, {
	"< 1 2 3",
	"true",
	"",
}, {
	"< 1 3 3",
	"false",
	"",
}, {
	"<",
	"",
	"1:1: <: missing argument",
}, {
	"< ()",
	"",
	"1:1: <: 1:3: (): not an integer",
}, {
	"< 1 ()",
	"",
	"1:1: <: 1:5: (): not an integer",
}, {
	"< true",
	"",
	"1:1: <: 1:3: true: not an integer",
}, {
	"< 1 true",
	"",
	"1:1: <: 1:5: true: not an integer",
}, {
	"< test",
	"",
	"1:1: <: 1:3: test: could not resolve symbol",
}, {
	"< 1 test",
	"",
	"1:1: <: 1:5: test: could not resolve symbol",
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
}, {
	"not true",
	"false",
	"",
}, {
	"not false",
	"true",
	"",
}, {
	"not",
	"",
	"1:1: not: expected a single argument",
}, {
	"not true true",
	"",
	"1:1: not: expected a single argument",
}, {
	"not ()",
	"",
	"1:1: not: 1:5: (): not a boolean",
}, {
	"not 1",
	"",
	"1:1: not: 1:5: 1: not a boolean",
}, {
	"not test",
	"",
	"1:1: not: 1:5: test: could not resolve symbol",
}, {
	"and true",
	"true",
	"",
}, {
	"and false",
	"false",
	"",
}, {
	"and true true true",
	"true",
	"",
}, {
	"and true true false",
	"false",
	"",
}, {
	"and",
	"",
	"1:1: and: missing argument",
}, {
	"and ()",
	"",
	"1:1: and: 1:5: (): not a boolean",
}, {
	"and true ()",
	"",
	"1:1: and: 1:10: (): not a boolean",
}, {
	"and 1",
	"",
	"1:1: and: 1:5: 1: not a boolean",
}, {
	"and true 1",
	"",
	"1:1: and: 1:10: 1: not a boolean",
}, {
	"and test",
	"",
	"1:1: and: 1:5: test: could not resolve symbol",
}, {
	"and true test",
	"",
	"1:1: and: 1:10: test: could not resolve symbol",
}, {
	"or true",
	"true",
	"",
}, {
	"or false",
	"false",
	"",
}, {
	"or false false true",
	"true",
	"",
}, {
	"or false false false",
	"false",
	"",
}, {
	"or",
	"",
	"1:1: or: missing argument",
}, {
	"or ()",
	"",
	"1:1: or: 1:4: (): not a boolean",
}, {
	"or false ()",
	"",
	"1:1: or: 1:10: (): not a boolean",
}, {
	"or 1",
	"",
	"1:1: or: 1:4: 1: not a boolean",
}, {
	"or false 1",
	"",
	"1:1: or: 1:10: 1: not a boolean",
}, {
	"or test",
	"",
	"1:1: or: 1:4: test: could not resolve symbol",
}, {
	"or false test",
	"",
	"1:1: or: 1:10: test: could not resolve symbol",
}}

func TestLibOp(t *testing.T) {
	testLib(t, LibOp, libOpTests)
}
