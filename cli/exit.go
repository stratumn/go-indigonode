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
	"os"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

// Exit is a command that terminates the current process.
var Exit = BasicCmdWrapper{
	Cmd: BasicCmd{
		Name:  "exit",
		Use:   "exit [Status]",
		Short: "Exit program",
		Exec:  exitExec,
	},
}

func exitExec(ctx *BasicContext) error {
	argc := len(ctx.Args)
	if argc > 1 {
		return NewUseError("unexpected argument(s): " + strings.Join(ctx.Args[1:], " "))
	}

	status := 0

	if argc > 0 {
		statusArg := ctx.Args[0]
		var err error
		if status, err = strconv.Atoi(statusArg); err != nil {
			return errors.WithStack(ErrInvalidExitCode)
		}
	}

	ctx.CLI.Console().Println("Goodbye!")
	os.Exit(status)

	return nil
}
