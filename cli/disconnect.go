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
	"io"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/pflag"
)

// Disconnect is a command that closes the connection to the API server.
var Disconnect = BasicCmdWrapper{BasicCmd{
	Name:        "api-disconnect",
	Short:       "Disconnect from API server",
	ExecStrings: disconnectExec,
}}

func disconnectExec(
	ctx context.Context,
	cli CLI,
	w io.Writer,
	args []string,
	flags *pflag.FlagSet,
) error {
	if len(args) > 0 {
		return NewUseError("unexpected argument(s): " + strings.Join(args, " "))
	}

	if err := cli.Disconnect(); err != nil {
		return errors.WithStack(err)
	}

	cli.Console().Print("Disconnected.")

	return nil
}
