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

package cli_test

import (
	"fmt"
	"testing"

	"github.com/stratumn/go-indigonode/cli"
)

func TestBang(t *testing.T) {
	tests := []ExecTest{{
		"! echo",
		"\n",
		nil,
		nil,
	}, {
		`! echo "hello outside world!"`,
		"hello outside world!\n",
		nil,
		nil,
	}, {
		`! cat "" "hello outside world!"`,
		"hello outside world!",
		nil,
		nil,
	}, {
		`! cat () "hello outside world!"`,
		"hello outside world!",
		nil,
		nil,
	}, {
		`! grep outside "hello\noutside\nworld!"`,
		"outside\n",
		nil,
		nil,
	}, {
		`! grep "-v outside" "hello\noutside\nworld!"`,
		"hello\nworld!\n",
		nil,
		nil,
	}, {
		"! (a)",
		"",
		ErrAny,
		nil,
	}, {
		"!",
		"",
		ErrUse,
		nil,
	}, {
		"! echo a b c",
		"",
		ErrUse,
		nil,
	}, {
		"! 1234567890",
		"",
		ErrAny,
		nil,
	}}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("%d-%s", i, tt.Command), func(t *testing.T) {
			tt.Test(t, cli.Bang)
		})
	}
}
