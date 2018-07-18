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
	"strings"
)

// Connect is a command that creates a connection to the API server.
var Connect = BasicCmdWrapper{
	Cmd: BasicCmd{
		Name:  "api-connect",
		Use:   "api-connect [Multiaddress]",
		Short: "Connect or reconnect to API server",
		Exec:  execConnect,
	},
}

func execConnect(ctx *BasicContext) error {
	argc := len(ctx.Args)
	if argc > 1 {
		return NewUseError("unexpected argument(s): " + strings.Join(ctx.Args[1:], " "))
	}

	var addr string

	if argc > 0 {
		addr = ctx.Args[0]
	} else {
		addr = ctx.CLI.Address()
	}

	c := ctx.CLI.Console()

	c.Infof("Connecting to %q...\n", addr)

	if err := ctx.CLI.Connect(ctx.Ctx, addr); err != nil {
		c.Errorf("Could not connect to %q.\n", addr)
		return err
	}

	c.Successf("Connected to %q.\n", addr)

	return nil
}
