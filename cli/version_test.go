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

package cli_test

import (
	"fmt"
	"testing"

	"github.com/stratumn/go-indigonode/cli"
	"github.com/stratumn/go-indigonode/release"
)

func TestVersion(t *testing.T) {
	tests := []ExecTest{{
		"cli-version",
		release.Version + "@" + release.GitCommit + "\n",
		nil,
		nil,
	}, {
		"cli-version",
		release.Version + "@" + release.GitCommit + "\n",
		nil,
		nil,
	}, {
		"cli-version --git-commit-length 100",
		release.Version + "@" + release.GitCommit + "\n",
		nil,
		nil,
	}, {
		"cli-version earth",
		"",
		ErrUse,
		nil,
	}}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("%d-%s", i, tt.Command), func(t *testing.T) {
			tt.Test(t, cli.Version)
		})
	}
}
