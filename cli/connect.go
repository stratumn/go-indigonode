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

package cli

import (
	"strings"
)

// Connect is a command that creates a connection to the API server.
var Connect = BasicCmdWrapper{BasicCmd{
	Name:  "api-connect",
	Use:   "api-connect [Multiaddress]",
	Short: "Connect or reconnect to API server",
	Exec:  execConnect,
}}

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
