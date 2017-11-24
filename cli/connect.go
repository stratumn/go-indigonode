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
	"strings"

	"github.com/spf13/pflag"
)

// Connect is a command that creates a connection to the API server.
var Connect = BasicCmdWrapper{BasicCmd{
	Name:  "api-connect",
	Use:   "api-connect [Multiaddress]",
	Short: "Connect or reconnect to API server",
	Exec:  execConnect,
}}

func execConnect(ctx context.Context, cli CLI, args []string, flags *pflag.FlagSet) error {
	if len(args) > 1 {
		return NewUseError("unexpected argument(s): " + strings.Join(args[1:], " "))
	}

	var addr string

	if len(args) > 0 {
		addr = args[0]
	} else {
		addr = cli.Address()
	}

	c := cli.Console()

	c.Infof("Connecting to %q...\n", addr)

	if err := cli.Connect(ctx, addr); err != nil {
		c.Errorf("Could not connect to %q.\n", addr)
		return err
	}

	c.Successf("Connected to %q.\n", addr)

	return nil
}
