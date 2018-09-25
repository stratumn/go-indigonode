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
	"github.com/stratumn/go-node/release"
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
