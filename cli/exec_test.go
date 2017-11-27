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
	"bytes"
	"context"
	"strings"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/pkg/errors"
	"github.com/spf13/pflag"
	"github.com/stratumn/alice/cli"
	"github.com/stratumn/alice/cli/mockcli"
)

var (
	ErrAny = errors.New("any error")
	ErrUse = errors.New("usage error")
)

type ExecTest struct {
	Command string
	Want    string
	Err     error
	Expect  func(*mockcli.MockCLI)
}

func (e ExecTest) Test(t *testing.T, cmd cli.BasicCmd) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	c := mockcli.NewMockCLI(ctrl)
	buf := bytes.NewBuffer(nil)
	cons := cli.NewConsole(buf, false)

	c.EXPECT().Console().Return(cons).AnyTimes()
	if e.Expect != nil {
		e.Expect(c)
	}

	argv := strings.Split(e.Command, " ")[1:]

	var flags *pflag.FlagSet

	if cmd.Flags != nil {
		flags = cmd.Flags()
		if err := flags.Parse(argv); err != nil {
			t.Fatalf("%s: flag error: %s", e.Command, err)
		}
	} else {
		flags = pflag.NewFlagSet(e.Command, pflag.ContinueOnError)
	}

	err := errors.Cause(cmd.Exec(ctx, c, buf, argv, flags))

	switch {
	case e.Err == ErrAny && err != nil:
		// Pass.
	case e.Err == ErrUse:
		if _, ok := err.(*cli.UseError); !ok {
			t.Errorf("%s: error = %v want %v", e.Command, err, e.Err)
		}
	case err != e.Err:
		t.Errorf("%s: error = %v want %v", e.Command, err, e.Err)
	}

	got := buf.String()

	if got != e.Want {
		t.Errorf("%s =>\n%s\nwant\n\n%s", e.Command, got, e.Want)
	}
}
