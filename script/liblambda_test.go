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

var libLambdaTests = []libTest{{
	"(lambda (a b c) (echo a b c))",
	"(lambda (a b c) (echo a b c))",
	"",
}, {
	"(lambda () ((echo a) (echo b) (echo c)))",
	"(lambda () ((echo a) (echo b) (echo c)))",
	"",
}, {
	`(lambda () "a")`,
	`(lambda () "a")`,
	"",
}, {
	"(lambda)",
	"",
	"1:2: lambda: missing function arguments",
}, {
	"(lambda (a b c))",
	"",
	"1:2: lambda: missing function body",
}, {
	`(lambda ("a") ())`,
	"",
	"1:2: lambda: 1:10: function argument is not a symbol",
}, {
	`(lambda "a" ())`,
	"",
	"1:2: lambda: 1:9: function arguments are not a list",
}}

func TestLibLambda(t *testing.T) {
	testLib(t, LibLambda, libLambdaTests)
}
