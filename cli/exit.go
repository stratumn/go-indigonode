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
	"os"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/pflag"
)

// Exit is a command that terminates the current process.
var Exit = BasicCmdWrapper{BasicCmd{
	Name:  "exit",
	Use:   "exit [Status]",
	Short: "Exit command line interface",
	Exec:  exitExec,
}}

func exitExec(ctx context.Context, cli CLI, w io.Writer, args []string, flags *pflag.FlagSet) error {
	if len(args) > 1 {
		return NewUseError("unexpected argument(s): " + strings.Join(args[1:], " "))
	}

	status := 0

	if len(args) > 0 {
		statusArg := args[0]
		var err error
		if status, err = strconv.Atoi(statusArg); err != nil {
			return errors.WithStack(ErrInvalidExitCode)
		}
	}

	cli.Console().Println("Goodbye!")
	os.Exit(status)

	return nil
}
