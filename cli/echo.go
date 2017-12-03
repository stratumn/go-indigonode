// Copyright © 2017 Stratumn SAS
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

package cli

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/pflag"
)

// Echo is a command that outputs text.
var Echo = BasicCmdWrapper{BasicCmd{
	Name:        "echo",
	Short:       "Output text",
	Use:         "echo [Expressions...]",
	Flags:       echoFlags,
	ExecStrings: echoExec,
}}

func echoFlags() *pflag.FlagSet {
	flags := pflag.NewFlagSet("echo", pflag.ContinueOnError)
	flags.StringP("log", "s", "", "Print log message (debug, info, normal, success, warning, error)")

	return flags
}

func echoExec(
	ctx context.Context,
	cli CLI,
	w io.Writer,
	args []string,
	flags *pflag.FlagSet,
) error {
	s := strings.Join(args, " ")

	log, err := flags.GetString("log")
	if err != nil {
		return errors.WithStack(err)
	}

	cons := cli.Console()

	switch log {
	case "debug":
		cons.Debugln(s)
	case "normal":
		cons.Println(s)
	case "info":
		cons.Infoln(s)
	case "success":
		cons.Successln(s)
	case "warning":
		cons.Warningln(s)
	case "error":
		cons.Errorln(s)
	default:
		fmt.Fprintln(w, s)
	}

	return nil
}
