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
	"bytes"
	"context"
	"io/ioutil"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stratumn/go-node/cli"
	"github.com/stratumn/go-node/cli/mockcli"
	"github.com/stretchr/testify/assert"
)

func TestRline(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	c := mockcli.NewMockCLI(ctrl)
	buf := bytes.NewBuffer(nil)

	p := cli.NewRline(c)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	runCh := make(chan struct{})
	go func() {
		p.Run(ctx, ioutil.NopCloser(buf))
		close(runCh)
	}()

	c.EXPECT().Suggest(gomock.Any()).Return([]cli.Suggest{{
		Text: "help",
	}})
	c.EXPECT().Run(ctx, "help")
	buf.WriteString("he\t\n")

	select {
	case <-runCh:
	case <-time.After(time.Second):
		assert.Fail(t, "prompt did not stop")
	}
}
