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
	"fmt"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/pflag"
	"github.com/stratumn/go-indigonode/release"
)

// Version is a command that displays the client version string.
var Version = BasicCmdWrapper{
	Cmd: BasicCmd{
		Name:  "cli-version",
		Short: "Output program version string",
		Flags: versionFlags,
		Exec:  versionExec,
	},
}

func versionFlags() *pflag.FlagSet {
	flags := pflag.NewFlagSet("cli-version", pflag.ContinueOnError)
	flags.Int("git-commit-length", 0, "Truncate Git commit hash to specified length")
	return flags
}

func versionExec(ctx *BasicContext) error {
	if len(ctx.Args) > 0 {
		return NewUseError("unexpected argument(s): " + strings.Join(ctx.Args, " "))
	}

	l, err := ctx.Flags.GetInt("git-commit-length")
	if err != nil {
		return errors.WithStack(err)
	}

	commit := release.GitCommit

	if l > 0 && l < len(commit) {
		commit = commit[:l]
	}

	fmt.Fprintln(ctx.Writer, release.Version+"@"+commit)

	return nil
}
