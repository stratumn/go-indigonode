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

package cli

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/pflag"
)

// Echo is a command that outputs text.
var Echo = BasicCmdWrapper{
	Cmd: BasicCmd{
		Name:  "echo",
		Short: "Output text",
		Use:   "echo [Expressions...]",
		Flags: echoFlags,
		Exec:  echoExec,
	},
}

func echoFlags() *pflag.FlagSet {
	flags := pflag.NewFlagSet("echo", pflag.ContinueOnError)
	flags.StringP("log", "s", "", "Print log message (debug, info, normal, success, warning, error)")

	return flags
}

func echoExec(ctx *BasicContext) error {
	s := strings.Join(ctx.Args, " ")

	log, err := ctx.Flags.GetString("log")
	if err != nil {
		return errors.WithStack(err)
	}

	cons := ctx.CLI.Console()

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
		fmt.Fprintln(ctx.Writer, s)
	}

	return nil
}
