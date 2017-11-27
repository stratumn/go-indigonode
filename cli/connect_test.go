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

package cli_test

import (
	"fmt"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/pkg/errors"
	"github.com/stratumn/alice/cli"
	"github.com/stratumn/alice/cli/mockcli"
)

func TestConnect(t *testing.T) {
	addr := "/ip4/127.0.0.1/tcp/8904"
	err := errors.New("could not connect")

	tt := []ExecTest{{
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

	for i, test := range tt {
		t.Run(fmt.Sprintf("%d-%s", i, test.Command), func(t *testing.T) {
			test.Test(t, cli.Connect.Cmd)
		})
	}
}
