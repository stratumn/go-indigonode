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

	"github.com/golang/mock/gomock"
	"github.com/pkg/errors"
	"github.com/stratumn/go-indigonode/cli"
	"github.com/stratumn/go-indigonode/cli/mockcli"
)

func TestConnect(t *testing.T) {
	addr := "/ip4/127.0.0.1/tcp/8904"
	err := errors.New("could not connect")

	tests := []ExecTest{{
		"api-connect",
		`Connecting to "` + addr + `"...
Connected to "` + addr + `".
`,
		nil,
		func(c *mockcli.MockCLI) {
			c.EXPECT().Address().Return(addr).AnyTimes()
			c.EXPECT().Connect(gomock.Any(), addr).Return(nil)
		},
	}, {
		"api-connect " + addr,
		`Connecting to "` + addr + `"...
Connected to "` + addr + `".
`,
		nil,
		func(c *mockcli.MockCLI) {
			c.EXPECT().Address().Return("/ip4/127.0.0.1/tcp/8905").AnyTimes()
			c.EXPECT().Connect(gomock.Any(), addr).Return(nil)
		},
	}, {
		"api-connect earth moon",
		"",
		ErrUse,
		nil,
	}, {
		"api-connect",
		`Connecting to "` + addr + `"...
Could not connect to "` + addr + `".
`,
		err,
		func(c *mockcli.MockCLI) {
			c.EXPECT().Address().Return(addr).AnyTimes()
			c.EXPECT().Connect(gomock.Any(), addr).Return(err)
		},
	}}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("%d-%s", i, tt.Command), func(t *testing.T) {
			tt.Test(t, cli.Connect)
		})
	}
}
