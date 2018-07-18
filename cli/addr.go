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
)

// Addr is a command that displays the API server's address.
var Addr = BasicCmdWrapper{
	Cmd: BasicCmd{
		Name:  "api-address",
		Short: "Output API server address",
		Exec:  addrExec,
	},
}

func addrExec(ctx *BasicContext) error {
	if len(ctx.Args) > 0 {
		return NewUseError("unexpected argument(s): " + strings.Join(ctx.Args, " "))
	}

	fmt.Fprintln(ctx.Writer, ctx.CLI.Address())

	return nil
}
