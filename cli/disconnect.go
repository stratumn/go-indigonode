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

	"github.com/pkg/errors"
)

// Disconnect is a command that closes the connection to the API server.
var Disconnect = BasicCmdWrapper{
	Cmd: BasicCmd{
		Name:  "api-disconnect",
		Short: "Disconnect from API server",
		Exec:  disconnectExec,
	},
}

func disconnectExec(ctx *BasicContext) error {
	if len(ctx.Args) > 0 {
		return NewUseError("unexpected argument(s): " + strings.Join(ctx.Args, " "))
	}

	if err := ctx.CLI.Disconnect(); err != nil {
		return errors.WithStack(err)
	}

	ctx.CLI.Console().Println("Disconnected.")

	return nil
}
