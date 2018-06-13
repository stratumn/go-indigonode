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

package cli_test

import (
	"bytes"
	"context"
	"io/ioutil"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stratumn/alice/cli"
	"github.com/stratumn/alice/cli/mockcli"
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
