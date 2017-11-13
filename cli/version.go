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
	"github.com/stratumn/alice/release"
)

// Version is a command that displays the client version string.
var Version = BasicCmdWrapper{BasicCmd{
	Name:  "cli-version",
	Short: "Display command line interface version string",
	Exec:  versionExec,
}}

func versionExec(ctx context.Context, cli *CLI, args []string, flags *pflag.FlagSet) error {
	if len(args) > 0 {
		return NewUseError("unexpected argument(s): " + strings.Join(args, " "))
	}

	cli.Console().Println(release.Version + "@" + release.GitCommit)

	return nil
}